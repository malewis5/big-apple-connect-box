// Command big-apple-connect-box collects evidence about a home internet
// connection and saves it to evidence.db. The same program runs on a
// dedicated box and as an app on a volunteer's own computer.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/malewis5/big-apple-connect-box/apps/box/internal/collector"
	"github.com/malewis5/big-apple-connect-box/apps/box/internal/store"
)

const (
	databasePath  = "evidence.db"
	lightInterval = 5 * time.Minute // Netrics' schedule for light measurements
	scansPerFull  = 12              // one full scan an hour
	maxTargetLen  = 44              // width of the report's target column

	// monitorSaveEvery is how often the monitor's pings are written to the
	// database. Writing in batches spares a Raspberry Pi's memory card; a
	// power cut loses at most this much.
	monitorSaveEvery = time.Minute
)

func main() {
	log.SetFlags(0)
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	kind := flag.String("kind", collector.KindFull, "scan to run once: full or light")
	daemon := flag.Bool("daemon", false, "keep running: a ping every 10 seconds, a light scan every 5 minutes, a full scan every hour")
	version := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *version {
		fmt.Println(collector.Version())
		return nil
	}
	if *kind != collector.KindFull && *kind != collector.KindLight {
		return fmt.Errorf("unknown kind %q: use full or light", *kind)
	}

	db, err := store.Open(databasePath)
	if err != nil {
		return err
	}
	defer db.Close()

	if *daemon {
		return runDaemon(db)
	}
	scan, err := runScan(db, *kind)
	report(os.Stdout, scan)
	return err
}

// runDaemon runs the scan schedule and the ping monitor until the program is
// told to stop (Ctrl+C, or the system stopping it), then records that the
// monitor stopped. A scan still in progress at that moment is not saved.
func runDaemon(db *store.Store) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	runID, err := db.StartMonitor(time.Now(), collector.MonitorInterval, collector.PingCommand())
	if err != nil {
		return err
	}
	fmt.Printf("monitor run %d: pinging every %s\n", runID, collector.MonitorInterval)

	t := &targets{}
	go runSchedule(db, t)
	ping := func() []collector.Ping { return collector.PingAll(t.get()) }
	save := func(pings []collector.Ping) error { return db.SavePings(runID, pings) }
	if err := runMonitor(ctx, ping, save); err != nil {
		// Some pings were lost, so the run is left without a clean stop.
		return fmt.Errorf("monitor run %d: save the last pings: %w", runID, err)
	}
	return db.StopMonitor(runID, time.Now())
}

// runSchedule scans until the process is stopped: a full scan first and
// then every hour, a light scan every five minutes between. Each scan
// updates the addresses the monitor pings.
func runSchedule(db *store.Store, t *targets) {
	for n := 0; ; n++ {
		kind := collector.KindLight
		if n%scansPerFull == 0 {
			kind = collector.KindFull
		}
		started := time.Now()
		scan, err := runScan(db, kind)
		if err != nil {
			log.Print(err)
		}
		t.update(scan)
		time.Sleep(time.Until(started.Add(lightInterval)))
	}
}

// runMonitor pings every target each interval until ctx is done, saving the
// pings in batches and once more before it returns. It returns an error only
// when that last save fails, meaning pings were lost.
func runMonitor(ctx context.Context, ping func() []collector.Ping, save func([]collector.Ping) error) error {
	ticker := time.NewTicker(collector.MonitorInterval)
	defer ticker.Stop()

	var batch []collector.Ping
	saved := time.Now()
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		// On failure the batch is kept and tried again with the next one.
		if err := save(batch); err != nil {
			return err
		}
		batch, saved = nil, time.Now()
		return nil
	}

	for {
		batch = append(batch, ping()...)
		if time.Since(saved) >= monitorSaveEvery {
			if err := flush(); err != nil {
				log.Print(err)
			}
		}
		select {
		case <-ctx.Done():
			return flush()
		case <-ticker.C:
		}
	}
}

// targets holds the router and last-mile addresses for the monitor to ping:
// the latest each scan found. An address a scan didn't find, such as during
// an outage, keeps its earlier value.
type targets struct {
	mu                sync.Mutex
	gateway, lastMile string
}

func (t *targets) update(scan *collector.Scan) {
	gateway, lastMile := collector.TargetsFrom(scan)
	t.mu.Lock()
	defer t.mu.Unlock()
	if gateway != "" {
		t.gateway = gateway
	}
	if lastMile != "" {
		t.lastMile = lastMile
	}
}

func (t *targets) get() (gateway, lastMile string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.gateway, t.lastMile
}

// runScan performs one scan and saves it.
func runScan(db *store.Store, kind string) (*collector.Scan, error) {
	scan := collector.Run(kind)
	id, err := db.Save(scan)
	if err != nil {
		return scan, fmt.Errorf("save scan: %w", err)
	}
	fmt.Printf("scan %d (%s, %s) saved: %d probes\n", id, scan.Kind, scan.ConnectionType, len(scan.Probes))
	return scan, nil
}

// report prints a scan's observations, then the probes that found none.
func report(w io.Writer, scan *collector.Scan) {
	table := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(table, "\nSUBJECT\tATTRIBUTE\tVALUE\tTARGET")
	for _, p := range scan.Probes {
		for _, o := range p.Observations {
			fmt.Fprintf(table, "%s\t%s\t%s\t%s\n", o.Subject, o.Attribute, o.Value, describeTarget(p))
		}
	}
	fmt.Fprintln(table, "\nMETHOD\tTARGET\tERROR\t(probes that returned nothing readable)")
	for _, p := range scan.Probes {
		if len(p.Observations) == 0 {
			fmt.Fprintf(table, "%s\t%s\t%s\t\n", p.Method, p.Target, p.Err)
		}
	}
	table.Flush()
}

// describeTarget shortens a probe's target for the report and notes the
// probe it ran during.
func describeTarget(p *collector.Probe) string {
	target := p.Target
	if len(target) > maxTargetLen {
		target = target[:maxTargetLen-3] + "..."
	}
	if p.ConcurrentWith != nil {
		target += "  [during " + p.ConcurrentWith.Target + "]"
	}
	return target
}
