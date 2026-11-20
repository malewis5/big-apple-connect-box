package main

import (
	"errors"
	"os"
	"os/exec"
	"time"
)

// A host is the box the program runs on: its clock, what its files hold and
// what its commands print.
type host interface {
	Now() time.Time
	ReadFile(path string) (string, error)
	// Run runs a command and returns everything it printed. When the
	// command can't be run, Run returns the reason instead.
	Run(name string, args ...string) string
}

// system is the box itself.
type system struct{}

func (system) Now() time.Time { return time.Now() }

func (system) ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	return string(content), err
}

func (system) Run(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), "LC_ALL=C") // print in English, whatever the box's language
	output, err := cmd.CombinedOutput()
	// A command that ran and failed has printed why. One that couldn't be
	// run has printed nothing, so the reason is all there is.
	var failed *exec.ExitError
	if err != nil && !errors.As(err, &failed) {
		return err.Error()
	}
	return string(output)
}
