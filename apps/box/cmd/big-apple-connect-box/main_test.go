package main

import (
	"context"
	"errors"
	"testing"

	"github.com/malewis5/big-apple-connect-box/apps/box/internal/collector"
)

func TestRunMonitorSavesBeforeStopping(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // stop after the first round

	ping := func() []collector.Ping { return []collector.Ping{{Target: "1.1.1.1"}, {Target: "8.8.8.8"}} }
	var saved []collector.Ping
	save := func(p []collector.Ping) error { saved = append(saved, p...); return nil }

	if err := runMonitor(ctx, ping, save); err != nil {
		t.Fatal(err)
	}
	if len(saved) != 2 {
		t.Errorf("saved %d pings, want 2", len(saved))
	}
}

func TestRunMonitorReportsALostLastBatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ping := func() []collector.Ping { return []collector.Ping{{Target: "1.1.1.1"}} }
	failed := errors.New("disk full")
	save := func([]collector.Ping) error { return failed }

	if err := runMonitor(ctx, ping, save); !errors.Is(err, failed) {
		t.Errorf("runMonitor = %v, want the save error, so the run isn't marked as stopped cleanly", err)
	}
}
