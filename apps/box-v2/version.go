package main

import "runtime/debug"

// version is the build of this program. It is reported to M-Lab and saved
// with every scan.
var version = versionOf(debug.ReadBuildInfo())

// versionOf reads the commit that `go build` records in a program. It returns
// the commit's first 12 characters, followed by "+dirty" when the files had
// changes that weren't in that commit. A program without that record, as
// `go run` and `go test` build it, is "dev".
func versionOf(info *debug.BuildInfo, ok bool) string {
	if !ok {
		return "dev"
	}
	var revision string
	var dirty bool
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	if revision == "" {
		return "dev"
	}
	if len(revision) > 12 {
		revision = revision[:12]
	}
	if dirty {
		revision += "+dirty"
	}
	return revision
}
