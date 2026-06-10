package mysql

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-mysql-org/go-mysql/replication"

	"db-log-explorer/internal/events"
)

// LoadDetail re-parses a single event at the stored file offset.
func LoadDetail(src *Source, summary events.EventSummary) (events.EventDetail, error) {
	if src == nil || src.file == nil {
		return events.EventDetail{}, fmt.Errorf("source not open")
	}

	if _, err := src.file.Seek(summary.Offset, io.SeekStart); err != nil {
		return events.EventDetail{}, fmt.Errorf("seek: %w", err)
	}

	ev, err := src.parser.Parse(src.file)
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
