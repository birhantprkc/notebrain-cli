package main

import (
	"testing"
)

func TestRunMainExitCodes(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want int
	}{
		{"usage error", []string{"--bogus-flag"}, exitUsage},
		{"runtime error", []string{"search"}, exitError},
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
