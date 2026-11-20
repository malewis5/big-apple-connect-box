package main

import (
	"runtime/debug"
	"testing"
)

func TestVersionOf(t *testing.T) {
	commit := debug.BuildSetting{Key: "vcs.revision", Value: "5d607fc5af54c0dd2295e99ef5a200e34dd49b0a"}
	for _, c := range []struct {
		name     string
		settings []debug.BuildSetting
		ok       bool
		want     string
	}{
		{"built from a commit", []debug.BuildSetting{commit, {Key: "vcs.modified", Value: "false"}}, true, "5d607fc5af54"},
		{"built with uncommitted changes", []debug.BuildSetting{commit, {Key: "vcs.modified", Value: "true"}}, true, "5d607fc5af54+dirty"},
		{"no commit recorded", []debug.BuildSetting{{Key: "GOOS", Value: "linux"}}, true, "dev"},
		{"no build record", nil, false, "dev"},
	} {
		if got := versionOf(&debug.BuildInfo{Settings: c.settings}, c.ok); got != c.want {
			t.Errorf("%s: %q, want %q", c.name, got, c.want)
		}
	}
}
