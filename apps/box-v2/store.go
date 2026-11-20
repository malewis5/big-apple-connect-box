package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Times are stored as UTC with millisecond precision.
const timeFormat = "2006-01-02T15:04:05.000Z"

// busyTimeout is how long, in ms, a save waits for the database when another
// run of the program is saving to it. A scan and a round of pings can finish
// at the same moment; each takes a few ms to save.
const busyTimeout = "5000"

// One row per scan. Every column after error is M-Lab's summary field of
// the same name; NULL where M-Lab's client left it unset.
const schema = `
CREATE TABLE IF NOT EXISTS scans (
	id                     INTEGER PRIMARY KEY,
	started_at             TEXT    NOT NULL,
	completed_at           TEXT    NOT NULL,
	version                TEXT    NOT NULL, -- the build of this program that ran the scan
	passed                 INTEGER NOT NULL CHECK (passed IN (0, 1)),
	error                  TEXT,             -- why the scan failed; NULL when it passed
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

-- One row per ping: what was pinged and what came back, with the state of
-- the box when the ping's round began. The pings of one round hold the same
-- at, boot_id and uptime. Nothing here is a conclusion: which rounds make an
-- outage, and when the box was off, are worked out from these rows.
CREATE TABLE IF NOT EXISTS pings (
	id                 INTEGER PRIMARY KEY,
	at                 TEXT    NOT NULL, -- the box's clock
	clock_synchronized INTEGER NOT NULL CHECK (clock_synchronized IN (0, 1)), -- whether the time service had set that clock since the box started
	boot_id            TEXT    NOT NULL, -- the kernel's id for this start-up
	uptime             REAL    NOT NULL, -- the kernel's seconds since the box started
	version            TEXT    NOT NULL, -- the build of this program
	target             TEXT    NOT NULL, -- the address pinged
	received           INTEGER NOT NULL, -- replies ping counted; 0 too when it printed no count
	rtt_ms             REAL,             -- the reply's round-trip time
	output             TEXT              -- what ping printed, or why it couldn't be run, when there was no reply time to read
) STRICT;
`

// A Scan is one run of the M-Lab test.
type Scan struct {
	StartedAt   time.Time
	CompletedAt time.Time
	// Version is the build of this program that ran the scan.
	Version string
	// Passed is true when the download and the upload both started and both
	// ran to completion.
	Passed bool
	// Error is why the scan failed. It is empty when the scan passed.
	Error string
	Summary
}

// openDB opens the database at path, creating its tables if needed.
func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=busy_timeout("+busyTimeout+")")
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
			started_at, completed_at, version, passed, error,
			ServerFQDN, ServerIP, ClientIP,
			DownloadUUID, DownloadThroughput, DownloadLatency, DownloadRetransmission,
			UploadUUID, UploadThroughput, UploadLatency
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		scan.StartedAt.UTC().Format(timeFormat), scan.CompletedAt.UTC().Format(timeFormat),
		scan.Version, scan.Passed, text(scan.Error),
		scan.ServerFQDN, scan.ServerIP, scan.ClientIP,
		scan.DownloadUUID, scan.DownloadThroughput, scan.DownloadLatency, scan.DownloadRetransmission,
		scan.UploadUUID, scan.UploadThroughput, scan.UploadLatency,
	)
	if err != nil {
		return 0, fmt.Errorf("save scan: %w", err)
	}
	return res.LastInsertId()
}

// savePings stores the pings of one round, all of them or none.
func savePings(db *sql.DB, pings []Ping) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("save pings: %w", err)
	}
	defer tx.Rollback()
	for _, p := range pings {
		_, err := tx.Exec(`
			INSERT INTO pings (at, clock_synchronized, boot_id, uptime, version, target, received, rtt_ms, output)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			p.At.UTC().Format(timeFormat), p.ClockSynchronized, p.BootID, p.Uptime, p.Version,
			p.Target, p.Received, p.RTT, text(p.Output),
		)
		if err != nil {
			return fmt.Errorf("save pings: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("save pings: %w", err)
	}
	return nil
}
