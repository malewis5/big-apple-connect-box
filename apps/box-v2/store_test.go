package main

import (
	"database/sql"
	"math"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// Every column is read back, each with a different value, so a column saved
// under the wrong name shows up.
func TestSave(t *testing.T) {
	db, err := openDB(filepath.Join(t.TempDir(), "scans.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// A time outside UTC is stored as UTC.
	started := time.Date(2026, 11, 19, 10, 4, 5, 123_000_000, time.FixedZone("EST", -5*60*60))
	passed := Scan{
		StartedAt:   started,
		CompletedAt: started.Add(21 * time.Second),
		Version:     "5d607fc5af54",
		Passed:      true,
		Summary: Summary{
			ServerFQDN:             text("ndt-mlab1-lga05.mlab-oti.measurement-lab.org"),
			ServerIP:               text("198.51.100.7"),
			ClientIP:               text("203.0.113.5"),
			DownloadUUID:           text("dl-uuid"),
			DownloadThroughput:     number(93.5),
			DownloadLatency:        number(12.5),
			DownloadRetransmission: number(0.25),
			UploadUUID:             text("ul-uuid"),
			UploadThroughput:       number(21.25),
			UploadLatency:          number(15.75),
		},
	}
	failed := Scan{
		StartedAt:   started,
		CompletedAt: started,
		Version:     "5d607fc5af54+dirty",
		Error:       "download: no route to host\nupload: no route to host",
	}
	for _, scan := range []Scan{passed, failed} {
		if _, err := save(db, scan); err != nil {
			t.Fatal(err)
		}
	}

	want := [][]any{
		{
			int64(1), "2026-11-19T15:04:05.123Z", "2026-11-19T15:04:26.123Z",
			"5d607fc5af54", int64(1), nil, // a scan that passed has no error
			"ndt-mlab1-lga05.mlab-oti.measurement-lab.org", "198.51.100.7", "203.0.113.5",
			"dl-uuid", 93.5, 12.5, 0.25,
			"ul-uuid", 21.25, 15.75,
		},
		// A failed scan is still saved, with why it failed, and NULL where
		// M-Lab returned nothing.
		{
			int64(2), "2026-11-19T15:04:05.123Z", "2026-11-19T15:04:05.123Z",
			"5d607fc5af54+dirty", int64(0), "download: no route to host\nupload: no route to host",
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		},
	}
	if got := readScans(t, db); !reflect.DeepEqual(got, want) {
		t.Errorf("scans:\n got  %v\n want %v", got, want)
	}
}

// Opening an existing database keeps its scans.
func TestOpenExisting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scans.db")
	for i := range 2 {
		db, err := openDB(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := save(db, Scan{StartedAt: time.Now(), CompletedAt: time.Now()}); err != nil {
			t.Fatal(err)
		}
		if n := len(readScans(t, db)); n != i+1 {
			t.Errorf("open %d: %d scans, want %d", i+1, n, i+1)
		}
		db.Close()
	}
}

// readScans returns every column of every scan, in the table's column order.
func readScans(t *testing.T, db *sql.DB) [][]any {
	t.Helper()
	return readRows(t, db, `SELECT * FROM scans ORDER BY id`)
}

// readRows returns every value of every row a query returns.
func readRows(t *testing.T, db *sql.DB, query string) [][]any {
	t.Helper()
	rows, err := db.Query(query)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		t.Fatal(err)
	}

	var scans [][]any
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		if err := rows.Scan(pointers...); err != nil {
			t.Fatal(err)
		}
		scans = append(scans, values)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return scans
}

// Every column of a round is read back as it was saved: a reply as its
// round-trip time, and a ping with no reply as what ping printed.
func TestSavePings(t *testing.T) {
	db, err := openDB(filepath.Join(t.TempDir(), "scans.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// A time outside UTC is stored as UTC.
	round := Ping{
		At:                time.Date(2026, 11, 20, 10, 41, 0, 12_000_000, time.FixedZone("EST", -5*60*60)),
		ClockSynchronized: true,
		BootID:            "be4c3e01-e1a0-4a1f-bd4f-03b2e4364578",
		Uptime:            57630.62,
		Version:           "cc63008199f7",
	}
	router, internet := round, round
	router.Target, router.Received, router.RTT = "192.168.0.1", 1, number(0.283)
	internet.Target, internet.Output = "1.1.1.1", "1 packets transmitted, 0 received, 100% packet loss, time 0ms"
	// A later round, before the clock was set.
	unset := Ping{At: round.At.Add(time.Minute), BootID: "7c1e", Uptime: 48.25, Version: "cc63008199f7", Target: "8.8.8.8", Output: "ping: connect: Network is unreachable"}

	if err := savePings(db, []Ping{router, internet}); err != nil {
		t.Fatal(err)
	}
	if err := savePings(db, []Ping{unset}); err != nil {
		t.Fatal(err)
	}

	want := [][]any{
		{int64(1), "2026-11-20T15:41:00.012Z", int64(1), "be4c3e01-e1a0-4a1f-bd4f-03b2e4364578", 57630.62, "cc63008199f7", "192.168.0.1", int64(1), 0.283, nil},
		{int64(2), "2026-11-20T15:41:00.012Z", int64(1), "be4c3e01-e1a0-4a1f-bd4f-03b2e4364578", 57630.62, "cc63008199f7", "1.1.1.1", int64(0), nil, "1 packets transmitted, 0 received, 100% packet loss, time 0ms"},
		{int64(3), "2026-11-20T15:42:00.012Z", int64(0), "7c1e", 48.25, "cc63008199f7", "8.8.8.8", int64(0), nil, "ping: connect: Network is unreachable"},
	}
	if got := readRows(t, db, `SELECT * FROM pings ORDER BY id`); !reflect.DeepEqual(got, want) {
		t.Errorf("pings:\n got  %v\n want %v", got, want)
	}
}

// A scan and a round of pings can finish at the same moment. A save that
// finds the database busy waits for it instead of failing.
func TestSaveWhileBusy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scans.db")
	db, err := openDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	other, err := openDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()

	for _, c := range []struct {
		name string
		save func() error
	}{
		{"a scan", func() error { _, err := save(db, Scan{StartedAt: time.Now(), CompletedAt: time.Now()}); return err }},
		{"a round of pings", func() error { return savePings(db, []Ping{{At: time.Now(), Target: "1.1.1.1"}}) }},
	} {
		// The other program is partway through saving.
		busy, err := other.Begin()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := busy.Exec(`INSERT INTO pings (at, clock_synchronized, boot_id, uptime, version, target, received) VALUES ('', 0, '', 0.0, '', 'other', 0)`); err != nil {
			t.Fatal(err)
		}
		released := make(chan error, 1)
		go func() {
			time.Sleep(300 * time.Millisecond)
			released <- busy.Commit()
		}()

		if err := c.save(); err != nil {
			t.Errorf("%s: %v", c.name, err)
		}
		if err := <-released; err != nil {
			t.Fatal(err)
		}
	}
	if scans, pings := len(readScans(t, db)), len(readRows(t, db, `SELECT id FROM pings WHERE target <> 'other'`)); scans != 1 || pings != 1 {
		t.Errorf("%d scans and %d pings saved, want 1 of each", scans, pings)
	}
}

// A round is saved whole or not at all, so that no round is ever short of
// a ping that was made.
func TestSavePingsAllOrNone(t *testing.T) {
	db, err := openDB(filepath.Join(t.TempDir(), "scans.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	good := Ping{At: time.Now(), BootID: "be4c", Uptime: 60, Target: "192.168.0.1", Received: 1, RTT: number(0.283)}
	bad := good
	bad.Target, bad.Uptime = "1.1.1.1", math.NaN() // not a number, which the uptime column refuses
	if err := savePings(db, []Ping{good, bad}); err == nil {
		t.Error("no error, want one")
	}
	if n := len(readRows(t, db, `SELECT id FROM pings`)); n != 0 {
		t.Errorf("%d pings saved, want none", n)
	}
}
