# Research: Scoped Binlog Investigation with Pre-Open Analysis

**Date**: 2026-06-10  
**Feature**: `003-scoped-binlog-investigation`

## 1. Analysis Pass Implementation

**Decision**: Implement `AnalyzeStream` using the same `ParseSingleEvent` loop as indexing, with a shared `classifyChangeEvent` helper that returns `(isUserData bool, timestamp)` without building `EventSummary` or decoding row payloads.

**Rationale**:
- Reuses proven go-mysql integration from feature 002.
- Guarantees analysis counts match indexer inclusion rules (housekeeping exclusion).
- No second parser or raw header-only decode path to maintain.

**Alternatives considered**:

| Alternative | Rejected because |
|-------------|------------------|
| Raw header-only parse skipping event body | Fragile across event types; go-mysql still reads full event for checksum |
| Store analysis results on disk | Out of scope; adds I/O complexity |
| Skip analysis when CLI scope provided | Still need min/max for validation and status display; run shortened analysis or validate scope against quick min scan |

## 2. Analysis vs Indexing Cost

**Decision**: Analysis reads the full file sequentially (required for max timestamp and approximate count) but avoids `EventSummary` allocation, index merge/sort, and row decode. UI copy: **"Analyzing file…"** not "instant."

**Rationale**: Binlog files do not expose max timestamp without scanning. Analysis is still materially cheaper than scoped indexing when the chosen scope is a small fraction of the file (indexing stops early; analysis does not).

**Alternatives considered**: Stop analysis early — rejected; cannot know max timestamp or offer Last hour/day presets without full scan.

## 3. Single-Pass Entire File (FR-019)

**Decision**: Open workflow uses **at most two full file reads** per source for the Entire-file path:

1. **Pass 1 — Analysis (US1)**: Metadata only (size from `Stat`, min/max ts, `~count`); required before scope dialog so presets (Last hour/day/custom) have bounds.
2. **Pass 2 — Scoped index (US3)**: When user confirms **Entire file**, run `IndexStream` with scope `[MergedMin, MergedMax]` from cached analysis, **no early stop**, **no third pass**.

**FR-019 interpretation**: Avoid a **redundant third read** (analysis → separate analyze+index combo → index). Entire file after dialog = analysis + one index pass only (two reads total). Do **not** implement a separate combined pass after analysis completes.

**Rationale**: Pass 1 is required for informed scope choice on all dialog paths. Entire-file optimization is **one index pass** reusing cached bounds, not merging analysis and index into a single physical read while the dialog still needs prior metadata.

**Alternatives considered**:
- True single-read Entire (defer all metadata to index pass) — rejected; scope dialog cannot show min/max/`~count` without pass 1.
- Always three passes — rejected.
- Resume indexer from analysis file position without rewind — parser state does not carry over cleanly.

## 4. Large-File Warning Thresholds

**Decision**:
- **Size threshold**: ≥ **1 GiB** (1073741824 bytes) per source or aggregate session size
- **Event count threshold**: ≥ **500,000** approximate change events (session aggregate)

Either threshold triggers secondary confirmation for Entire file.

**Rationale**: Aligns with SC-001 (1 GB reference binlog) and practical DBA expectations; constants in `internal/sources/mysql/limits.go` (or top of `analyzer.go`).

**Alternatives considered**: Size-only — rejected; dense small files can still have high event counts.

## 5. Investigation Scope Presets

**Decision**:
- **Last hour (of file)**: `From = maxTimestamp - 1h`, `To = maxTimestamp`, inclusive
- **Last day (of file)**: `From = maxTimestamp - 24h`, `To = maxTimestamp`, inclusive
- **Custom**: DBA-entered; validated ⊆ `[minTimestamp, maxTimestamp]`
- **Entire file**: `From = minTimestamp`, `To = maxTimestamp` (or unbounded parse with no early stop)

**Rationale**: Spec clarification: presets relative to file bounds, not wall clock.

## 6. Early Stop After `To`

**Decision**: On scoped index, when a user-data or skipped event header has `timestamp > To`, stop parsing that source. If a later event would have `timestamp <= To` (non-monotonic), those events are missed—document in quickstart.

**Rationale**: Best-effort matches spec; MySQL binlog v4 timestamps are mostly monotonic per file.

**Alternatives considered**: Continue parsing after first `> To` — defeats performance goal.

## 7. Scope Change Keybinding

**Decision**: **`s`** opens investigation scope dialog (global, when not in another modal).

**Rationale**: Short, mnemonic (scope), unused in 002 key map; keeps `f` for secondary filters only.

**Alternatives considered**: Extend filter modal — rejected; spec requires separate flows for scope vs table/op filter.

## 8. Narrowing Scope Without Re-Parse

**Decision**: v1 implements in-memory narrow when `newFrom >= currentFrom && newTo <= currentTo` and current index was built with scoped indexing (not Entire file with possible gaps). Otherwise re-index.

**Rationale**: Spec allows re-index always; optimization reduces wait on common "refine to smaller window" path.

## 9. Secondary Filter Time Fields

**Decision**: Remove `TimeStart`/`TimeEnd` from `filters.Criteria` and filter editor UI. Time window is exclusively `InvestigationScope`.

**Rationale**: Avoids duplicate/confusing time controls; scope is load criterion per spec US6.

**Migration note**: Existing 002 filter tests N/A (no automated tests); manual validation updates in quickstart.

## 10. CLI Timestamp Format

**Decision**: Accept `2006-01-02 15:04:05` and `2006-01-02` (start of day for `--from`, end of day 23:59:59 for `--to` when date-only).

**Rationale**: Matches filter editor placeholder from 002; date-only is common for DBA "investigate this day" workflow.

## 11. Multi-Source Scope

**Decision**: One `InvestigationScope` applies to all open sources. Analysis shows per-source and aggregate rows in scope dialog.

**Rationale**: Spec edge case; matches 002 merged chronological list model.

## 12. Source State Machine Extension

**Decision**: Add states: `analyzing` → (scope pending) → `indexing` → `ready`. Scope pending is a session-level flag (`AwaitingScope`) rather than source state when analysis done but dialog not confirmed.

**Rationale**: Multiple sources may analyze at different rates; session gates indexing until scope confirmed for all pending sources.
