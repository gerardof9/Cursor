# Data Model: Open and Explore Database Change Events

**Date**: 2026-06-10  
**Feature**: `002-explore-binlog-events`

## Overview

All structures live in memory for the duration of a session. No persistence layer. The explorer session is the aggregate root.

## Entities

### BinlogSource (`sources/mysql`)

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `string` | Stable session identifier (e.g., short hash of path) |
| `Path` | `string` | Absolute filesystem path |
| `Format` | `BinlogFormat` | Row / statement / mixed (detected while parsing) |
| `State` | `SourceState` | `opening`, `indexing`, `ready`, `error` |
| `Error` | `string` | Last error message if `State == error` |
| `IndexedCount` | `int` | Summaries produced so far |
| `File` | `*os.File` | Open file handle for detail reload |

**Validation**:
- Path must exist and be readable
- File must parse a valid `FormatDescriptionEvent` on open

### Operation (`events`)

```text
INSERT | UPDATE | DELETE | DDL | Unknown
```

Mapped from `RowsEvent` action or `QueryEvent` SQL classification.

### ReplicationFormat (`events`)

```text
Row | Statement | Unknown
```

### EventSummary (`events`) — index entry

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `uint64` | Monotonic session-wide index id |
| `SourceID` | `string` | Reference to BinlogSource |
| `Offset` | `uint64` | Byte position in binlog file for lazy detail load |
| `Timestamp` | `time.Time` | Event header timestamp |
| `Operation` | `Operation` | INSERT, UPDATE, DELETE, DDL |
| `Schema` | `string` | Database name if known |
| `Table` | `string` | Table name if known |
| `Format` | `ReplicationFormat` | Row or statement |
| `TxHint` | `string` | Optional transaction identifier when available |

**List display columns**: Timestamp, Operation, Schema, Table, Source (basename).

### EventDetail (`events`) — loaded on selection

| Field | Type | Description |
|-------|------|-------------|
| `Summary` | `EventSummary` | Embedded summary fields |
| `SQL` | `string` | Statement text for QueryEvent / DDL |
| `RowValues` | `[]RowChange` | Decoded row images when row format |
| `Complete` | `bool` | All parseable fields extracted |
| `Notes` | `[]string` | e.g., "Row images unavailable (statement format)" |

### RowChange (`events`)

| Field | Type | Description |
|-------|------|-------------|
| `Before` | `[]string` | Column values before change (updates/deletes) |
| `After` | `[]string` | Column values after change (inserts/updates) |

### FilterCriteria (`filters`)

| Field | Type | Description |
|-------|------|-------------|
| `Operations` | `[]Operation` | Empty = all operations |
| `Schema` | `string` | Empty = all schemas |
| `Table` | `string` | Empty = all tables; match against `schema.table` |
| `TimeStart` | `*time.Time` | Nil = unbounded start |
| `TimeEnd` | `*time.Time` | Nil = unbounded end; inclusive boundaries |

**Rules**:
- All non-empty criteria combined with AND
- Table match: case-insensitive; supports `schema.table` or table name alone
- Time: `TimeStart <= event.Timestamp <= TimeEnd` when bounds set

### ExplorerSession (`explorer`)

| Field | Type | Description |
|-------|------|-------------|
| `Sources` | `[]*BinlogSource` | Open binlog files |
| `Index` | `[]EventSummary` | Merged chronological index (all sources) |
| `Filtered` | `[]int` | Indices into `Index` passing current filter |
| `Filter` | `FilterCriteria` | Active filter |
| `SelectedIndex` | `int` | Position in `Filtered` (-1 = none) |
| `LoadedDetail` | `*EventDetail` | Cache for current selection only |

## State Transitions

### BinlogSource

```text
opening → indexing → ready
opening → error
indexing → error
```

### ExplorerSession (selection on filter change)

```text
filter applied → recompute Filtered
  if previous selection event still in Filtered → keep nearest index
  else → SelectedIndex = -1, clear LoadedDetail, status message
```

### Indexing (per source)

```text
goroutine start → parse loop → batch Msg to UI → append to Index → resort merge → ready
```

## Relationships

```text
ExplorerSession 1──* BinlogSource
ExplorerSession 1──* EventSummary (via Index)
EventSummary *──1 BinlogSource (SourceID)
EventDetail 1──1 EventSummary
FilterCriteria applied to → EventSummary slice → produces Filtered indices
```

## Invariants

- `Index` sorted ascending by `(Timestamp, SourceID, Offset)`
- Housekeeping binlog events never appear in `Index`
- At most one `LoadedDetail` cached; invalidated on selection change
- `Filtered` empty with active filter → UI shows empty state message (not an error)
