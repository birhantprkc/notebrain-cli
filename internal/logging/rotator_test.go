package logging

import (
	"os"
	"path/filepath"
	"testing"
)

var chunk = []byte("0123456789\n") // 11 bytes

func TestRotatingWriterRotatesAndPrunes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nb.log")

	w, err := NewRotatingWriter(path, 100, 2)
	if err != nil {
		t.Fatalf("NewRotatingWriter: %v", err)
	}
	for i := range 40 {
		if _, err := w.Write(chunk); err != nil {
			t.Fatalf("Write %d: %v", i, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	for _, name := range []string{"nb.log", "nb.log.1", "nb.log.2"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected %s to exist: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "nb.log.3")); err == nil {
		t.Error("expected nb.log.3 to be pruned")
	}

	info, err := os.Stat(filepath.Join(dir, "nb.log"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() > 100+int64(len(chunk)) {
		t.Errorf("active log file too large: %d bytes", info.Size())
	}
}

func TestRotatingWriterReopenPreservesBackups(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nb.log")

	w, err := NewRotatingWriter(path, 100, 3)
	if err != nil {
		t.Fatalf("NewRotatingWriter: %v", err)
	}
	for range 12 {
		if _, err := w.Write(chunk); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "nb.log.1")); err != nil {
		t.Fatalf("expected nb.log.1 after first run: %v", err)
	}

	// Simulate a restart with a lower backup cap: the existing backup must
	// be counted, and any shift must prune beyond the new cap.
	w2, err := NewRotatingWriter(path, 100, 1)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	for range 30 {
		if _, err := w2.Write(chunk); err != nil {
			t.Fatalf("Write after reopen: %v", err)
		}
	}
	if err := w2.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "nb.log.1")); err != nil {
		t.Errorf("expected nb.log.1 to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "nb.log.2")); err == nil {
		t.Error("expected nb.log.2 to be pruned")
	}
}

func TestRotatingWriterAppendsAfterReopen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nb.log")

	w, err := NewRotatingWriter(path, 1<<20, 2)
	if err != nil {
		t.Fatalf("NewRotatingWriter: %v", err)
	}
	if _, err := w.Write([]byte("first\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	w2, err := NewRotatingWriter(path, 1<<20, 2)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if _, err := w2.Write([]byte("second\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := w2.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "first\nsecond\n" {
		t.Errorf("expected append across reopen, got %q", string(data))
	}
}

func TestRotatingWriterCreatesParentDir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "logs", "nb.log")
	w, err := NewRotatingWriter(path, 100, 2)
	if err != nil {
		t.Fatalf("NewRotatingWriter: %v", err)
	}
	if _, err := w.Write([]byte("hello\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected log file to exist: %v", err)
	}
}
