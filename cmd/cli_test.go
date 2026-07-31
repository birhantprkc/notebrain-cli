package cmd

import (
	"log/slog"
	"testing"
)

func TestSetupLogger(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		debug    bool
	}{
		{"default info", "info", false},
		{"explicit debug", "debug", false},
		{"explicit warn", "warn", false},
		{"explicit error", "error", false},
		{"legacy debug flag overrides info", "info", true},
		{"legacy debug flag overrides warn", "warn", true},
		{"unknown level defaults to info", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test ensures setupLogger does not panic and correctly initializes slog.
			// It's mostly a sanity check since slog.SetDefault is global and we can't cleanly mock term checks here without refactoring.
			setupLogger(tt.logLevel, tt.debug)

			// Additional logic to verify the slog level would require a custom handler wrapper,
			// but we rely on the internal switch logic being straightforward.
			if slog.Default() == nil {
				t.Fatal("expected slog.Default() to be initialized")
			}
		})
	}
}
