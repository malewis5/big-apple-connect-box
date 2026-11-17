package collector

import "testing"

func TestParseReply(t *testing.T) {
	cases := map[string]struct {
		output, rtt, ttl, raw string
	}{
		"macOS": {
			"PING 1.1.1.1 (1.1.1.1): 56 data bytes\n64 bytes from 1.1.1.1: icmp_seq=0 ttl=57 time=5.123 ms\n\n--- 1.1.1.1 ping statistics ---\n1 packets transmitted, 1 packets received, 0.0% packet loss\n",
			"5.123", "57", "64 bytes from 1.1.1.1: icmp_seq=0 ttl=57 time=5.123 ms",
		},
		"Linux": {
			"PING 1.1.1.1 (1.1.1.1) 56(84) bytes of data.\n64 bytes from 1.1.1.1: icmp_seq=1 ttl=57 time=5.12 ms\n\n--- 1.1.1.1 ping statistics ---\n1 packets transmitted, 1 received, 0% packet loss, time 0ms\n",
			"5.12", "57", "64 bytes from 1.1.1.1: icmp_seq=1 ttl=57 time=5.12 ms",
		},
		"Windows": {
			"\r\nPinging 1.1.1.1 with 32 bytes of data:\r\nReply from 1.1.1.1: bytes=32 time=6ms TTL=57\r\n\r\nPing statistics for 1.1.1.1:\r\n",
			"6", "57", "Reply from 1.1.1.1: bytes=32 time=6ms TTL=57",
		},
		"Windows, under a millisecond": {
			"Pinging 192.168.0.1 with 32 bytes of data:\r\nReply from 192.168.0.1: bytes=32 time<1ms TTL=64\r\n",
			"<1", "64", "Reply from 192.168.0.1: bytes=32 time<1ms TTL=64",
		},
		"Linux, no reply": {
			"PING 1.1.1.1 (1.1.1.1) 56(84) bytes of data.\n\n--- 1.1.1.1 ping statistics ---\n1 packets transmitted, 0 received, 100% packet loss, time 0ms\n",
			"", "", "",
		},
		"Linux, unreachable": {
			"PING 192.168.0.1 (192.168.0.1) 56(84) bytes of data.\nFrom 192.168.0.50 icmp_seq=1 Destination Host Unreachable\n",
			"", "", "",
		},
		"Windows, timed out": {
			"Pinging 1.1.1.1 with 32 bytes of data:\r\nRequest timed out.\r\n",
			"", "", "",
		},
		// Windows reports an unreachable address as a "Reply from" the
		// computer itself, with no time. It is not a reply from the target.
		"Windows, unreachable": {
			"Pinging 1.1.1.1 with 32 bytes of data:\r\nReply from 192.168.0.50: Destination host unreachable.\r\n",
			"", "", "",
		},
		"no output": {"", "", "", ""},
	}
	for name, c := range cases {
		rtt, ttl, raw := parseReply(c.output)
		if rtt != c.rtt || ttl != c.ttl || raw != c.raw {
			t.Errorf("%s: got %q %q %q, want %q %q %q", name, rtt, ttl, raw, c.rtt, c.ttl, c.raw)
		}
	}
}

func TestTargetsFrom(t *testing.T) {
	scan := &Scan{Probes: []*Probe{
		{Observations: []Observation{{Subject: subjectGateway, Attribute: "ip_address", Value: "192.168.0.1"}}},
		{Observations: []Observation{
			{Subject: subjectPath, Attribute: "hop_1_ip_address", Value: "192.168.0.1"},
			{Subject: subjectLastMile, Attribute: "ip_address", Value: "173.68.66.1"},
		}},
	}}
	if gateway, lastMile := TargetsFrom(scan); gateway != "192.168.0.1" || lastMile != "173.68.66.1" {
		t.Errorf("TargetsFrom = %q %q", gateway, lastMile)
	}
	if gateway, lastMile := TargetsFrom(&Scan{}); gateway != "" || lastMile != "" {
		t.Errorf("TargetsFrom(empty scan) = %q %q, want nothing", gateway, lastMile)
	}
}
