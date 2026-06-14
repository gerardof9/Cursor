# Implementation Plan: Scoped Binlog Investigation with Pre-Open Analysis

**Branch**: `003-scoped-binlog-investigation` | **Date**: 2026-06-10 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/003-scoped-binlog-investigation/spec.md`, extending feature 002.

## Summary

Extend the MySQL Binlog Explorer with a three-phase open workflow: **file analysis** (lightweight sequential scan for size, min/max timestamps, approximate change-event count), **investigation scope selection** (mandatory dialog or CLI `--from`/`--to`), and **scoped indexing** (parse-time time window with early stop after `To`). Preserve feature 002 browse, on-demand detail, and secondary filters (schema, table, operation) over the scoped in-memory index. Remove post-index date filtering from the filter editor—time range becomes a load criterion, not a view filter.

## Technical Context

**Language/Version**: Go 1.22+

**Primary Dependencies** (unchanged from 002):
- `github.com/charmbracelet/bubbletea` — TUI application loop
- `github.com/charmbracelet/bubbles` — textinput, help
- `github.com/charmbracelet/lipgloss` — styling
- `github.com/go-mysql-org/go-mysql/replication` — binlog parsing (`ParseSingleEvent`)

**Storage**: Local MySQL binlog files; in-memory analysis cache and scoped `[]EventSummary` index; on-demand detail via file offset

**Testing**: Manual validation only per constitution ([quickstart.md](./quickstart.md))

**Target Platform**: Terminal on Linux, macOS, Windows; cross-compile for Linux server deployment via SSH

**Project Type**: TUI application (Go native)

**Performance Goals**:
- SC-001: First scoped index events within 30 seconds for 1 GB binlog with 1-day scope (<5% of time span)
- SC-001: Scoped indexing ≥50% faster than full-file index on representative 1 GB sample
- SC-005: Secondary filter apply <1 second for 10k scoped events
- Analysis pass materially faster than full scoped index (no summary accumulation, no row decode)

**Constraints**:
- Analysis and scoped index are sequential file reads (binlog is not random-access by timestamp)
- Early stop after `To` is best-effort (mostly monotonic timestamps)
- Inclusive scope boundaries `[From, To]`
- No automated test infrastructure
- No persistent disk index

**Scale/Scope**:
- 1–10 open binlog files per session
- Binlogs up to several GB
- Large-file warning threshold: **≥1 GiB** or **≥500,000** approximate change events (see [research.md](./research.md))

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Reference: `.specify/memory/constitution.md`

- [x] **DBA First**: Scope dialog prevents blind multi-GB indexing; CLI `--from`/`--to` for SSH workflow; keyboard-driven scope change (`s`)
- [x] **Go Native**: Extends existing Go packages; single binary
- [x] **TUI First**: Analysis progress, scope dialog, and warning overlays in TUI—not batch CLI
- [x] **Explorer, Not Viewer**: Scoped index + secondary filters; same investigation semantics as 002
- [x] **Incremental Development**: Vertical slices: analysis → scope dialog → scoped index → CLI → scope change → filter split
- [x] **Simplicity Over Engineering**: Reuse `mapEvent` classification; one new `analyzer.go`; scope struct passed to indexer; no SQLite sidecar
- [x] **Performance Awareness**: Early stop, analysis without summaries, optional single-pass for Entire file
- [x] **Pragmatic Extensibility**: Changes confined to `sources/mysql`, `explorer`, `ui`; no plugin system
- [x] **Spec-Driven Development**: Spec 003 clarified; plan derived from spec
- [x] **No Automated Testing**: Manual quickstart scenarios only

**Post-design re-check**: All gates pass. No Complexity Tracking entries required.

## Project Structure

### Documentation (this feature)

```text
specs/003-scoped-binlog-investigation/
├── plan.md              # This file
├── research.md          # Phase 0 decisions
├── data-model.md        # Phase 1 entities and state
├── quickstart.md        # Manual validation guide
├── contracts/           # CLI, UI, scope dialog, keybindings
└── tasks.md             # Phase 2 (/speckit-tasks — not yet created)
```

### Source Code (repository root — additions marked +)

```text
cmd/
└── binlog-explorer/
    └── main.go                 # + --from, --to flags; pass launch scope to session

internal/
├── events/
│   └── scope.go                # + InvestigationScope, ScopePreset, FileAnalysisSummary
├── explorer/
│   └── session.go              # + analysis cache, active scope, scope-change orchestration
├── filters/
│   └── filter.go               # Remove TimeStart/TimeEnd (time = scope, not filter)
├── sources/
│   └── mysql/
│       ├── source.go           # + SourceState: analyzing; analysis cache on Source
│       ├── analyzer.go         # + AnalyzeStream: min/max ts, ~count, progress
│       ├── indexer.go          # + scoped IndexStream(scope); early stop after To
│       └── detail.go           # unchanged semantics
└── ui/
    ├── app.go                  # + analysis/scope/index phase coordination
    ├── scopedialog.go          # + scope selection + large-file warning modal
    ├── openfile.go             # triggers analysis → scope flow after open
    ├── filterbar.go            # remove time fields; schema/table/op only
    ├── keys.go                 # + s = change scope
    └── ...                     # list, detail unchanged
```

**Structure Decision**: Extend feature 002 layout with minimal new surface: `analyzer.go`, `scope.go`, `scopedialog.go`. Session remains aggregate root. Scope and analysis metadata live on `mysql.Source` and `explorer.Session`. No new top-level packages.

## Implementation Phases (Vertical Slice)

### Phase A — File Analysis Pass (US1, P1)

1. Add `SourceStateAnalyzing` and `AnalyzeStream` in `sources/mysql/analyzer.go`.
2. Reuse event classification from `indexer.go` (`mapEvent` or shared `classifyEvent`) to count user-data events without building `EventSummary`.
3. Track `minTimestamp`, `maxTimestamp`, `approxChangeCount`, `bytesRead` during scan.
4. Background goroutine per source; `AnalysisBatchMsg` / `AnalysisDoneMsg` to UI (mirror indexer pattern).
5. Show status: `Analyzing: 42% | mysql-bin.000123`.
6. Allow cancel (`Esc` in scope/analysis blocking state) → close source, abort open.
7. Cache `FileAnalysisResult` on `Source` for session reuse (FR-005).

### Phase B — Scope Selection Dialog (US2, P1)

1. Add `ui/scopedialog.go`: modal with options 1–4, custom From/To inputs, aggregate multi-source summary.
2. Block browse list until scope confirmed (FR-010).
3. Large-file secondary warning when Entire file selected and thresholds exceeded (FR-008).
4. Presets: Last hour/day relative to `maxTimestamp` from analysis.
5. Validate custom range ⊆ `[minTimestamp, maxTimestamp]`.

### Phase C — Scoped Indexing (US3, P1)

1. Add `events.InvestigationScope` with `From`, `To`, `Preset` (optional).
2. Extend `IndexStream(src, scope, emit)`:
   - Skip events with `timestamp < From`
   - Emit summaries for `From <= timestamp <= To`
   - Stop source parse when `timestamp > To` (best-effort)
3. Status: `Indexing scope: 2025-06-08..2025-06-09 | 67%`.
4. **Entire-file path (FR-019)**: After analysis cached min/max, Entire scope runs **one** `IndexStream` over `[MergedMin, MergedMax]` with early stop disabled — **second read only**; never a third combined pass after analysis.

### Phase D — CLI Scope Flags (US5, P2)

1. Add `--from` and `--to` to `main.go` (format: `2006-01-02 15:04:05` or `2006-01-02`).
2. Both required together; error if only one provided.
3. When both set: analysis → apply scope → skip dialog → scoped index.

### Phase E — Change Scope In Session (US4, P2)

1. Key `s` opens scope dialog with cached analysis (no re-analyze).
2. On scope apply: clear `Index` for affected sources, reset selection/detail, restart scoped indexer.
3. Narrowing optimization: if new `[From, To]` ⊆ previous scope, filter `Index` in memory; else re-index.
4. Widening or Entire file: always re-parse.

### Phase F — Secondary Filters Split (US6, P3)

1. Remove time range fields from filter editor (`filterbar.go`, `filters.Criteria`).
2. Status bar: `12 / 340 events (filtered)` where 340 is scoped total.
3. Scope label in top bar when active: `Scope: 2025-06-08 .. 2025-06-09`.
4. Verify inspect-on-filtered unchanged from 002.

## Complexity Tracking

> No constitution violations. Table intentionally empty.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |

## Relationship to Feature 002

| 002 behavior | 003 change |
|--------------|------------|
| Index entire file on open | Analysis → scope dialog → scoped index |
| Filter time range in `f` modal | Time range moved to scope dialog / CLI |
| `FilterCriteria.TimeStart/TimeEnd` | Removed; use `InvestigationScope` |
| Progressive background indexing | Unchanged pattern; scoped + early stop |
| On-demand detail | Unchanged |

## Manual Validation Milestones

See [quickstart.md](./quickstart.md) for VS-1 through VS-8 scenarios mapped to user stories.
