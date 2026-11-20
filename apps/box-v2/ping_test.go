package main

import (
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// fakeHost is a box with the files and the command output given.
type fakeHost struct {
	now     time.Time
	files   map[string]string
	outputs map[string]string // what each command line prints

	ran   []string // the command lines run, in order
	calls []string // everything asked of the box, in order
}

// Now moves on a second each time it is asked, as a real clock would.
func (h *fakeHost) Now() time.Time {
	h.calls = append(h.calls, "clock")
	now := h.now
	h.now = h.now.Add(time.Second)
	return now
}

func (h *fakeHost) ReadFile(path string) (string, error) {
	h.calls = append(h.calls, "read "+path)
	content, ok := h.files[path]
	if !ok {
		return "", fs.ErrNotExist
	}
	return content, nil
}

func (h *fakeHost) Run(name string, args ...string) string {
	line := strings.Join(append([]string{name}, args...), " ")
	h.ran = append(h.ran, line)
	h.calls = append(h.calls, "run "+line)
	return h.outputs[line]
}

// The Raspberry Pi's own files and routing table, word for word.
const (
	bootIDFile   = "be4c3e01-e1a0-4a1f-bd4f-03b2e4364578\n"
	uptimeFile   = "57630.62 230240.11\n"
	defaultRoute = "default via 192.168.0.1 dev eth0 proto dhcp src 192.168.0.105 metric 100 \n"
	pingRouter   = "ping -n -c 1 -W 2 192.168.0.1"
	pingOne      = "ping -n -c 1 -W 2 1.1.1.1"
	pingEight    = "ping -n -c 1 -W 2 8.8.8.8"
	showRoute    = "ip route show default"
	testBootID   = "be4c3e01-e1a0-4a1f-bd4f-03b2e4364578"
	testUptime   = 57630.62
)

// piHost is the Raspberry Pi with its clock synchronized and a router.
func piHost(outputs map[string]string) *fakeHost {
	outputs[showRoute] = defaultRoute
	return &fakeHost{
		now: time.Date(2026, 11, 20, 15, 41, 0, 0, time.UTC),
		files: map[string]string{
			"/proc/sys/kernel/random/boot_id":    bootIDFile,
			"/proc/uptime":                       uptimeFile,
			"/run/systemd/timesync/synchronized": "",
		},
		outputs: outputs,
	}
}

// What ping printed on the Raspberry Pi, word for word.
const (
	pingReply = `PING 192.168.0.1 (192.168.0.1) 56(84) bytes of data.
64 bytes from 192.168.0.1: icmp_seq=1 ttl=64 time=0.283 ms

--- 192.168.0.1 ping statistics ---
1 packets transmitted, 1 received, 0% packet loss, time 0ms
rtt min/avg/max/mdev = 0.283/0.283/0.283/0.000 ms
`
	pingNoReply = `PING 192.0.2.1 (192.0.2.1) 56(84) bytes of data.

--- 192.0.2.1 ping statistics ---
1 packets transmitted, 0 received, 100% packet loss, time 0ms

`
	pingNoRoute = "ping: connect: Network is unreachable\n"
	// Something on the way answered that the address can't be reached.
	pingUnreachable = `PING 192.168.0.250 (192.168.0.250) 56(84) bytes of data.
From 192.168.0.105 icmp_seq=1 Destination Host Unreachable

--- 192.168.0.250 ping statistics ---
1 packets transmitted, 0 received, +1 errors, 100% packet loss, time 0ms

`
)

// A reply is saved as its round-trip time. Anything else is saved as what
// ping printed, word for word.
func TestPingResult(t *testing.T) {
	const notFound = `exec: "ping": executable file not found in $PATH`
	for _, c := range []struct {
		name     string
		output   string
		received int
		rtt      *float64
		kept     string
	}{
		{"a reply", pingReply, 1, number(0.283), ""},
		{"sent, and nothing came back", pingNoReply, 0, nil, strings.TrimSpace(pingNoReply)},
		{"told it can't be reached", pingUnreachable, 0, nil, strings.TrimSpace(pingUnreachable)},
		{"couldn't be sent", pingNoRoute, 0, nil, "ping: connect: Network is unreachable"},
		// A reply line with no count after it, as if ping was cut off: nothing is read into it.
		{"a reply and no count", "64 bytes from 192.168.0.1: icmp_seq=1 ttl=64 time=0.283 ms\n", 0, nil, "64 bytes from 192.168.0.1: icmp_seq=1 ttl=64 time=0.283 ms"},
		{"ping couldn't be run", notFound, 0, nil, notFound},
		{"nothing printed", "", 0, nil, ""},
	} {
		received, rtt, kept := pingResult(c.output)
		if received != c.received || kept != c.kept {
			t.Errorf("%s: received %d, kept %q; want %d, %q", c.name, received, kept, c.received, c.kept)
		}
		if (rtt == nil) != (c.rtt == nil) || rtt != nil && *rtt != *c.rtt {
			t.Errorf("%s: round trip %v, want %v", c.name, deref(rtt), deref(c.rtt))
		}
	}
}

// A round pings the router and two internet addresses. Every ping carries
// the same reading of the clock, the boot id and the uptime, taken once when
// the round began.
func TestPingRound(t *testing.T) {
	host := piHost(map[string]string{pingRouter: pingReply, pingOne: pingNoReply, pingEight: pingReply})
	began := host.now

	pings, err := pingRound(host)
	if err != nil {
		t.Fatal(err)
	}

	round := Ping{At: began, ClockSynchronized: true, BootID: testBootID, Uptime: testUptime, Version: version}
	router, one, eight := round, round, round
	router.Target, router.Received, router.RTT = "192.168.0.1", 1, number(0.283)
	one.Target, one.Output = "1.1.1.1", strings.TrimSpace(pingNoReply)
	eight.Target, eight.Received, eight.RTT = "8.8.8.8", 1, number(0.283)
	if want := []Ping{router, one, eight}; !reflect.DeepEqual(pings, want) {
		t.Errorf("pings:\n got  %+v\n want %+v", pings, want)
	}
	if want := []string{showRoute, pingRouter, pingOne, pingEight}; !reflect.DeepEqual(host.ran, want) {
		t.Errorf("ran %q, want %q", host.ran, want)
	}
	// The box's state is read once, before anything is pinged. Whether the
	// clock has been set is read before the clock itself: read the other way
	// round, a clock set between the two would be saved wrong and marked right.
	want := []string{
		"read /run/systemd/timesync/synchronized", "clock",
		"read /proc/sys/kernel/random/boot_id", "read /proc/uptime",
		"run " + showRoute, "run " + pingRouter, "run " + pingOne, "run " + pingEight,
	}
	if !reflect.DeepEqual(host.calls, want) {
		t.Errorf("asked of the box:\n got  %q\n want %q", host.calls, want)
	}
}

// The router is the first default route's "via" address.
func TestDefaultRouter(t *testing.T) {
	for _, c := range []struct{ name, output, want string }{
		{"one route", defaultRoute, "192.168.0.1"},
		{"two routes: the first, which the box prefers", "default via 192.168.0.1 dev eth0 proto dhcp metric 100 \ndefault via 10.0.0.1 dev wlan0 proto dhcp metric 600 \n", "192.168.0.1"},
		{"a route with no router", "default dev ppp0 scope link \n", ""},
		{"no route", "", ""},
		{"ip couldn't be run", `exec: "ip": executable file not found in $PATH`, ""},
	} {
		if got := defaultRouter(c.output); got != c.want {
			t.Errorf("%s: %q, want %q", c.name, got, c.want)
		}
	}
}

// With no route there is no router to ping. The two internet addresses are
// still pinged, and what ping says about them is kept.
func TestPingRoundNoRoute(t *testing.T) {
	host := piHost(map[string]string{pingOne: pingNoRoute, pingEight: pingNoRoute})
	host.outputs[showRoute] = ""
	began := host.now

	pings, err := pingRound(host)
	if err != nil {
		t.Fatal(err)
	}

	round := Ping{At: began, ClockSynchronized: true, BootID: testBootID, Uptime: testUptime, Version: version, Output: "ping: connect: Network is unreachable"}
	one, eight := round, round
	one.Target, eight.Target = "1.1.1.1", "8.8.8.8"
	if want := []Ping{one, eight}; !reflect.DeepEqual(pings, want) {
		t.Errorf("pings:\n got  %+v\n want %+v", pings, want)
	}
}

// Until the time service has set the clock, a round says so. It does not
// wait.
func TestPingRoundClockNotSynchronized(t *testing.T) {
	host := piHost(map[string]string{pingRouter: pingReply, pingOne: pingNoReply, pingEight: pingNoReply})
	delete(host.files, "/run/systemd/timesync/synchronized")

	pings, err := pingRound(host)
	if err != nil {
		t.Fatal(err)
	}
	if len(pings) != 3 {
		t.Fatalf("%d pings, want 3", len(pings))
	}
	for _, p := range pings {
		if p.ClockSynchronized {
			t.Errorf("%s: clock synchronized, want not", p.Target)
		}
	}
}

// A round that can't read the boot id or the uptime pings nothing: its rows
// could not be placed in time.
func TestPingRoundUnreadableBox(t *testing.T) {
	for _, missing := range []string{"/proc/sys/kernel/random/boot_id", "/proc/uptime"} {
		host := piHost(map[string]string{})
		delete(host.files, missing)
		pings, err := pingRound(host)
		if err == nil || len(pings) != 0 || len(host.ran) != 0 {
			t.Errorf("without %s: %d pings, ran %q, error %v; want an error and nothing run", missing, len(pings), host.ran, err)
		}
	}
}

// A run saves its round. Pings that got no reply are what it is there to
// record, so they are not an error.
func TestRunPings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scans.db")
	host := piHost(map[string]string{pingRouter: pingReply, pingOne: pingNoReply, pingEight: pingNoRoute})

	if err := runPings(path, host); err != nil {
		t.Errorf("error %v, want none", err)
	}

	db, err := openDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	want := [][]any{
		{"2026-11-20T15:41:00.000Z", int64(1), testBootID, testUptime, version, "192.168.0.1", int64(1), 0.283, nil},
		{"2026-11-20T15:41:00.000Z", int64(1), testBootID, testUptime, version, "1.1.1.1", int64(0), nil, strings.TrimSpace(pingNoReply)},
		{"2026-11-20T15:41:00.000Z", int64(1), testBootID, testUptime, version, "8.8.8.8", int64(0), nil, "ping: connect: Network is unreachable"},
	}
	if got := readRows(t, db, `SELECT at, clock_synchronized, boot_id, uptime, version, target, received, rtt_ms, output FROM pings ORDER BY id`); !reflect.DeepEqual(got, want) {
		t.Errorf("saved:\n got  %v\n want %v", got, want)
	}
}

// When the database can't be opened there is nowhere to save a round, so
// nothing is pinged.
func TestRunPingsCannotOpenDatabase(t *testing.T) {
	host := piHost(map[string]string{})
	err := runPings(filepath.Join(t.TempDir(), "missing", "scans.db"), host)
	if err == nil || len(host.ran) != 0 {
		t.Errorf("error %v, ran %q: want an error and nothing run", err, host.ran)
	}
}

// A round that can't be placed in time is an error, and saves nothing.
func TestRunPingsUnreadableBox(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scans.db")
	host := piHost(map[string]string{})
	delete(host.files, "/proc/uptime")

	if err := runPings(path, host); err == nil {
		t.Error("no error, want one")
	}
	db, err := openDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if n := len(readRows(t, db, `SELECT id FROM pings`)); n != 0 {
		t.Errorf("%d pings saved, want none", n)
	}
}
