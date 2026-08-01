package main

import (
	"testing"
)

func TestRunMainExitCodes(t *testing.T) {
	t.Setenv("HOME", t.TempDir()) // no user config, no user chroma path
	tests := []struct {
		name string
		args []string
		want int
	}{
		{"usage error", []string{"--bogus-flag"}, exitUsage},
		{"missing query", []string{"search"}, exitUsage}, // missing query is a usage error
		{"missing vault", []string{"ingest"}, exitUsage}, // missing --vault-path is a usage error
		{"bad log level", []string{"search", "q", "--log-level=verbose"}, exitUsage},
		{"version subcommand", []string{"version"}, exitOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runMain(tt.args); got != tt.want {
				t.Errorf("runMain(%v) = %d, want %d", tt.args, got, tt.want)
			}
		})
	}
}

func TestPanicRecovery(t *testing.T) {
	code := 0
	func() {
		defer recoverMain(&code)
		panic("boom")
	}()
	if code != exitError {
		t.Errorf("panic recovery exit code = %d, want %d", code, exitError)
	}
}
