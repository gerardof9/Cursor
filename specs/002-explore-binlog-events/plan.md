# Implementation Plan: Open and Explore Database Change Events

**Branch**: `002-explore-binlog-events` | **Date**: 2026-06-10 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/002-explore-binlog-events/spec.md` plus vertical-slice implementation guidance.

## Summary

Deliver the first usable MySQL Binlog Explorer as a complete Go TUI vertical slice: open one or more local binlog files (CLI and in-session), stream-parse into a lightweight event index, browse a chronological change-event list, inspect full detail on selection, and filter by time range, schema/table, and operation type. The design favors concrete packages over abstractions, manual validation with real binlogs, and responsiveness on large files without premature optimization.

## Technical Context

**Language/Version**: Go 1.22+

**Primary Dependencies**:
- `github.com/charmbracelet/bubbletea` — TUI application loop
- `github.com/charmbracelet/bubbles` — list, viewport, textinput, help
- `github.com/charmbracelet/lipgloss` — styling
- `github.com/go-mysql-org/go-mysql/replication` — MySQL binlog event parsing (row + statement)

**Storage**: Local MySQL `.binlog` / relay log files; in-memory event index (`[]EventSummary`); on-demand detail reads via stored file offsets

**Testing**: Manual validation only per constitution (see [quickstart.md](./quickstart.md))

**Target Platform**: Terminal on Linux, macOS, Windows; single `binlog-explorer` executable

**Project Type**: TUI desktop application (Go native)

**Performance Goals**:
- First indexed events visible within 30 seconds for 1 GB binlogs (SC-001)
- List navigation feels immediate on 10k+ indexed events (keyboard response < 100ms perceived)
- Filter application on indexed metadata < 1 second for 10k events (SC-003)

**Constraints**:
- Stream-parse; no full in-memory payload retention (FR-018/019)
- Housekeeping binlog events excluded from browsable list (FR-003a)
- Inclusive time-range filter boundaries (clarification)
- No automated test infrastructure

**Scale/Scope**:
- 1–10 open binlog files per session (practical DBA use)
- Binlogs up to several GB; millions of indexed change events
- MySQL 5.7 and 8.0 row/statement formats (see [research.md](./research.md))

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Reference: `.specify/memory/constitution.md`

- [x] **DBA First**: Keyboard-driven split-pane explorer; dense list columns (time, op, schema, table)
- [x] **Go Native**: Go + Bubble Tea stack only; single binary distribution
- [x] **TUI First**: Full workflow in TUI from iteration one; CLI args only for initial file paths
- [x] **Explorer, Not Viewer**: Structured summaries, filters, and detail panes—not raw mysqlbinlog dump
- [x] **Incremental Development**: Single vertical slice covering open → index → browse → inspect → filter
- [x] **Simplicity Over Engineering**: Concrete `sources/mysql` package; no plugins, DI, or service layers
- [x] **Performance Awareness**: Stream index + lazy detail; progressive indexing in background goroutine
- [x] **Pragmatic Extensibility**: Generic naming (`events`, `explorer`, `sources/mysql`); one source type only
- [x] **Spec-Driven Development**: Spec clarified; plan derived from spec + user implementation notes
- [x] **No Automated Testing**: Manual quickstart validation scenarios only

**Post-design re-check**: All gates pass. No Complexity Tracking entries required.

## Project Structure

### Documentation (this feature)

```text
specs/002-explore-binlog-events/
├── plan.md              # This file
├── research.md          # Phase 0 decisions
├── data-model.md        # Phase 1 entities and state
├── quickstart.md        # Manual validation guide
├── contracts/           # CLI, UI layout, keybindings
└── tasks.md             # Phase 2 (/speckit-tasks — not yet created)
```

### Source Code (repository root)

```text
cmd/
└── binlog-explorer/
    └── main.go                 # Entry: parse CLI args, run Bubble Tea program

internal/
├── events/                     # Domain types: Operation, EventSummary, EventDetail, SourceRef
├── explorer/
│   └── session.go              # Session: open sources, merged index, selection, filter state
├── filters/
│   └── filter.go               # FilterCriteria + apply over index
├── sources/
│   └── mysql/
│       ├── source.go           # Binlog file open, format detection
│       ├── indexer.go          # Stream-parse → append EventSummary + file offset
│       └── detail.go           # Load EventDetail on demand from offset
└── ui/
    ├── app.go                  # Root tea.Model; coordinates sub-views
    ├── list.go                 # Event list (bubbles/list)
    ├── detail.go               # Detail viewport
    ├── filterbar.go            # Filter input / active filter display
    ├── openfile.go             # In-session path input for additional binlogs
    └── keys.go                 # Key map and help text

go.mod
go.sum
```

**Structure Decision**: Five internal packages mirror the natural workflow boundaries (types, session orchestration, filtering, MySQL I/O, TUI) without interface hierarchies or provider frameworks. `explorer.Session` is the single in-memory coordinator passed into the UI model. All MySQL-specific logic stays under `sources/mysql`.

## Implementation Phases (Vertical Slice)

### Phase A — Bootstrap & Open (P1)

1. Initialize `go.mod`, Bubble Tea shell, empty split layout.
2. Accept CLI file paths; validate and register each as a `BinlogSource`.
3. In-session open via keybinding (`o`) → path text input → append source.
4. Surface load/index errors in status area without crashing.

### Phase B — Index & Browse (P2)

1. Background indexer goroutine per source: stream-parse with `go-mysql` replication parser.
2. Map parsed events to `EventSummary`; skip housekeeping events.
3. Merge summaries from all sources; sort by timestamp (stable tie-break: source path + offset).
4. Render sortable list with columns: time, operation, schema, table, source file (abbreviated).
5. Keyboard navigation: ↑/↓, PgUp/PgDn, Home/End; cursor always visible.
6. On indexer mid-read failure: mark source `error`, show status message; other sources remain usable.
7. On housekeeping-only or empty-of-changes binlog: show empty-index message; keep source registered.

### Phase C — Inspect Detail (P3)

1. On selection, call `detail.Load(summary)` using stored file offset.
2. Detail pane: timestamp, operation, schema/table, format (row/statement), transaction id if present, row data or SQL text, completeness flags.
3. Indicate missing row images for statement-based events (FR-017).

### Phase D — Filter (P4)

1. Filter dialog or cycle keybindings for operation, schema/table, time range.
2. Apply filters over in-memory index (AND combination).
3. Inclusive time boundaries; clear-all binding (`c` or `C`).
4. Empty-result message when no matches.

### Phase E — Investigation polish (P5)

1. Filtered navigation preserves detail pane behavior.
2. On filter change: move selection to nearest match or clear with status message.
3. Status bar: event count (filtered/total), active filters, indexing progress %.

## Complexity Tracking

> No constitution violations. Table intentionally empty.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |
