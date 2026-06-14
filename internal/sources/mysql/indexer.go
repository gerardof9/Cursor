package mysql

import (
	"fmt"
	"io"

	"github.com/go-mysql-org/go-mysql/replication"

	"db-log-explorer/internal/events"
)

const batchSize = 64

// IndexBatch is a chunk of summaries produced during stream-parse.
type IndexBatch struct {
	SourceID  string
	Summaries []events.EventSummary
	Done      bool
	Err       error
}

// IndexStream parses user-data events within scope from the current file position.
// When scope is nil, all events are indexed (legacy behavior).
func IndexStream(src *Source, scope *events.InvestigationScope, emit func(events.EventSummary)) error {
	return src.withFileLock(func() error {
		return indexStreamLocked(src, scope, emit)
	})
}

func indexStreamLocked(src *Source, scope *events.InvestigationScope, emit func(events.EventSummary)) error {
	src.State = StateIndexing
	src.IndexedCount = 0
	src.ActiveScope = scope

	if _, err := src.file.Seek(0, io.SeekStart); err != nil {
		src.MarkError(err)
		return fmt.Errorf("rewind file: %w", err)
	}

	magic := make([]byte, 4)
	if _, err := io.ReadFull(src.file, magic); err != nil {
		src.MarkError(err)
		return fmt.Errorf("read magic: %w", err)
	}

	parser := src.newParser()
	f := src.file
	pastTo := false

	for {
		offset, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			src.MarkError(err)
			return err
		}

		done, err := parser.ParseSingleEvent(f, func(ev *replication.BinlogEvent) error {
			if pastTo {
				return nil
			}

			ts, userData := ClassifyEvent(ev)
			if scope != nil {
				if ts.Before(scope.From) {
					return nil
				}
				if ts.After(scope.To) {
					pastTo = true
					return nil
				}
			}

			if !userData {
				return nil
			}

			summary, ok := MapEvent(ev, src, offset)
			if ok {
				emit(summary)
				src.IndexedCount++
			}
			return nil
		})
		if err != nil {
			src.MarkError(err)
			return fmt.Errorf("parse at offset %d: %w", offset, err)
		}
		if done || pastTo {
			src.MarkReady()
			return nil
		}

		if pos, err := f.Seek(0, io.SeekCurrent); err == nil {
			src.BytesRead = pos
		}
	}
}
