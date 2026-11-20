package main

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// The pings table holds no conclusions. These queries are how conclusions
// are reasoned from its rows, and TestReasoning checks that the rows are
// enough for each one.
//
// Two limits hold whatever the query. A new boot id says the box started
// again, not why: a power cut, a pulled plug and a restart look the same.
// And a start-up whose clock was never set can be measured, in seconds
// since it began, but not placed on the clock.
const reasoning = `
WITH
-- A round is the pings that share one reading of the boot id and uptime.
-- A ping was sent if it got a reply or ping printed its count; ping ran if
-- it was sent or ping said the address or the network was unreachable.
rounds AS (
	SELECT boot_id, uptime, min(id) AS id, max(at) AS at, max(clock_synchronized) AS synced,
	       max(CASE WHEN target NOT IN ('1.1.1.1', '8.8.8.8') THEN received END) AS router,
	       max(CASE WHEN target IN ('1.1.1.1', '8.8.8.8') THEN received END) AS internet,
	       max(output IS NULL OR output LIKE '%packets transmitted%' OR output LIKE '%unreachable%') AS ran
	FROM pings GROUP BY boot_id, uptime
),
-- When each start-up began: the time of its latest round saved after the
-- clock was set, less that round's seconds since start-up. NULL for a
-- start-up whose clock was never set.
boots AS (
	SELECT boot_id, min(id) AS first_id, max(uptime) AS last_uptime,
	       (SELECT unixepoch(s.at, 'subsec') - s.uptime FROM rounds s
	        WHERE s.boot_id = r.boot_id AND s.synced = 1 ORDER BY s.id DESC LIMIT 1) AS started
	FROM rounds r GROUP BY boot_id
),
-- Every round with how far its pings got, and its true time whatever the
-- box's clock said then. That time is NULL in a start-up whose clock was
-- never set.
timeline AS (
	SELECT r.id, r.boot_id, r.uptime, r.at AS box_clock, r.synced, b.started, b.started + r.uptime AS t,
	       CASE WHEN r.internet = 1 THEN 'internet answered'
	            WHEN NOT r.ran THEN 'pings did not run'
	            WHEN r.router = 1 THEN 'router answered, internet silent'
	            ELSE 'router not reached' END AS state,
	       lag(r.uptime) OVER start_up AS previous_uptime, lead(r.uptime) OVER start_up AS next_uptime
	FROM rounds r JOIN boots b USING (boot_id)
	WINDOW start_up AS (PARTITION BY r.boot_id ORDER BY r.id)
),
-- Neighbouring rounds in the same state, as one stretch. Rounds are a minute
-- apart, so a step of more than 90 seconds means rounds are missing, and a
-- stretch never reaches across missing rounds or into another start-up. It
-- lasts until the next round begins, or to its own last round when none
-- follows.
stretches AS (
	SELECT boot_id, state, min(id) AS id, count(*) AS rounds, started,
	       min(uptime) AS began, max(CASE WHEN next_uptime - uptime <= 90 THEN next_uptime ELSE uptime END) AS ended
	FROM (SELECT *, sum(state IS NOT previous_state OR uptime - previous_uptime > 90) OVER (ORDER BY id) AS stretch
	      FROM (SELECT *, lag(state) OVER (PARTITION BY boot_id ORDER BY id) AS previous_state FROM timeline))
	GROUP BY stretch
)
`

func TestReasoning(t *testing.T) {
	db, err := openDB(filepath.Join(t.TempDir(), "scans.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	clock := func(hms string) time.Time {
		at, err := time.Parse(time.RFC3339, "2026-12-01T"+hms+"Z")
		if err != nil {
			t.Fatal(err)
		}
		return at
	}
	const (
		answered = "answered"
		silent   = "silent"
		noRoute  = "no route"
		notRun   = "not run"
	)
	// round saves one round at the true time at, of a start-up that began at
	// started. The box's clock is slow by the amount given.
	round := func(bootID string, started, at time.Time, synchronized bool, slow time.Duration, router, one, eight string) {
		t.Helper()
		reading := Ping{At: at.Add(-slow), ClockSynchronized: synchronized, BootID: bootID, Uptime: at.Sub(started).Seconds(), Version: "test"}
		var pings []Ping
		for _, p := range []struct{ target, result string }{{"192.168.0.1", router}, {"1.1.1.1", one}, {"8.8.8.8", eight}} {
			ping := reading
			ping.Target = p.target
			switch p.result {
			case "": // no route, so no router to ping
				continue
			case answered:
				ping.Received, ping.RTT = 1, number(0.3)
			case silent:
				ping.Output = "1 packets transmitted, 0 received, 100% packet loss, time 0ms"
			case noRoute:
				ping.Output = "ping: connect: Network is unreachable"
			case notRun:
				ping.Output = `exec: "ping": executable file not found in $PATH`
			}
			pings = append(pings, ping)
		}
		if err := savePings(db, pings); err != nil {
			t.Fatal(err)
		}
	}

	// The first start-up, at 09:59:30, with the clock right throughout.
	first := clock("09:59:30")
	for at := clock("10:00:00"); !at.After(clock("11:29:00")); at = at.Add(time.Minute) {
		router, one, eight := answered, answered, answered
		switch hm := at.Format("15:04"); {
		case hm >= "10:30" && hm <= "10:41": // the internet is down for 12 minutes
			one, eight = silent, silent
		case hm >= "10:50" && hm <= "10:52": // ping can't be run for 3 minutes
			router, one, eight = notRun, notRun, notRun
		case hm == "11:00": // one address misses once
			one = silent
		case hm == "11:05": // both miss, for a single round
			one, eight = silent, silent
		case hm >= "11:10" && hm <= "11:14": // the cable is out for 5 minutes
			router, one, eight = "", noRoute, noRoute
		case hm == "11:20" || hm == "11:25": // both miss, either side of missing rounds
			one, eight = silent, silent
		case hm >= "11:21" && hm <= "11:24": // no round is saved for 4 minutes
			continue
		}
		round("first", first, at, true, 0, router, one, eight)
	}
	// The power goes at about 11:29:40 and comes back at 11:40:00 with no
	// internet. The box's clock carries on from where it stopped, 10 minutes
	// 20 seconds slow. The power goes again at about 11:51 before the clock
	// is ever set.
	dark := clock("11:40:00")
	for at := clock("11:40:35"); !at.After(clock("11:50:35")); at = at.Add(time.Minute) {
		round("dark", dark, at, false, 10*time.Minute+20*time.Second, answered, silent, silent)
	}
	// The box starts again at 12:10:00, still with no internet and with its
	// clock now 29 minutes 20 seconds slow, until the internet and the time
	// server answer at 12:25:35.
	second := clock("12:10:00")
	for at := clock("12:10:35"); !at.After(clock("12:40:35")); at = at.Add(time.Minute) {
		if at.Before(clock("12:25:35")) {
			round("second", second, at, false, 29*time.Minute+20*time.Second, answered, silent, silent)
		} else {
			round("second", second, at, true, 0, answered, answered, answered)
		}
	}

	for _, c := range []struct {
		conclusion string
		query      string
		want       [][]any
	}{
		{
			// Between two start-ups whose clocks were set. A start-up in
			// between whose clock never was is known only by how long it ran.
			"when the box was off",
			`SELECT time(round(p.started + p.last_uptime), 'unixepoch'), time(round(b.started), 'unixepoch'),
			        round((b.started - p.started - p.last_uptime - unset.ran) / 60), unset.boots, round(unset.ran / 60)
			 FROM boots b
			 JOIN boots p ON p.first_id = (SELECT max(first_id) FROM boots WHERE first_id < b.first_id AND started IS NOT NULL)
			 JOIN (SELECT b.boot_id, count(x.boot_id) AS boots, coalesce(sum(x.last_uptime), 0) AS ran
			       FROM boots b LEFT JOIN boots x ON x.started IS NULL AND x.first_id < b.first_id
			            AND x.first_id > (SELECT max(first_id) FROM boots WHERE first_id < b.first_id AND started IS NOT NULL)
			       GROUP BY b.boot_id) unset USING (boot_id)
			 WHERE b.started IS NOT NULL`,
			// Last seen, started again, minutes off in all, start-ups in
			// between whose clock was never set, minutes those ran.
			[][]any{{"11:29:00", "12:10:00", 30.0, int64(1), 11.0}},
		},
		{
			"when the internet was out of reach for two rounds or more, and how far the pings got",
			`SELECT state, time(round(started + began), 'unixepoch'), time(round(started + ended), 'unixepoch'), rounds, ended - began, began < 120
			 FROM stretches WHERE state <> 'internet answered' AND rounds >= 2 ORDER BY id`,
			// What happened, from, until, rounds, seconds, whether it began as the box started.
			[][]any{
				{"router answered, internet silent", "10:30:00", "10:42:00", int64(12), 720.0, int64(0)},
				// Not an outage: nothing was measured.
				{"pings did not run", "10:50:00", "10:53:00", int64(3), 180.0, int64(0)},
				{"router not reached", "11:10:00", "11:15:00", int64(5), 300.0, int64(0)},
				// The start-up whose clock was never set: how long, but not when.
				{"router answered, internet silent", nil, nil, int64(11), 600.0, int64(1)},
				// Began as the box started: the wait for the line to come back after the power cut.
				{"router answered, internet silent", "12:10:35", "12:25:35", int64(15), 900.0, int64(1)},
			},
		},
		{
			// 11:20 and 11:25 are not one outage: the rounds between them are missing.
			"single rounds that missed, which are not claimed as outages",
			`SELECT state, time(round(started + began), 'unixepoch'), ended - began FROM stretches WHERE state <> 'internet answered' AND rounds = 1 ORDER BY id`,
			[][]any{
				{"router answered, internet silent", "11:05:00", 60.0},
				{"router answered, internet silent", "11:20:00", 0.0},
				{"router answered, internet silent", "11:25:00", 60.0},
			},
		},
		{
			"when the box was running and saved no round, which says nothing about the internet",
			`SELECT boot_id, time(round(started + previous_uptime), 'unixepoch'), time(round(t), 'unixepoch'), uptime - previous_uptime
			 FROM timeline WHERE uptime - previous_uptime > 90 ORDER BY id`,
			[][]any{{"first", "11:20:00", "11:25:00", 300.0}}, // the last round before, the first after, seconds
		},
		{
			"the true time of the first round of each start-up saved before the clock was set",
			`SELECT boot_id, time(box_clock), time(round(t), 'unixepoch'), uptime FROM timeline
			 WHERE id IN (SELECT min(id) FROM timeline WHERE synced = 0 GROUP BY boot_id) ORDER BY id`,
			// The start-up, the box's clock, the true time, seconds since start-up.
			[][]any{
				{"dark", "11:30:15", nil, 35.0}, // its clock was never set, so there is no true time
				{"second", "11:41:15", "12:10:35", 35.0},
			},
		},
		{
			"how many pings each address was sent, and how many it left unanswered",
			`SELECT target, sum(sent), sum(sent AND received = 0)
			 FROM (SELECT target, received, output IS NULL OR output LIKE '%packets transmitted%' AS sent FROM pings)
			 GROUP BY target ORDER BY target`,
			[][]any{{"1.1.1.1", int64(120), int64(42)}, {"192.168.0.1", int64(120), int64(0)}, {"8.8.8.8", int64(120), int64(41)}},
		},
	} {
		if got := readRows(t, db, reasoning+c.query); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s:\n got  %v\n want %v", c.conclusion, got, c.want)
		}
	}
}
