package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctorCmdFreshVault(t *testing.T) {
	globals := &Globals{
		Ctx:        context.Background(),
		ChromaPath: t.TempDir(),
	}
	if err := (&DoctorCmd{}).Run(globals); err != nil {
		t.Fatalf("doctor on fresh vault: %v", err)
	}
}

func TestDoctorCmdUnconfiguredChroma(t *testing.T) {
	globals := &Globals{Ctx: context.Background()}
	if err := (&DoctorCmd{}).Run(globals); err != nil {
		t.Fatalf("doctor without chroma path: %v", err)
	}
}

func TestDoctorCmdCorruptDatabase(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "chroma.sqlite3"), []byte("garbage-not-sqlite"), 0600); err != nil {
		t.Fatal(err)
	}
	origProbe := probeStoreOpenFn
	probeStoreOpenFn = func(string) probeResult {
		return probeResult{detail: "stubbed probe failure"}
	}
	t.Cleanup(func() { probeStoreOpenFn = origProbe })

	globals := &Globals{
		Ctx:        context.Background(),
		ChromaPath: dir,
	}
	err := (&DoctorCmd{}).Run(globals)
	if err == nil {
		t.Fatal("expected doctor to report the corrupt database")
	}
	if !strings.Contains(err.Error(), "database problem") {
		t.Errorf("error %q should mention the problem count", err)
	}
}
