// Copyright © 2026 nmdra. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package ingest

import (
	"log/slog"
	"math"
	"time"
)

// ProgressUpdate represents a status update during file processing.
type ProgressUpdate struct {
	Done    int
	Total   int
	Current string
}

// RunProgress logs ingestion progress: a debug event per file (quiet at the
// default info level) plus an info-level summary every 10 percent, so long
// runs keep the user informed without per-file noise.
func RunProgress(totalFiles int, progressCh <-chan ProgressUpdate) {
	start := time.Now()
	nextDecile := 10

	for u := range progressCh {
		percent := 0.0
		if totalFiles > 0 {
			percent = math.Round(float64(u.Done)/float64(totalFiles)*10000) / 100
		}
		slog.Debug("ingestion progress",
			"processed", u.Done,
			"total", totalFiles,
			"percent", percent,
			"current", u.Current,
			"elapsed_ms", time.Since(start).Milliseconds())

		if totalFiles > 0 && percent >= float64(nextDecile) && nextDecile <= 100 {
			slog.Info("ingestion progress",
				"processed", u.Done,
				"total", totalFiles,
				"percent", percent,
				"elapsed_ms", time.Since(start).Milliseconds())
			nextDecile += 10
		}
	}
	slog.Info("ingestion completed",
		"total_files", totalFiles,
		"duration_ms", time.Since(start).Milliseconds())
}
