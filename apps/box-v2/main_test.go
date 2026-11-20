package main

import (
	"testing"
	"time"

	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/spec"
	"github.com/m-lab/tcp-info/tcp"
)

// A test's duration comes from the counter its throughput is computed from:
// the client's for a download, the server's for an upload.
func TestDuration(t *testing.T) {
	download := &ndt7.LatestMeasurements{
		Client: spec.Measurement{AppInfo: &spec.AppInfo{ElapsedTime: 9_800_000}},
		// The server's counter is not the download's duration.
		Server: spec.Measurement{TCPInfo: &spec.TCPInfo{ElapsedTime: 1}},
	}
	upload := &ndt7.LatestMeasurements{
		Client: spec.Measurement{AppInfo: &spec.AppInfo{ElapsedTime: 1}},
		Server: spec.Measurement{TCPInfo: &spec.TCPInfo{
			LinuxTCPInfo: tcp.LinuxTCPInfo{BytesReceived: 1},
			ElapsedTime:  3_200_000,
		}},
	}
	for _, c := range []struct {
		name    string
		kind    spec.TestKind
		results *ndt7.LatestMeasurements
		want    time.Duration
	}{
		{"complete download", spec.TestDownload, download, 9800 * time.Millisecond},
		{"upload that broke off", spec.TestUpload, upload, 3200 * time.Millisecond},
		{"download with no measurement", spec.TestDownload, &ndt7.LatestMeasurements{}, 0},
		{"upload with no measurement", spec.TestUpload, &ndt7.LatestMeasurements{}, 0},
	} {
		if got := duration(c.kind, c.results); got != c.want {
			t.Errorf("%s: %s, want %s", c.name, got, c.want)
		}
	}
}
