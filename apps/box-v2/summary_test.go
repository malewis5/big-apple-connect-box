package main

import (
	"testing"

	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/spec"
	"github.com/m-lab/tcp-info/tcp"
)

func TestMakeSummary(t *testing.T) {
	results := map[spec.TestKind]*ndt7.LatestMeasurements{
		spec.TestDownload: {
			ConnectionInfo: &spec.ConnectionInfo{UUID: "dl-uuid", Client: "203.0.113.5:54321", Server: "198.51.100.7:443"},
			Client: spec.Measurement{
				// 12.5 MB in 1 s is 100 Mbit/s.
				AppInfo: &spec.AppInfo{NumBytes: 12_500_000, ElapsedTime: 1_000_000},
			},
			Server: spec.Measurement{
				TCPInfo: &spec.TCPInfo{LinuxTCPInfo: tcp.LinuxTCPInfo{
					BytesSent: 100_000, BytesRetrans: 1_000, MinRTT: 12_500,
				}},
			},
		},
		spec.TestUpload: {
			// The upload's addresses replace the download's, as in M-Lab's client.
			ConnectionInfo: &spec.ConnectionInfo{UUID: "ul-uuid", Client: "[2001:db8::5]:54322", Server: "[2001:db8::7]:443"},
			Server: spec.Measurement{
				TCPInfo: &spec.TCPInfo{
					// 2.5 MB in 1 s is 20 Mbit/s.
					LinuxTCPInfo: tcp.LinuxTCPInfo{BytesReceived: 2_500_000, MinRTT: 15_000},
					ElapsedTime:  1_000_000,
				},
			},
		},
	}

	s := makeSummary("ndt-mlab1-lga05.mlab-oti.measurement-lab.org", results)

	checkText(t, "ServerFQDN", s.ServerFQDN, "ndt-mlab1-lga05.mlab-oti.measurement-lab.org")
	checkText(t, "ServerIP", s.ServerIP, "2001:db8::7")
	checkText(t, "ClientIP", s.ClientIP, "2001:db8::5")
	checkText(t, "DownloadUUID", s.DownloadUUID, "dl-uuid")
	checkNumber(t, "DownloadThroughput", s.DownloadThroughput, 100)
	checkNumber(t, "DownloadLatency", s.DownloadLatency, 12.5)
	checkNumber(t, "DownloadRetransmission", s.DownloadRetransmission, 1)
	checkText(t, "UploadUUID", s.UploadUUID, "ul-uuid")
	checkNumber(t, "UploadThroughput", s.UploadThroughput, 20)
	checkNumber(t, "UploadLatency", s.UploadLatency, 15)
}

// When nothing ran, every value is unset rather than zero.
func TestMakeSummaryNothingRan(t *testing.T) {
	s := makeSummary("", map[spec.TestKind]*ndt7.LatestMeasurements{})
	if s != (Summary{}) {
		t.Errorf("got %+v, want every field nil", s)
	}
}

// M-Lab's client sets throughput only when time elapsed, and retransmission
// only when bytes were sent; latency whenever the server sent TCP_INFO.
func TestMakeSummaryPartial(t *testing.T) {
	s := makeSummary("server", map[spec.TestKind]*ndt7.LatestMeasurements{
		spec.TestDownload: {
			Client: spec.Measurement{AppInfo: &spec.AppInfo{NumBytes: 10}},
			Server: spec.Measurement{TCPInfo: &spec.TCPInfo{}},
		},
	})
	if s.DownloadThroughput != nil || s.DownloadRetransmission != nil {
		t.Errorf("throughput %v, retransmission %v: want both unset", s.DownloadThroughput, s.DownloadRetransmission)
	}
	checkNumber(t, "DownloadLatency", s.DownloadLatency, 0)
	if s.DownloadUUID != nil || s.ClientIP != nil || s.UploadThroughput != nil {
		t.Errorf("got %+v: want no UUID, address or upload", s)
	}
}

func checkText(t *testing.T, name string, got *string, want string) {
	t.Helper()
	if got == nil || *got != want {
		t.Errorf("%s = %v, want %q", name, deref(got), want)
	}
}

func checkNumber(t *testing.T, name string, got *float64, want float64) {
	t.Helper()
	if got == nil || *got != want {
		t.Errorf("%s = %v, want %v", name, deref(got), want)
	}
}

func deref[T any](v *T) any {
	if v == nil {
		return nil
	}
	return *v
}
