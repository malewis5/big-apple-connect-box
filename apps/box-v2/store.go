package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Times are stored as UTC with millisecond precision.
const timeFormat = "2006-01-02T15:04:05.000Z"

// One row per scan. Every column after passed is M-Lab's summary field of
// the same name; NULL where M-Lab's client left it unset.
const schema = `
CREATE TABLE IF NOT EXISTS scans (
	id                     INTEGER PRIMARY KEY,
	started_at             TEXT    NOT NULL,
	completed_at           TEXT    NOT NULL,
	passed                 INTEGER NOT NULL CHECK (passed IN (0, 1)),
	ServerFQDN             TEXT,
	ServerIP               TEXT,
	ClientIP               TEXT,
	DownloadUUID           TEXT,
	DownloadThroughput     REAL, -- Mbit/s
	DownloadLatency        REAL, -- ms
	DownloadRetransmission REAL, -- %
	UploadUUID             TEXT,
	UploadThroughput       REAL, -- Mbit/s
	UploadLatency          REAL  -- ms
) STRICT;
`

// A Scan is one run of the M-Lab test.
type Scan struct {
	StartedAt   time.Time
	CompletedAt time.Time
	// Passed is true when the download and the upload both started and both
	// ran to completion.
	Passed bool
	Summary
}

// openDB opens the database at path, creating the scans table if needed.
func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("create tables in %s: %w", path, err)
	}
	return db, nil
}

// save stores a scan and returns its id.
func save(db *sql.DB, scan Scan) (int64, error) {
	res, err := db.Exec(`
		INSERT INTO scans (
			started_at, completed_at, passed,
			ServerFQDN, ServerIP, ClientIP,
			DownloadUUID, DownloadThroughput, DownloadLatency, DownloadRetransmission,
			UploadUUID, UploadThroughput, UploadLatency
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		scan.StartedAt.UTC().Format(timeFormat), scan.CompletedAt.UTC().Format(timeFormat), scan.Passed,
		scan.ServerFQDN, scan.ServerIP, scan.ClientIP,
		scan.DownloadUUID, scan.DownloadThroughput, scan.DownloadLatency, scan.DownloadRetransmission,
		scan.UploadUUID, scan.UploadThroughput, scan.UploadLatency,
	)
	if err != nil {
		return 0, fmt.Errorf("save scan: %w", err)
	}
	return res.LastInsertId()
}
