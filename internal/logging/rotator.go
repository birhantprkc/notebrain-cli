// Package logging provides file-based logging helpers for the CLI.
package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Default rotation limits, applied when a zero or negative value is given.
const (
	DefaultMaxSize    = 10 << 20 // 10 MiB
	DefaultMaxBackups = 5
)

// RotatingWriter is an io.Writer that rolls the target log file over once it
// reaches maxSize bytes. Rotated files are kept as numbered backups
// (path.1, path.2, ...); backups beyond maxBackups are pruned. Backups are
// counted on open, so limits stay bounded across restarts.
type RotatingWriter struct {
	path       string
	maxSize    int64
	maxBackups int

	f       *os.File
	size    int64
	backups int
	err     error
}

// NewRotatingWriter opens (creating if needed) the log file at path. Parent
// directories are created as needed. maxSize and maxBackups fall back to the
// package defaults when zero or negative.
func NewRotatingWriter(path string, maxSize int64, maxBackups int) (*RotatingWriter, error) {
	if maxSize <= 0 {
		maxSize = DefaultMaxSize
	}
	if maxBackups <= 0 {
		maxBackups = DefaultMaxBackups
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("stat log file: %w", err)
	}
	return &RotatingWriter{
		path:       path,
		maxSize:    maxSize,
		maxBackups: maxBackups,
		f:          f,
		size:       info.Size(),
		backups:    existingBackups(path),
	}, nil
}

// Write appends p to the log file, rotating first when the size limit would
// be crossed. Errors are recorded and returned; callers that ignore them can
// inspect Err later.
func (w *RotatingWriter) Write(p []byte) (int, error) {
	if w.f == nil {
		return 0, os.ErrClosed
	}
	if w.size+int64(len(p)) > w.maxSize {
		if err := w.rotate(); err != nil {
			w.recordErr(err)
			return 0, err
		}
	}
	n, err := w.f.Write(p)
	w.size += int64(n)
	if err != nil {
		w.recordErr(err)
	}
	return n, err
}

// Close closes the current log file. Close on a closed writer is a no-op.
func (w *RotatingWriter) Close() error {
	if w.f == nil {
		return nil
	}
	err := w.f.Close()
	w.f = nil
	return err
}

// Err returns the first write or rotation error encountered, if any.
func (w *RotatingWriter) Err() error {
	return w.err
}

func (w *RotatingWriter) recordErr(err error) {
	if w.err == nil {
		w.err = err
	}
}

// rotate closes the current file, shifts backups up by one (pruning beyond
// maxBackups), and reopens a fresh file at path.
func (w *RotatingWriter) rotate() error {
	if err := w.f.Close(); err != nil {
		return err
	}
	w.f = nil

	// Shift in descending order so renames never clobber a file that still
	// needs to be renamed. Anything beyond the cap is removed instead.
	for i := w.backups; i >= 1; i-- {
		if i+1 > w.maxBackups {
			_ = os.Remove(w.backupPath(i))
		} else {
			_ = os.Rename(w.backupPath(i), w.backupPath(i+1))
		}
	}
	if err := os.Rename(w.path, w.backupPath(1)); err != nil {
		return fmt.Errorf("rotate log file: %w", err)
	}
	if w.backups < w.maxBackups {
		w.backups++
	}

	f, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("reopen log file: %w", err)
	}
	w.f = f
	w.size = 0
	return nil
}

func (w *RotatingWriter) backupPath(n int) string {
	return fmt.Sprintf("%s.%d", w.path, n)
}

// existingBackups counts the numeric backups already present next to path,
// so rotation bounds survive process restarts.
func existingBackups(path string) int {
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		return 0
	}
	base := filepath.Base(path)
	n := 0
	for _, e := range entries {
		name := e.Name()
		suffix, ok := strings.CutPrefix(name, base+".")
		if !ok || suffix == "" {
			continue
		}
		if allDigits(suffix) {
			n++
		}
	}
	return n
}

func allDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
