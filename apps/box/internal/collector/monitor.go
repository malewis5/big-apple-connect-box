package collector

// The monitor: while the program runs with -daemon, it sends one ping to the
// router, the provider's first stop and two internet addresses every ten
// seconds, and keeps every ping as sent and as answered.
//
// It decides nothing. There is no "up" or "down" and no outage here: those
// are worked out afterwards from the pings, by whatever rule the reader
// chooses, and the pings stay as the evidence.

import (
	"context"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	// MonitorInterval is how often the monitor pings every target.
	MonitorInterval = 10 * time.Second

	// monitorTimeout bounds one ping command, so a stuck command can't
	// delay the next round.
	monitorTimeout = 8 * time.Second
)

var (
	// The internet addresses the monitor pings. Addresses, not names, so a
	// name lookup failing doesn't look like the connection failing.
	monitorDestinations = []string{"1.1.1.1", "8.8.8.8"}

	// One reply line, on each system:
	//
	//	macOS, Linux: 64 bytes from 1.1.1.1: icmp_seq=0 ttl=57 time=5.123 ms
	//	Windows:      Reply from 1.1.1.1: bytes=32 time=6ms TTL=57
	//	Windows:      Reply from 192.168.0.1: bytes=32 time<1ms TTL=64
	replyTimePattern = regexp.MustCompile(`\bfrom\b.*\btime([=<])([\d.]+) ?ms`)
	replyTTLPattern  = regexp.MustCompile(`(?i)\bttl=(\d+)`)
)

// A Ping is one ping the monitor sent: one packet to one address, and what
// came back.
type Ping struct {
	Subject    string // gateway, last_mile or internet
	Target     string // the address pinged
	StartedAt  time.Time
	FinishedAt time.Time
	RTT        string // round-trip time in ms, as reported; empty when no reply was read
	TTL        string // as reported; empty when no reply was read
	Raw        string // the reply line RTT and TTL were read from
	Output     string // everything ping printed, kept when no reply line was found
	Err        string // the error returned, word for word
}

// PingCommand is the command the monitor runs for each target on this
// system, with the address shown as "<target>".
func PingCommand() string {
	return strings.Join(hostPlatform().ping("<target>", 1), " ")
}

// TargetsFrom returns the router and last-mile addresses a scan found, or ""
// for one it did not find.
func TargetsFrom(scan *Scan) (gateway, lastMile string) {
	for _, p := range scan.Probes {
		if v := find(p.Observations, subjectGateway, "ip_address"); v != "" && gateway == "" {
			gateway = v
		}
		if v := find(p.Observations, subjectLastMile, "ip_address"); v != "" && lastMile == "" {
			lastMile = v
		}
	}
	return gateway, lastMile
}

// PingAll sends one ping to each target at the same time: the router and the
// last-mile stop when known, then the internet addresses. The pings come back
// in that order.
func PingAll(gateway, lastMile string) []Ping {
	var pings []Ping
	add := func(subject, host string) {
		if host != "" {
			pings = append(pings, Ping{Subject: subject, Target: host})
		}
	}
	add(subjectGateway, gateway)
	add(subjectLastMile, lastMile)
	for _, host := range monitorDestinations {
		add(subjectInternet, host)
	}

	os := hostPlatform()
	var wg sync.WaitGroup
	for i := range pings {
		wg.Go(func() { pingOnce(os, &pings[i]) })
	}
	wg.Wait()
	return pings
}

// pingOnce runs the system's ping for one packet and records what came back.
func pingOnce(os platform, p *Ping) {
	args := os.ping(p.Target, 1)
	ctx, cancel := context.WithTimeout(context.Background(), monitorTimeout)
	defer cancel()
	p.StartedAt = time.Now().UTC()
	out, err := exec.CommandContext(ctx, args[0], args[1:]...).CombinedOutput()
	p.FinishedAt = time.Now().UTC()
	if err != nil {
		p.Err = err.Error()
	}
	p.RTT, p.TTL, p.Raw = parseReply(string(out))
	if p.Raw == "" {
		p.Output = string(out)
	}
}

// parseReply reads the first reply line from a ping's output. A round-trip
// time Windows prints as "time<1ms" is kept as "<1", as reported. It returns
// empty strings when no reply line was found: a timeout, an unreachable
// address or an error.
func parseReply(text string) (rtt, ttl, raw string) {
	for _, line := range lines(text) {
		m := replyTimePattern.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		rtt = m[2]
		if m[1] == "<" {
			rtt = "<" + rtt
		}
		if t := replyTTLPattern.FindStringSubmatch(line); t != nil {
			ttl = t[1]
		}
		return rtt, ttl, strings.TrimSpace(line)
	}
	return "", "", ""
}
