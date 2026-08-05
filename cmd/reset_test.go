package cmd

import (
	"context"
	"testing"
)

func TestResetCmdYesSkipsPrompt(t *testing.T) {
	fs := &fakeStore{}
	withFakeStore(t, fs)

	cmd := &ResetCmd{Yes: true}
	if err := cmd.Run(&Globals{Ctx: context.Background()}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if fs.resetCalls != 1 {
		t.Errorf("resetCalls = %d, want 1", fs.resetCalls)
	}
}

func TestResetCmdPromptConfirms(t *testing.T) {
	fs := &fakeStore{}
	withFakeStore(t, fs)
	withStdin(t, "yes\n")

	if err := (&ResetCmd{}).Run(&Globals{Ctx: context.Background()}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if fs.resetCalls != 1 {
		t.Errorf("resetCalls = %d, want 1", fs.resetCalls)
	}
}

func TestResetCmdPromptDeclines(t *testing.T) {
	fs := &fakeStore{}
	withFakeStore(t, fs)
	withStdin(t, "no\n")

	if err := (&ResetCmd{}).Run(&Globals{Ctx: context.Background()}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if fs.resetCalls != 0 {
		t.Errorf("resetCalls = %d, want 0 (declined)", fs.resetCalls)
	}
}
