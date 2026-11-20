package main

import (
	"net"

	"github.com/m-lab/ndt7-client-go"
	"github.com/m-lab/ndt7-client-go/spec"
)

// Summary holds the values M-Lab's ndt7-client prints at the end of a test,
// under M-Lab's names. Units are M-Lab's: throughput in Mbit/s, latency in
// ms, retransmission in %.
//
// M-Lab's client defines Upload.Retransmission but never sets it, so it is
// not here.
//
// A value M-Lab's client leaves unset is nil here, where M-Lab prints zero,
// so a test that didn't run can't be mistaken for one that measured 0.
type Summary struct {
	ServerFQDN             *string
	ServerIP               *string
	ClientIP               *string
	DownloadUUID           *string
	DownloadThroughput     *float64
	DownloadLatency        *float64
	DownloadRetransmission *float64
	UploadUUID             *string
	UploadThroughput       *float64
	UploadLatency          *float64
}

// makeSummary computes the summary exactly as makeSummary does in
// github.com/m-lab/ndt7-client-go v0.10.1, internal/runner/runner.go. That
// package is internal, so it can't be imported.
func makeSummary(fqdn string, results map[spec.TestKind]*ndt7.LatestMeasurements) Summary {
	s := Summary{ServerFQDN: text(fqdn)}
	var server, client string

	if dl, ok := results[spec.TestDownload]; ok && dl != nil {
		if c := dl.ConnectionInfo; c != nil {
			s.DownloadUUID = text(c.UUID)
			client, server = c.Client, c.Server
		}
		// Throughput at the receiver: the client.
		if a := dl.Client.AppInfo; a != nil && a.ElapsedTime > 0 {
			elapsed := float64(a.ElapsedTime) / 1e06
			s.DownloadThroughput = number(8.0 * float64(a.NumBytes) / elapsed / (1000.0 * 1000.0))
		}
		if t := dl.Server.TCPInfo; t != nil {
			// Retransmission and latency at the sender: the server.
			if t.BytesSent > 0 {
				s.DownloadRetransmission = number(float64(t.BytesRetrans) / float64(t.BytesSent) * 100)
			}
			s.DownloadLatency = number(float64(t.MinRTT) / 1000)
		}
	}

	if ul, ok := results[spec.TestUpload]; ok && ul != nil {
		if c := ul.ConnectionInfo; c != nil {
			s.UploadUUID = text(c.UUID)
			client, server = c.Client, c.Server
		}
		if t := ul.Server.TCPInfo; t != nil {
			// Throughput and latency at the receiver: the server.
			if t.ElapsedTime > 0 {
				elapsed := float64(t.ElapsedTime) / 1e06
				s.UploadThroughput = number(8.0 * float64(t.BytesReceived) / elapsed / (1000.0 * 1000.0))
			}
			s.UploadLatency = number(float64(t.MinRTT) / 1000)
		}
	}

	if ip, _, err := net.SplitHostPort(client); err == nil {
		s.ClientIP = text(ip)
	}
	if ip, _, err := net.SplitHostPort(server); err == nil {
		s.ServerIP = text(ip)
	}
	return s
}

// text returns nil for an empty string: a value M-Lab's client left unset,
// or the error of a scan that passed.
func text(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func number(v float64) *float64 { return &v }
