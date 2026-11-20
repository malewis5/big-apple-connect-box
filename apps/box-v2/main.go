// Command big-apple-connect-box runs one M-Lab speed test (NDT7: a download,
// then an upload) and saves M-Lab's summary of it to scans.db.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/spec"
)

const (
	databasePath  = "scans.db"
	clientName    = "big-apple-connect-box" // reported to M-Lab
	clientVersion = "v2-dev"
	testTimeout   = 55 * time.Second // M-Lab's ndt7-client default

	// minDuration is M-Lab's own rule for a complete test: its published
	// data counts only tests that ran at least 9 seconds. A full download or
	// upload runs for about 10.
	minDuration = 9 * time.Second
)

func main() {
	log.SetFlags(0)
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	db, err := openDB(databasePath)
	if err != nil {
		return err
	}
	defer db.Close()

	scan, testErr := runTest(ndt7.NewClient(clientName, clientVersion))
	report(scan)
	id, err := save(db, scan)
	if err != nil {
		return errors.Join(testErr, err)
	}
	fmt.Printf("scan %d saved to %s\n", id, databasePath)
	if testErr != nil {
		return fmt.Errorf("scan %d failed: %w", id, testErr)
	}
	return nil
}

// runTest runs the download and then the upload, as M-Lab's ndt7-client
// does. The scan passed when the error is nil.
func runTest(client *ndt7.Client) (Scan, error) {
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	scan := Scan{StartedAt: time.Now()}
	var errs []error
	for _, t := range []struct {
		kind  spec.TestKind
		start func(context.Context) (<-chan spec.Measurement, error)
	}{
		{spec.TestDownload, client.StartDownload},
		{spec.TestUpload, client.StartUpload},
	} {
		measurements, err := t.start(ctx)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", t.kind, err))
			continue
		}
		for range measurements { // the client keeps the latest; wait for the end
		}
		// The client doesn't report a test that broke off partway, so check
		// how long it ran.
		if d := duration(t.kind, client.Results()[t.kind]); d < minDuration {
			errs = append(errs, fmt.Errorf("%s: stopped after %s; a complete test runs at least %s",
				t.kind, d.Round(100*time.Millisecond), minDuration))
		}
	}
	scan.CompletedAt = time.Now()

	scan.Summary = makeSummary(client.FQDN, client.Results())
	scan.Passed = len(errs) == 0
	return scan, errors.Join(errs...)
}

// duration returns how long a test ran, read from the counter M-Lab's
// summary divides by for that test's throughput.
func duration(kind spec.TestKind, r *ndt7.LatestMeasurements) time.Duration {
	var microseconds int64
	if kind == spec.TestDownload {
		if a := r.Client.AppInfo; a != nil {
			microseconds = a.ElapsedTime
		}
	} else if t := r.Server.TCPInfo; t != nil {
		microseconds = t.ElapsedTime
	}
	return time.Duration(microseconds) * time.Microsecond
}

// report prints a scan's results.
func report(scan Scan) {
	result := "failed"
	if scan.Passed {
		result = "passed"
	}
	fmt.Printf("%s in %s\n", result, scan.CompletedAt.Sub(scan.StartedAt).Round(time.Second))
	fmt.Printf("  Server:   %s (%s)\n", str(scan.ServerFQDN), str(scan.ServerIP))
	fmt.Printf("  Download: %s Mbit/s, latency %s ms, retransmission %s %%\n",
		num(scan.DownloadThroughput), num(scan.DownloadLatency), num(scan.DownloadRetransmission))
	fmt.Printf("  Upload:   %s Mbit/s, latency %s ms\n",
		num(scan.UploadThroughput), num(scan.UploadLatency))
}

func str(v *string) string {
	if v == nil {
		return "-"
	}
	return *v
}

func num(v *float64) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("%.1f", *v)
}
