# Data Model: Scoped Binlog Investigation with Pre-Open Analysis

**Date**: 2026-06-10  
**Feature**: `003-scoped-binlog-investigation`

## Overview

Extends feature 002 in-memory session model with **file analysis results**, **investigation scope**, and a **pre-index gate**. Secondary filters no longer include time range. All structures remain in memory; no persistence.

## New / Extended Entities

### FileAnalysisResult (`events` or `sources/mysql`)

| Field | Type | Description |
|-------|------|-------------|
| `SourceID` | `string` | Reference to BinlogSource |
| `FileSize` | `int64` | Bytes from stat |
| `FileSizeHuman` | `string` | Display e.g. `8.2 GB` |
| `MinTimestamp` | `time.Time` | Earliest event timestamp seen (zero if none) |
| `MaxTimestamp` | `time.Time` | Latest event timestamp seen (zero if none) |
| `ApproxChangeCount` | `int64` | Approximate user-data events (`~` in UI) |
| `Complete` | `bool` | Analysis finished successfully |
| `Error` | `string` | Error message if analysis failed |

**Validation**: Populated only after successful analysis pass; cached on `Source` until close.

### InvestigationScope (`events`)

| Field | Type | Description |
|-------|------|-------------|
| `From` | `time.Time` | Inclusive start |
| `To` | `time.Time` | Inclusive end |
| `Preset` | `ScopePreset` | `Entire`, `LastHour`, `LastDay`, `Custom` |

```text
ScopePreset: Entire | LastHour | LastDay | Custom
```

**Rules**:
- `From <= To`
- Custom range must fall within analyzed `[MinTimestamp, MaxTimestamp]` per source (session uses merged bounds)
- Presets computed from session merged `MaxTimestamp`

### SessionAnalysisSummary (`explorer`)

Aggregated view for scope dialog:

| Field | Type | Description |
|-------|------|-------------|
| `TotalSize` | `int64` | Sum of source file sizes |
| `MergedMin` | `time.Time` | Min of source mins |
| `MergedMax` | `time.Time` | Max of source maxes |
| `ApproxChangeCount` | `int64` | Sum of approximate counts |
| `SourceCount` | `int` | Number of sources analyzed |

### BinlogSource (`sources/mysql`) — extensions

| Field | Type | Description |
|-------|------|-------------|
| `Analysis` | `*FileAnalysisResult` | Cached analysis; nil until complete |
| `ActiveScope` | `*InvestigationScope` | Scope used for last/current index |

**SourceState** extended:

```text
opening → analyzing → indexing → ready
opening → error
analyzing → error
indexing → error
```

Remove direct `opening → indexing` path from 002; analysis always intervenes.

### FilterCriteria (`filters`) — reduced

| Field | Type | Description |
|-------|------|-------------|
| `Operations` | `[]Operation` | Empty = all |
| `Schema` | `string` | Empty = all |
| `Table` | `string` | Empty = all |

**Removed**: `TimeStart`, `TimeEnd` (replaced by `InvestigationScope`).

### ExplorerSession (`explorer`) — extensions

| Field | Type | Description |
|-------|------|-------------|
| `InvestigationScope` | `*InvestigationScope` | Active scope; nil until confirmed |
| `AwaitingScope` | `bool` | True after analysis, before scope dialog confirm |
| `LaunchScope` | `*InvestigationScope` | Optional scope from CLI `--from`/`--to` |
| `AnalysisSummary` | `SessionAnalysisSummary` | Merged analysis for dialog |

**Existing fields** (`Index`, `Filtered`, `Filter`, etc.) unchanged in role; `Index` contains only scoped events.

## State Transitions

### Open workflow (single source)

```text
OpenSource → analyzing (AnalyzeStream)
  → analysis complete → AwaitingScope=true → scope dialog
  → scope confirmed → indexing (IndexStream with scope)
  → ready
```

### Open workflow (CLI with --from/--to)

```text
OpenSource → analyzing
  → analysis complete → apply LaunchScope → indexing (skip dialog)
  → ready
```

### Scope change in session (key `s`)

```text
scope dialog → new scope
  if widen or entire → clear Index → re-index
  if narrow ⊆ current → filter Index in memory (optional) OR re-index
  → reset Selected, Detail, secondary Filter
```

### Entire file (large) confirmation

```text
select Entire → if over threshold → warning modal
  Y → indexing (prefer single-pass with analysis)
  F → return to scope dialog
```

## Relationships

```text
ExplorerSession 1──1 InvestigationScope (active, when set)
ExplorerSession 1──* BinlogSource
BinlogSource  1──1 FileAnalysisResult (cached)
BinlogSource  1──1 InvestigationScope (last applied)
InvestigationScope governs → IndexStream emission
FilterCriteria (secondary) applied to → Index → Filtered
```

## Invariants

- Browse `Index` populated only after `InvestigationScope` confirmed
- All events in `Index` satisfy `From <= Timestamp <= To` of active scope
- `Analysis` cached per source; not recomputed on scope change unless source reopened
- Secondary filters never trigger binlog re-parse
- Housekeeping events never in `Index` (unchanged from 002)
- Approximate counts in analysis UI prefixed with `~`; post-index count is exact for scoped set

## Constants (implementation)

| Constant | Value | Purpose |
|----------|-------|---------|
| `LargeFileSizeBytes` | 1 GiB | Entire-file warning |
| `LargeFileEventCount` | 500,000 | Entire-file warning |

Defined in plan/research; implemented in `sources/mysql` package.
