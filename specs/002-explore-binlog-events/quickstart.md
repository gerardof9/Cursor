# Quickstart: Manual Validation

**Feature**: `002-explore-binlog-events`  
**Date**: 2026-06-10

## Prerequisites

- Go 1.22+ installed
- At least two MySQL binary log samples:
  - **Small**: < 50 MB, known tables and operations
  - **Large**: ≥ 100 MB (ideally approaching 1 GB) for performance checks
- Terminal ≥ 80×24

## Build & Run

```bash
go build -o binlog-explorer ./cmd/binlog-explorer
./binlog-explorer /path/to/mysql-bin.000001
```

Open multiple files at launch:

```bash
./binlog-explorer /path/to/mysql-bin.000001 /path/to/mysql-bin.000002
```

Empty start (in-session open):

```bash
./binlog-explorer
# Press o, enter path, Enter
```

## Validation Scenarios

### VS-1 Open Binlogs (P1)

1. Launch with one valid binlog via CLI → events begin appearing in list.
2. Press `o`, add a second valid binlog → merged chronological list grows.
3. Launch with an invalid path → clear error; app remains usable if other files valid.
4. Launch with no args, open file in-session → same behavior as CLI open.

**Pass**: All four cases behave per [spec.md](./spec.md) User Story 1.

### VS-2 Browse Events (P2)

1. Confirm list shows only INSERT/UPDATE/DELETE/DDL (no Rotate/GTID noise).
2. Verify timestamps ascend; multi-file merge is chronological.
3. Navigate large sample with ↑/↓ and PgUp/PgDn — cursor visible, no multi-second stalls.

**Pass**: User Story 2 acceptance scenarios.

### VS-3 Inspect Details (P3)

1. Select a row-format UPDATE → detail shows table, operation, row values or partial data.
2. Select a statement-format event → detail notes row images unavailable if applicable.
3. Move selection rapidly → detail updates; no crash on incomplete metadata.

**Pass**: User Story 3 acceptance scenarios.

### VS-4 Filter Events (P4)

1. Filter by table name → only matching events shown.
2. Filter by operation type (e.g., DELETE only).
3. Filter by time range including boundary timestamp → boundary events included.
4. Press `c` → full list restored.

**Pass**: User Story 4 acceptance scenarios.

### VS-5 Investigate Activity (P5)

**Scenario**: "What happened to `mydb.orders` between 14:00 and 15:00?"

1. Open binlog covering that window.
2. Apply table filter `orders` (or `mydb.orders`) and time range.
3. Step through results with `j`/`k`; inspect each candidate in detail pane.
4. Answer the question without exiting app or using mysqlbinlog/grep.

**Pass**: Completed in < 5 minutes on familiar sample (SC-002).

### VS-6 Performance Spot Checks

| Check | Target |
|-------|--------|
| First events visible (1 GB file) | ≤ 30 seconds (SC-001) |
| Filter 10k+ index | ≤ 1 minute (SC-003) |
| Memory | Activity Monitor / `top` — no runaway growth during indexing |

## Known Limitations (v1)

- MariaDB binlogs: best-effort; document failures in session notes
- Duplicate/overlapping files: no deduplication; same event may appear twice
- No export, SQL reconstruction, or transaction grouping
- Filter logic is AND-only; no free-text search

## Reporting Issues

Record: binlog MySQL version, file size, operation performed, expected vs actual. Attach anonymized screenshot of status bar + list headers if UI issue.
