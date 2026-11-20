package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Run returns everything a command printed, on either output and whatever
// its exit status, since ping prints its reason for failing and then fails.
func TestSystemRun(t *testing.T) {
	if got, want := (system{}).Run("sh", "-c", "echo printed; echo complained >&2; exit 2"), "printed\ncomplained\n"; got != want {
		t.Errorf("output %q, want %q", got, want)
	}
}

// Commands print in English, which is what pingResult reads.
func TestSystemRunInEnglish(t *testing.T) {
	t.Setenv("LC_ALL", "de_DE.UTF-8")
	if got := (system{}).Run("sh", "-c", "echo $LC_ALL"); got != "C\n" {
		t.Errorf("LC_ALL %q, want C", got)
	}
}

// A command that can't be run gives the reason, so it is saved like any
// other output.
func TestSystemRunMissingCommand(t *testing.T) {
	got := (system{}).Run("no-such-command-on-this-box")
	if !strings.Contains(got, "no-such-command-on-this-box") || !strings.Contains(got, "not found") {
		t.Errorf("output %q, want the reason it couldn't be run", got)
	}
}

func TestSystemReadFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "uptime")
	if err := os.WriteFile(path, []byte("57630.62 230240.11\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got, err := (system{}).ReadFile(path); err != nil || got != "57630.62 230240.11\n" {
		t.Errorf("read %q, error %v", got, err)
	}
	if _, err := (system{}).ReadFile(path + ".missing"); err == nil {
		t.Error("no error for a missing file")
	}
}

func TestSystemNow(t *testing.T) {
	before := time.Now()
	if got := (system{}).Now(); got.Before(before) || got.After(time.Now()) {
		t.Errorf("now %s, want the current time", got)
	}
}
