package mysql

import (
	"fmt"
	"io"
	"time"

	"github.com/go-mysql-org/go-mysql/replication"

	"db-log-explorer/internal/events"
)

// AnalysisProgress is emitted during AnalyzeStream.
type AnalysisProgress struct {
	SourceID          string
	BytesRead         int64
	FileSize          int64
	MinTimestamp      time.Time
	MaxTimestamp      time.Time
	ApproxChangeCount int64
}

// AnalyzeStream performs a lightweight scan for metadata without building summaries.
func AnalyzeStream(src *Source, onProgress func(AnalysisProgress)) (*events.FileAnalysisResult, error) {
	var result *events.FileAnalysisResult
	err := src.withFileLock(func() error {
		var err error
		result, err = analyzeStreamLocked(src, onProgress)
		return err
	})
	return result, err
}

func analyzeStreamLocked(src *Source, onProgress func(AnalysisProgress)) (*events.FileAnalysisResult, error) {
	src.State = StateAnalyzing
	src.BytesRead = 0

	if _, err := src.file.Seek(0, io.SeekStart); err != nil {
		src.MarkError(err)
		return nil, fmt.Errorf("rewind file: %w", err)
	}

	magic := make([]byte, 4)
	if _, err := io.ReadFull(src.file, magic); err != nil {
		src.MarkError(err)
		return nil, fmt.Errorf("read magic: %w", err)
	}

	var (
		minTS   time.Time
		maxTS   time.Time
		count   int64
		hasTS   bool
		parser  = src.newParser()
		file    = src.file
		lastPct = -1
	)

	emitProgress := func() {
		if onProgress == nil {
			return
		}
		pos := src.BytesRead
		onProgress(AnalysisProgress{
			SourceID:          src.ID,
			BytesRead:         pos,
			FileSize:          src.FileSize,
			MinTimestamp:      minTS,
			MaxTimestamp:      maxTS,
			ApproxChangeCount: count,
		})
	}

	for {
		offset, err := file.Seek(0, io.SeekCurrent)
		if err != nil {
			src.MarkError(err)
			return nil, err
		}

		done, err := parser.ParseSingleEvent(file, func(ev *replication.BinlogEvent) error {
			ts, userData := ClassifyEvent(ev)
			if userData {
				count++
				if !hasTS || ts.Before(minTS) {
					minTS = ts
				}
				if !hasTS || ts.After(maxTS) {
					maxTS = ts
				}
				hasTS = true
			}
			return nil
		})
		if err != nil {
			src.MarkError(err)
			return nil, fmt.Errorf("parse at offset %d: %w", offset, err)
		}

		if pos, err := file.Seek(0, io.SeekCurrent); err == nil {
			src.BytesRead = pos
		}

		if src.FileSize > 0 {
			pct := int(src.BytesRead * 100 / src.FileSize)
			if pct != lastPct {
				lastPct = pct
				emitProgress()
			}
		}

		if done {
			emitProgress()
			result := &events.FileAnalysisResult{
				SourceID:          src.ID,
				FileSize:          src.FileSize,
				FileSizeHuman:     HumanSize(src.FileSize),
				MinTimestamp:      minTS,
				MaxTimestamp:      maxTS,
				ApproxChangeCount: count,
				Complete:          true,
			}
			src.Analysis = result
			return result, nil
		}
	}
}

// HumanSize formats bytes for display.
func HumanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for n2 := n / unit; n2 >= unit; n2 /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}
