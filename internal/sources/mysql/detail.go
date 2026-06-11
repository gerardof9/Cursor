package mysql

import (
	"fmt"
	"io"

	"github.com/go-mysql-org/go-mysql/replication"

	"db-log-explorer/internal/events"
)

// LoadDetail re-parses a single event at the stored file offset.
func LoadDetail(src *Source, summary events.EventSummary) (events.EventDetail, error) {
	if src == nil || src.file == nil {
		return events.EventDetail{}, fmt.Errorf("source not open")
	}

	ev, err := parseEventAtOffset(src, summary.Offset)
	if err != nil {
		return events.EventDetail{}, fmt.Errorf("parse detail: %w", err)
	}

	detail := events.EventDetail{
		Summary: summary,
	}

	switch e := ev.Event.(type) {
	case *replication.RowsEvent:
		detail.Complete = true
		switch e.Type() {
		case replication.EnumRowsEventTypeUpdate:
			for i := 0; i+1 < len(e.Rows); i += 2 {
				detail.RowValues = append(detail.RowValues, events.RowChange{
					Before: formatRow(e.Rows[i]),
					After:  formatRow(e.Rows[i+1]),
				})
			}
		case replication.EnumRowsEventTypeDelete:
			for _, row := range e.Rows {
				detail.RowValues = append(detail.RowValues, events.RowChange{
					Before: formatRow(row),
				})
			}
		default:
			for _, row := range e.Rows {
				detail.RowValues = append(detail.RowValues, events.RowChange{
					After: formatRow(row),
				})
			}
		}

	case *replication.QueryEvent:
		detail.SQL = string(e.Query)
		detail.Complete = true
		if summary.Format == events.FormatStatement {
			detail.Notes = append(detail.Notes, "Row images unavailable (statement format)")
		}
	default:
		detail.Notes = append(detail.Notes, "Limited detail for this event type")
	}

	if len(detail.Notes) == 0 && !detail.Complete {
		detail.Notes = append(detail.Notes, "Partial metadata only")
	}

	return detail, nil
}

// parseEventAtOffset replays the binlog from the format description through
// targetOffset so table-map state is available for row events.
func parseEventAtOffset(src *Source, targetOffset int64) (*replication.BinlogEvent, error) {
	parser := replication.NewBinlogParser()
	parser.SetVerifyChecksum(true)

	f := src.file
	if _, err := f.Seek(4, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek: %w", err)
	}

	var found *replication.BinlogEvent
	for {
		offset, err := f.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, fmt.Errorf("tell: %w", err)
		}
		if offset > targetOffset {
			return nil, fmt.Errorf("event not found at offset %d", targetOffset)
		}

		done, err := parser.ParseSingleEvent(f, func(ev *replication.BinlogEvent) error {
			if offset == targetOffset {
				found = ev
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		if found != nil {
			return found, nil
		}
		if done {
			break
		}
	}

	return nil, fmt.Errorf("event not found at offset %d", targetOffset)
}

func formatRow(row []interface{}) []string {
	out := make([]string, len(row))
	for i, v := range row {
		out[i] = formatValue(v)
	}
	return out
}

func formatValue(v interface{}) string {
	if v == nil {
		return "NULL"
	}
	switch val := v.(type) {
	case []byte:
		return string(val)
	default:
		return fmt.Sprintf("%v", v)
	}
}
