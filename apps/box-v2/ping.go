package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// The kernel's id for this start-up. It changes every time the box starts.
	bootIDPath = "/proc/sys/kernel/random/boot_id"
	// The kernel's count of seconds since the box started, then its idle time.
	uptimePath = "/proc/uptime"
	// systemd's time service creates this file when it first sets the clock
	// from a time server after the box starts.
	synchronizedPath = "/run/systemd/timesync/synchronized"
)

// internetTargets are the internet addresses a round pings: Cloudflare's and
// Google's. There are two, run by different operators, so that one of them
// being down doesn't look like the internet being down. They are addresses
// and not names, so that no name lookup is involved.
var internetTargets = []string{"1.1.1.1", "8.8.8.8"}

// A Ping is one ping and what came back, with the state of the box when the
// ping's round began. The pings of a round share that state exactly, which
// is what marks them as one round.
type Ping struct {
	At time.Time // the box's clock
	// ClockSynchronized is whether the time service had set that clock from
	// a time server since the box started. Until it has, At can't be relied
	// on; BootID and Uptime still place the ping.
	ClockSynchronized bool
	BootID            string
	Uptime            float64  // seconds since the box started
	Version           string   // the build of this program
	Target            string   // the address pinged
	Received          int      // replies ping counted; 0 too when it printed no count
	RTT               *float64 // the reply's round-trip time in ms
	// Output is what ping printed, or why it couldn't be run, kept when
	// there was no reply time to read.
	Output string
}

// runPings runs one round of pings on h and saves it to the database at
// path. A ping with no reply is what a round is there to record, so it is
// not an error.
func runPings(path string, h host) error {
	db, err := openDB(path)
	if err != nil {
		return err
	}
	defer db.Close()

	pings, err := pingRound(h)
	if err != nil {
		return err
	}
	// Printed before saving, so a round that can't be saved is at least in
	// the log.
	var results []string
	for _, p := range pings {
		result := "no reply"
		if p.RTT != nil {
			result = strconv.FormatFloat(*p.RTT, 'f', -1, 64) + " ms"
		}
		results = append(results, p.Target+" "+result)
	}
	fmt.Println(strings.Join(results, ", "))
	return savePings(db, pings)
}

// pingRound pings the router, if the box has a route to one, and then each
// of internetTargets, once each. It does not wait for the network or the
// clock: a round with no route or an unset clock is recorded as it is.
func pingRound(h host) ([]Ping, error) {
	// Whether the clock has been set is read before the clock. Read the
	// other way round, a clock set between the two would be saved as it was
	// before, and marked as set.
	_, err := h.ReadFile(synchronizedPath)
	round := Ping{ClockSynchronized: err == nil, At: h.Now(), Version: version}

	bootID, err := h.ReadFile(bootIDPath)
	if err != nil {
		return nil, fmt.Errorf("read the boot id: %w", err)
	}
	round.BootID = strings.TrimSpace(bootID)

	uptime, err := h.ReadFile(uptimePath)
	if err != nil {
		return nil, fmt.Errorf("read the uptime: %w", err)
	}
	seconds, _, _ := strings.Cut(strings.TrimSpace(uptime), " ")
	if round.Uptime, err = strconv.ParseFloat(seconds, 64); err != nil {
		return nil, fmt.Errorf("read the uptime: %w", err)
	}

	targets := internetTargets
	if router := defaultRouter(h.Run("ip", "route", "show", "default")); router != "" {
		targets = append([]string{router}, targets...)
	}
	var pings []Ping
	for _, target := range targets {
		p := round
		p.Target = target
		// One packet, waiting up to 2 seconds for its reply. -n prints
		// addresses as they are, so ping makes no name lookup of its own.
		p.Received, p.RTT, p.Output = pingResult(h.Run("ping", "-n", "-c", "1", "-W", "2", target))
		pings = append(pings, p)
	}
	return pings, nil
}

// defaultRouter reads the router's address from `ip route show default`:
//
//	default via 192.168.0.1 dev eth0 proto dhcp src 192.168.0.105 metric 100
//
// It returns "" when there is no such route. That includes ip failing: a
// round then has no ping of a router, the same as with no route.
func defaultRouter(output string) string {
	for _, line := range strings.Split(output, "\n") {
		if f := strings.Fields(line); len(f) >= 3 && f[0] == "default" && f[1] == "via" {
			return f[2]
		}
	}
	return ""
}

// What ping prints for a reply, and in its closing count.
var (
	pingTime     = regexp.MustCompile(`time=([0-9.]+) ms`)
	pingReceived = regexp.MustCompile(`([0-9]+) received`)
)

// pingResult reads what ping printed for one packet: how many replies it
// counted and the reply's round-trip time in ms. When there is no reply time
// to read, it returns the whole of what ping printed, to be kept as it is.
func pingResult(output string) (received int, rtt *float64, kept string) {
	if m := pingReceived.FindStringSubmatch(output); m != nil {
		received, _ = strconv.Atoi(m[1])
	}
	if m := pingTime.FindStringSubmatch(output); m != nil && received > 0 {
		if ms, err := strconv.ParseFloat(m[1], 64); err == nil {
			return received, &ms, ""
		}
	}
	return received, nil, strings.TrimSpace(output)
}
