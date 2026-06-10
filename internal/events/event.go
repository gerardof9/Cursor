package events

import "time"

// EventSummary is a lightweight index entry for browse and filter.
type EventSummary struct {
	ID        uint64
	SourceID  string
	Offset    int64
	Timestamp time.Time
	Operation Operation
	Schema    string
	Table     string
	Format    ReplicationFormat
	TxHint    string
	SourcePath string
}

// RowChange holds decoded row images when available.
type RowChange struct {
	Before []string
	After  []string
}

// EventDetail is the expanded on-demand view of a selected event.
type EventDetail struct {
	Summary   EventSummary
	SQL       string
	RowValues []RowChange
	Complete  bool
	Notes     []string
}
