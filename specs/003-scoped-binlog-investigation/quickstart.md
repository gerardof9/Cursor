# Quickstart: Scoped Binlog Investigation (Manual Validation)

**Feature**: `003-scoped-binlog-investigation`  
**Branch**: `003-scoped-binlog-investigation`

## Prerequisites

- Branch `003-scoped-binlog-investigation`; `go build ./cmd/binlog-explorer` verified (T001)
- Feature 002 baseline working (`binlog-explorer` builds and runs)
- Real MySQL binlog samples:
  - **Small**: <100 MB, known event span
  - **Large**: ≥1 GB or ≥500k change events (for warning path)
- Linux server or local path to binlogs

## Build

```bash
go build -o binlog-explorer ./cmd/binlog-explorer
```

Cross-compile for DB server:

```bash
GOOS=linux GOARCH=amd64 go build -o binlog-explorer ./cmd/binlog-explorer
```

## Validation Scenarios

### VS-1 — Analysis pass (US1)

1. Run `binlog-explorer /path/to/large.binlog`
2. **Expect**: Status shows `Analyzing…` with progress before any list events
3. **Expect**: Scope dialog shows size, time span, `~` approximate event count
4. **Expect**: No browse events until scope confirmed

### VS-2 — Scope dialog presets (US2)

1. Complete VS-1 analysis
2. Press `3` (Last day of file), Enter
3. **Expect**: Scoped indexing begins; status shows scope range
4. **Expect**: List events fall within last 24h of file max timestamp

### VS-3 — Large-file warning (US2, SC-004)

1. Open binlog ≥1 GiB (or dense file ≥500k events)
2. Select `1` Entire file
3. **Expect**: Secondary warning with size, span, approximate count
4. Press `F` → **Expect**: Return to scope dialog
5. Select custom narrow range → **Expect**: Indexing without full-file warning

### VS-4 — Scoped indexing performance (US3, SC-001)

1. Open large binlog; choose 1-day custom range covering <5% of span
2. Note time to first list events and total index completion
3. Repeat with Entire file (if feasible) or compare against 002 branch
4. **Expect**: Scoped run materially faster; first events within ~30s on 1 GB sample

### VS-5 — CLI scope (US5)

```bash
binlog-explorer --from "2025-06-08" --to "2025-06-09" /path/to/binlog
```

1. **Expect**: Brief analysis; scope dialog skipped
2. **Expect**: Index contains only events in range

**Negative test** (missing half of scope pair — exit at parse):

```bash
binlog-explorer --from "2025-06-08" /path/to/binlog
```

**Expect**: Exit 1 with message requiring both `--from` and `--to`

**Negative test** (out-of-bounds — in TUI after analysis, not exit at parse):

```bash
binlog-explorer --from "2099-01-01" --to "2099-01-02" /path/to/binlog
```

**Expect**: Analysis runs; TUI shows clear error; scoped indexing does not start; app remains usable until quit

### VS-6 — Change scope in session (US4)

1. Complete scoped index for day A
2. Press `s`; select day B (non-overlapping)
3. **Expect**: Prior list replaced; re-index progress shown; selection/detail reset
4. **Expect**: No full re-analysis (cached min/max in dialog)

### VS-7 — Secondary filters (US6, SC-005)

1. With scoped index loaded, press `f`; filter by table name
2. **Expect**: Instant filter; status `shown / scoped-total`
3. **Expect**: No `Analyzing` or `Indexing scope` during filter apply
4. Press `c` → **Expect**: Full scoped list restored

### VS-8 — End-to-end investigation (SC-006)

1. Open binlog with `--from`/`--to` or scope dialog
2. Filter to suspect table
3. Browse and inspect 3+ events
4. **Expect**: Answer "what happened on table X on day D" without external tools in <5 minutes

## Edge Case Checks

| Case | Steps | Expected |
|------|-------|----------|
| Empty / no change events | Open housekeeping-only binlog | Analysis ~0; empty-index message after scope |
| Mid-read failure | Truncate binlog or use corrupt tail | Source error; other sources OK |
| Multi-file | Open 2 binlogs at launch | Merged span in dialog; single scope applies |
| Cancel analysis | Esc during analyze | Open aborted; no index |

## Regression (002 behaviors preserved)

- On-demand detail at selection
- Keyboard list navigation
- Multi-source chronological merge within scope
- Partial CLI open failures (exit 2 with warnings)

## Performance Observation (optional)

During VS-4, note resident memory with narrow scope vs entire file (Activity Monitor / `top`). Scoped run should scale with scoped event count (SC-002).

## Validation Results (T040 / T041)

**Date**: 2026-06-10  
**Environment**: Windows dev host; build verified with `go build ./cmd/binlog-explorer`. Real binlog manual pass **not yet run** on this host.

| Scenario | Status | Notes |
|----------|--------|-------|
| VS-1 Analysis | Pending | Requires real binlog on TTY |
| VS-2 Scope presets | Pending | |
| VS-3 Large-file warning | Pending | Requires ≥1 GiB or dense sample |
| VS-4 Scoped performance | Pending | Compare on Linux server |
| VS-5 CLI scope | Pending | Negative parse tests verifiable without binlog |
| VS-6 Scope change (`s`) | Pending | |
| VS-7 Secondary filters | Pending | |
| VS-8 End-to-end | Pending | |
| T041 Regression | Partial | Build passes; 002 code paths preserved in scoped workflow |

**Next step**: Run VS-1–VS-8 on Linux server with production binlog samples and update this table.
