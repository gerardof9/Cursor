package events

// Operation classifies a user-data change event.
type Operation string

const (
	OpInsert  Operation = "INSERT"
	OpUpdate  Operation = "UPDATE"
	OpDelete  Operation = "DELETE"
	OpDDL     Operation = "DDL"
	OpUnknown Operation = "UNKNOWN"
)

func (o Operation) String() string {
	return string(o)
}

// ReplicationFormat indicates row-based or statement-based replication.
type ReplicationFormat string

const (
	FormatRow        ReplicationFormat = "ROW"
	FormatStatement  ReplicationFormat = "STATEMENT"
	FormatUnknown    ReplicationFormat = "UNKNOWN"
)

func (f ReplicationFormat) String() string {
	return string(f)
}
