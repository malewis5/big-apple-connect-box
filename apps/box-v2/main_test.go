package main

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/spec"
	"github.com/m-lab/tcp-info/tcp"
)

// fakeClient stands in for M-Lab's client. Each half of the test either
// fails to start or ends at once, leaving the measurements given for it.
type fakeClient struct {
	fqdn     string
	startErr map[spec.TestKind]error
	measured map[spec.TestKind]*ndt7.LatestMeasurements

	started []spec.TestKind // the halves asked to start, in order
	results map[spec.TestKind]*ndt7.LatestMeasurements
}

func (f *fakeClient) start(kind spec.TestKind) (<-chan spec.Measurement, error) {
	f.started = append(f.started, kind)
	// As in M-Lab's client, a half has an empty result from the moment it
	// is asked to start.
	if f.results == nil {
		f.results = map[spec.TestKind]*ndt7.LatestMeasurements{}
	}
	f.results[kind] = &ndt7.LatestMeasurements{}
	if err := f.startErr[kind]; err != nil {
		return nil, err
	}
	if m := f.measured[kind]; m != nil {
		f.results[kind] = m
	}
	ended := make(chan spec.Measurement)
	close(ended)
	return ended, nil
}

func (f *fakeClient) StartDownload(context.Context) (<-chan spec.Measurement, error) {
	return f.start(spec.TestDownload)
}

func (f *fakeClient) StartUpload(context.Context) (<-chan spec.Measurement, error) {
	return f.start(spec.TestUpload)
}

func (f *fakeClient) Results() map[spec.TestKind]*ndt7.LatestMeasurements { return f.results }

func (f *fakeClient) ServerFQDN() string { return f.fqdn }

// downloaded returns what a download at 100 Mbit/s leaves after running for d.
func downloaded(d time.Duration) *ndt7.LatestMeasurements {
	return &ndt7.LatestMeasurements{
		ConnectionInfo: &spec.ConnectionInfo{UUID: "dl-uuid", Client: "203.0.113.5:54321", Server: "198.51.100.7:443"},
		Client: spec.Measurement{AppInfo: &spec.AppInfo{
			NumBytes:    int64(12_500_000 * d.Seconds()),
			ElapsedTime: d.Microseconds(),
		}},
	}
}

// uploaded returns what an upload at 20 Mbit/s leaves after running for d.
func uploaded(d time.Duration) *ndt7.LatestMeasurements {
	return &ndt7.LatestMeasurements{
		ConnectionInfo: &spec.ConnectionInfo{UUID: "ul-uuid", Client: "203.0.113.5:54322", Server: "198.51.100.7:443"},
		Server: spec.Measurement{TCPInfo: &spec.TCPInfo{
			LinuxTCPInfo: tcp.LinuxTCPInfo{BytesReceived: int64(2_500_000 * d.Seconds())},
			ElapsedTime:  d.Microseconds(),
		}},
	}
}

// both is the two halves, in the order a scan runs them.
var both = []spec.TestKind{spec.TestDownload, spec.TestUpload}

// A download and an upload that both run to the end make a scan that passed.
func TestRunTestComplete(t *testing.T) {
	client := &fakeClient{
		fqdn: "ndt-mlab1-lga05.mlab-oti.measurement-lab.org",
		measured: map[spec.TestKind]*ndt7.LatestMeasurements{
			spec.TestDownload: downloaded(10 * time.Second),
			spec.TestUpload:   uploaded(10 * time.Second),
		},
	}
	before := time.Now()
	scan, err := runTest(client)
	if err != nil {
		t.Fatalf("error %v, want none", err)
	}
	if !scan.Passed || scan.Error != "" {
		t.Errorf("passed %v, error %q: want passed and no error", scan.Passed, scan.Error)
	}
	if !reflect.DeepEqual(client.started, both) {
		t.Errorf("started %v, want %v", client.started, both)
	}
	if scan.StartedAt.Before(before) || scan.CompletedAt.Before(scan.StartedAt) {
		t.Errorf("started %s, completed %s: want now, in that order", scan.StartedAt, scan.CompletedAt)
	}
	if scan.Version != version {
		t.Errorf("version %q, want %q", scan.Version, version)
	}
	checkText(t, "ServerFQDN", scan.ServerFQDN, "ndt-mlab1-lga05.mlab-oti.measurement-lab.org")
	checkNumber(t, "DownloadThroughput", scan.DownloadThroughput, 100)
	checkNumber(t, "UploadThroughput", scan.UploadThroughput, 20)
}

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

// A scan fails when either half can't start or stops early. The other half
// still runs, and what M-Lab did measure is kept.
func TestRunTestFailed(t *testing.T) {
	const full, short = 10 * time.Second, 3200 * time.Millisecond
	noRoute := errors.New("no route to host")
	for _, c := range []struct {
		name         string
		startErr     map[spec.TestKind]error
		download     *ndt7.LatestMeasurements
		upload       *ndt7.LatestMeasurements
		wantError    string
		wantDownload bool // whether a download throughput is kept
		wantUpload   bool
	}{
		{
			name:       "download can't start",
			startErr:   map[spec.TestKind]error{spec.TestDownload: noRoute},
			upload:     uploaded(full),
			wantError:  "download: no route to host",
			wantUpload: true,
		},
		{
			name:         "upload can't start",
			startErr:     map[spec.TestKind]error{spec.TestUpload: noRoute},
			download:     downloaded(full),
			wantError:    "upload: no route to host",
			wantDownload: true,
		},
		{
			name:      "neither can start",
			startErr:  map[spec.TestKind]error{spec.TestDownload: noRoute, spec.TestUpload: noRoute},
			wantError: "download: no route to host\nupload: no route to host",
		},
		{
			name:         "download stops early",
			download:     downloaded(short),
			upload:       uploaded(full),
			wantError:    "download: stopped after 3.2s; a complete test runs at least 9s",
			wantDownload: true,
			wantUpload:   true,
		},
		{
			name:         "upload stops early",
			download:     downloaded(full),
			upload:       uploaded(short),
			wantError:    "upload: stopped after 3.2s; a complete test runs at least 9s",
			wantDownload: true,
			wantUpload:   true,
		},
		{
			name:         "download just short of complete",
			download:     downloaded(8900 * time.Millisecond),
			upload:       uploaded(full),
			wantError:    "download: stopped after 8.9s; a complete test runs at least 9s",
			wantDownload: true,
			wantUpload:   true,
		},
		{
			name:       "download ends with nothing measured",
			upload:     uploaded(full),
			wantError:  "download: stopped after 0s; a complete test runs at least 9s",
			wantUpload: true,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			client := &fakeClient{
				startErr: c.startErr,
				measured: map[spec.TestKind]*ndt7.LatestMeasurements{
					spec.TestDownload: c.download,
					spec.TestUpload:   c.upload,
				},
			}
			scan, err := runTest(client)
			if err == nil || err.Error() != c.wantError {
				t.Errorf("error %v, want %q", err, c.wantError)
			}
			if scan.Passed || scan.Error != c.wantError {
				t.Errorf("passed %v, error %q: want failed with %q", scan.Passed, scan.Error, c.wantError)
			}
			if !reflect.DeepEqual(client.started, both) {
				t.Errorf("started %v, want %v", client.started, both)
			}
			if got := scan.DownloadThroughput != nil; got != c.wantDownload {
				t.Errorf("download throughput kept: %v, want %v", got, c.wantDownload)
			}
			if got := scan.UploadThroughput != nil; got != c.wantUpload {
				t.Errorf("upload throughput kept: %v, want %v", got, c.wantUpload)
			}
		})
	}
}

// Nine seconds is the shortest complete test.
func TestRunTestShortestComplete(t *testing.T) {
	scan, err := runTest(&fakeClient{measured: map[spec.TestKind]*ndt7.LatestMeasurements{
		spec.TestDownload: downloaded(9 * time.Second),
		spec.TestUpload:   uploaded(9 * time.Second),
	}})
	if err != nil || !scan.Passed {
		t.Errorf("passed %v, error %v: want passed", scan.Passed, err)
	}
}

// A failed scan is still saved, with why it failed, and run returns the
// failure so the program exits with an error.
func TestRunFailed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scans.db")
	noRoute := errors.New("no route to host")
	client := &fakeClient{startErr: map[spec.TestKind]error{spec.TestDownload: noRoute, spec.TestUpload: noRoute}}

	err := run(path, client)

	const reason = "download: no route to host\nupload: no route to host"
	if want := "scan 1 failed: " + reason; err == nil || err.Error() != want {
		t.Errorf("error %v, want %q", err, want)
	}
	want := [][]any{{int64(0), reason, version, nil, nil}}
	if got := savedScans(t, path); !reflect.DeepEqual(got, want) {
		t.Errorf("saved:\n got  %v\n want %v", got, want)
	}
}

// savedScans returns the outcome of every scan in the database at path: passed, error,
// version and the two throughputs.
func savedScans(t *testing.T, path string) [][]any {
	t.Helper()
	db, err := openDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	return readRows(t, db, `SELECT passed, error, version, DownloadThroughput, UploadThroughput FROM scans ORDER BY id`)
}

// A scan that passed is saved with what it measured and no error, and run
// returns none.
func TestRunPassed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scans.db")
	client := &fakeClient{measured: map[spec.TestKind]*ndt7.LatestMeasurements{
		spec.TestDownload: downloaded(10 * time.Second),
		spec.TestUpload:   uploaded(10 * time.Second),
	}}

	if err := run(path, client); err != nil {
		t.Errorf("error %v, want none", err)
	}
	want := [][]any{{int64(1), nil, version, 100.0, 20.0}}
	if got := savedScans(t, path); !reflect.DeepEqual(got, want) {
		t.Errorf("saved:\n got  %v\n want %v", got, want)
	}
}

// When the database can't be opened there is nowhere to save a scan, so no
// test is run.
func TestRunCannotOpenDatabase(t *testing.T) {
	client := &fakeClient{}
	err := run(filepath.Join(t.TempDir(), "missing", "scans.db"), client)
	if err == nil {
		t.Error("no error, want one")
	}
	if len(client.started) != 0 {
		t.Errorf("started %v, want nothing", client.started)
	}
}

// M-Lab's own client, unable to connect, behaves as fakeClient assumes: each
// half returns an error and the scan fails with nothing measured. No
// connection is made.
func TestRunTestCannotConnect(t *testing.T) {
	client := ndt7.NewClient(clientName, version)
	client.Server = "ndt.example:443" // a named server, so M-Lab isn't asked for one
	client.Dialer.NetDialContext = func(context.Context, string, string) (net.Conn, error) {
		return nil, errors.New("no route to host")
	}

	scan, err := runTest(mlab{client})

	const want = "download: no route to host\nupload: no route to host"
	if err == nil || err.Error() != want {
		t.Errorf("error %v, want %q", err, want)
	}
	if scan.Passed || scan.Error != want {
		t.Errorf("passed %v, error %q: want failed with %q", scan.Passed, scan.Error, want)
	}
	// The client names the server it tried, and measured nothing.
	checkText(t, "ServerFQDN", scan.ServerFQDN, "ndt.example")
	if scan.DownloadThroughput != nil || scan.UploadThroughput != nil || scan.ClientIP != nil {
		t.Errorf("got %+v: want nothing measured", scan.Summary)
	}
}

// An argument the program doesn't know is an error. It must never fall
// through to a scan, which is a real speed test.
func TestStartUnknownArgument(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, args := range [][]string{{"pings"}, {"ping", "now"}, {"--help"}, {""}} {
		if err := start(args); err == nil || !strings.Contains(err.Error(), "usage") {
			t.Errorf("%q: error %v, want the usage", args, err)
		}
	}
	if _, err := os.Stat(databasePath); err == nil {
		t.Error("a database was created")
	}
}
