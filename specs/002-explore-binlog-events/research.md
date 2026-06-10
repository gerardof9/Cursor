# Research: Open and Explore Database Change Events

**Date**: 2026-06-10  
**Feature**: `002-explore-binlog-events`

## 1. MySQL Binlog Parsing Library

**Decision**: Use `github.com/go-mysql-org/go-mysql/replication` (`BinlogParser` streaming from `os.File`).

**Rationale**:
- Actively maintained fork of `siddontang/go-mysql`; widely used for replication tooling.
- Parses raw binlog files offline without a live MySQL connection.
- Supports row events (`RowsEvent`) and statement events (`QueryEvent`) required by FR-016.
- Exposes event headers (timestamp, log position) needed for index and lazy detail reload.

**Alternatives considered**:
| Alternative | Rejected because |
|-------------|------------------|
| Shell out to `mysqlbinlog` | Violates explorer goal; parsing external text is fragile and slow |
| Custom binlog parser | High complexity; reinventing a well-solved problem |
| Live replication stream (`BinlogSyncer`) | Out of scope; spec requires local file access only |

## 2. Go Toolchain Version

**Decision**: Go 1.22 minimum.

**Rationale**: Stable, available on all target platforms; supports modern stdlib and Charm ecosystem requirements.

**Alternatives considered**: Go 1.21 (unnecessary downgrade); Go 1.23 (no required language features—acceptable but not mandated).

## 3. Event Index Strategy

**Decision**: Append-only `[]events.EventSummary` per session merge, each entry storing `(sourceID, fileOffset, headerTimestamp, operation, schema, table, format)`.

**Rationale**:
- Satisfies FR-018 stream-parse with lightweight metadata.
- Filtering and chronological browse operate on summaries only.
- File offset enables FR-019 on-demand detail re-parse without storing payloads.
- Simple slice + sort merge for multi-file sessions.

**Alternatives considered**:
| Alternative | Rejected because |
|-------------|------------------|
| Full in-memory event store | Violates clarification; poor scaling on GB files |
| SQLite sidecar index | Extra dependency and I/O complexity for v1 |
| Window-only parse (no index) | Cannot filter or jump efficiently across large files |

## 4. Housekeeping Event Filtering

**Decision**: During indexing, emit summaries only for:
- `RowsEvent` → INSERT / UPDATE / DELETE (row format)
- `QueryEvent` → DDL and statement DML when not internal housekeeping (`BEGIN`, `COMMIT`, `ROLLBACK` excluded from list)

Skip: `RotateEvent`, `FormatDescriptionEvent`, `GTIDEvent`, `PreviousGTIDsEvent`, `XIDEvent`, `TableMapEvent` (internal to row decode, not listed), heartbeat/anonymous gtid events.

**Rationale**: Matches clarification Q1—user-data change events only.

## 5. Progressive Indexing UX

**Decision**: Index each source in a background goroutine; send `tea.Msg` updates to append batches to the merged list while UI remains interactive.

**Rationale**: SC-001 requires first events within 30s on 1 GB files—user must browse before full index completes.

**Alternatives considered**: Block UI until full index (poor DBA experience on large files).

## 6. Supported MySQL Versions

**Decision**: Target MySQL 5.7 and 8.0 binlog formats (binlog v4). Best-effort on MariaDB compat binlogs; document limitations in quickstart.

**Rationale**: Covers predominant production versions; `FormatDescriptionEvent` from parser handles per-file format metadata.

## 7. In-Session File Open

**Decision**: `o` key opens a bubbles `textinput` overlay for absolute/relative path entry (no GUI file picker—terminal portable).

**Rationale**: Satisfies FR-002 without cross-platform file-dialog dependencies; DBA-first keyboard workflow.

**Alternatives considered**: `zenity`/OS file dialogs (platform-specific, violates simplicity).

## 8. TUI Layout

**Decision**: Fixed split layout:
- Left (~60% width): scrollable event list
- Right (~40% width): detail viewport
- Bottom strip: status (counts, filters, errors, index progress)
- Top strip: optional filter summary / help toggle (`?`)

**Rationale**: Wireshark/LazyGit-style persistent context; list + detail visible simultaneously (investigation workflow P5).

## 9. Filter Implementation

**Decision**: In-memory AND filter over index:
- `OperationFilter`: set of INSERT/UPDATE/DELETE/DDL
- `TableFilter`: case-insensitive schema.table substring or exact match
- `TimeFilter`: inclusive `[start, end]` on event header timestamp

**Rationale**: Matches spec FR-008–010 and clarification Q5; no query language needed for v1.

**Deferred**: Free-text search, OR logic, saved filter presets.

## 10. Duplicate / Overlapping Binlog Files

**Decision**: Allow opening duplicates; merged list may contain duplicate events. Show full source path in list column so DBA can distinguish. No deduplication in v1.

**Rationale**: Low-impact edge case from spec; dedup requires GTID/position logic out of scope.
