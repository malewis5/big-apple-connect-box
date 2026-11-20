package main

import (
	"database/sql"
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
