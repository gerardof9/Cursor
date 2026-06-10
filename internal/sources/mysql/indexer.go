package mysql

import (
	"fmt"
	"io"
	"strings"
	"time"

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

// IndexStream parses user-data events from the current file position.
func IndexStream(src *Source, emit func(events.EventSummary)) error {
	src.State = StateIndexing
	parser := src.parser
	f := src.file

	for {
		offset, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			src.MarkError(err)
			return err
		}

		ev, err := parser.Parse(f)
		if err == io.EOF {
			src.MarkReady()
			return nil
		}
		if err != nil {
			src.MarkError(err)
			return fmt.Errorf("parse at offset %d: %w", offset, err)
		}

		if pos, err := f.Seek(0, io.SeekCurrent); err == nil {
			src.BytesRead = pos
		}

		summary, ok := mapEvent(ev, src, offset)
		if ok {
			emit(summary)
			src.IndexedCount++
		}
	}
}

func mapEvent(ev *replication.BinlogEvent, src *Source, offset int64) (events.EventSummary, bool) {
	header := ev.Header
	ts := time.Unix(int64(header.Timestamp), 0)

	switch e := ev.Event.(type) {
	case *replication.RowsEvent:
		op := rowsOperation(e)
		if op == events.OpUnknown {
			return events.EventSummary{}, false
		}
		schemaName, tableName := tableNames(e.Table)
		return events.EventSummary{
			ID:         0,
			SourceID:   src.ID,
			Offset:     offset,
			Timestamp:  ts,
			Operation:  op,
			Schema:     schemaName,
			Table:      tableName,
			Format:     events.FormatRow,
			SourcePath: src.Path,
		}, true

	case *replication.QueryEvent:
		query := strings.TrimSpace(string(e.Query))
		upper := strings.ToUpper(query)
		if isHousekeepingQuery(upper) {
			return events.EventSummary{}, false
		}
		op := queryOperation(upper)
		if op == events.OpUnknown {
			return events.EventSummary{}, false
		}
		schemaName := string(e.Schema)
		return events.EventSummary{
			ID:         0,
			SourceID:   src.ID,
			Offset:     offset,
			Timestamp:  ts,
			Operation:  op,
			Schema:     schemaName,
			Format:     events.FormatStatement,
			SourcePath: src.Path,
		}, true
	}

	return events.EventSummary{}, false
}

func rowsOperation(e *replication.RowsEvent) events.Operation {
	switch e.Type() {
	case replication.EnumRowsEventTypeInsert:
		return events.OpInsert
	case replication.EnumRowsEventTypeUpdate:
		return events.OpUpdate
	case replication.EnumRowsEventTypeDelete:
		return events.OpDelete
	default:
		return events.OpUnknown
	}
}

func tableNames(t *replication.TableMapEvent) (string, string) {
	if t == nil {
		return "", ""
	}
	return string(t.Schema), string(t.Table)
}

func isHousekeepingQuery(upper string) bool {
	switch {
	case upper == "BEGIN", upper == "COMMIT", upper == "ROLLBACK":
		return true
	case strings.HasPrefix(upper, "SET "), strings.HasPrefix(upper, "USE "):
		return true
	default:
		return false
	}
}

func queryOperation(upper string) events.Operation {
	switch {
	case strings.HasPrefix(upper, "INSERT"):
		return events.OpInsert
	case strings.HasPrefix(upper, "UPDATE"):
		return events.OpUpdate
	case strings.HasPrefix(upper, "DELETE"):
		return events.OpDelete
	case strings.HasPrefix(upper, "CREATE"),
		strings.HasPrefix(upper, "ALTER"),
		strings.HasPrefix(upper, "DROP"),
		strings.HasPrefix(upper, "TRUNCATE"),
		strings.HasPrefix(upper, "RENAME"):
		return events.OpDDL
	default:
		return events.OpDDL
	}
}
