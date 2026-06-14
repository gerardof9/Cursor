# Tasks: Scoped Binlog Investigation with Pre-Open Analysis

**Input**: Design documents from `specs/003-scoped-binlog-investigation/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md; feature 002 baseline implemented

**Tests**: Per constitution (Principle XII), automated tests are NOT used. Manual validation steps reference `quickstart.md` (VS-1 through VS-8).

**Organization**: Tasks grouped by user story for independent implementation and manual validation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1–US6)
- Include exact file paths in descriptions

## Path Conventions

- **Go TUI application**: `cmd/binlog-explorer/`, `internal/` at repository root
- Package layout: `internal/events`, `internal/explorer`, `internal/filters`, `internal/sources/mysql`, `internal/ui`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Feature 003 prerequisites on top of feature 002 baseline

- [x] T001 Verify `go build ./cmd/binlog-explorer` succeeds on branch `003-scoped-binlog-investigation` and document baseline in `specs/003-scoped-binlog-investigation/quickstart.md` prerequisites section
- [x] T002 [P] Add large-file warning constants (`LargeFileSizeBytes` = 1 GiB, `LargeFileEventCount` = 500000) in `internal/sources/mysql/limits.go` per `specs/003-scoped-binlog-investigation/research.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared types and session/source extensions required by all user stories

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T003 [P] Define `InvestigationScope`, `ScopePreset`, and `FileAnalysisResult` types in `internal/events/scope.go` per `specs/003-scoped-binlog-investigation/data-model.md`
- [x] T004 [P] Extract shared user-data event classification helper from `internal/sources/mysql/indexer.go` into `internal/sources/mysql/classify.go` for reuse by analyzer and indexer
- [x] T005 Add `SourceStateAnalyzing`, analysis cache fields, and `ActiveScope` to `internal/sources/mysql/source.go` per `specs/003-scoped-binlog-investigation/data-model.md`
- [x] T006 Extend `internal/explorer/session.go` with `InvestigationScope`, `LaunchScope`, `AwaitingScope`, `SessionAnalysisSummary`, and analysis merge helpers

**Checkpoint**: Project compiles; new types and session fields exist; open flow not yet changed

---

## Phase 3: User Story 1 - Analyze Binlog Before Indexing (Priority: P1) 🎯 MVP

**Goal**: Lightweight analysis pass reports size, min/max timestamps, and approximate change-event count before any browse index is built

**Independent Validation**: VS-1 in `specs/003-scoped-binlog-investigation/quickstart.md` — analysis progress, metadata display, cancel aborts open, no browse index during analysis

### Implementation for User Story 1

- [x] T007 [US1] Implement `AnalyzeStream` with min/max timestamp and approximate count tracking in `internal/sources/mysql/analyzer.go` using `classify.go` (no `EventSummary` accumulation); on mid-read failure mark source `error`, keep session and other sources usable (FR-018, feature 002 pattern)
- [x] T008 [US1] Add `AnalysisDoneMsg` / progress updates and background analysis goroutine in `internal/ui/app.go` (mirror indexer message pattern); surface analysis failures in status bar without crashing the TUI
- [x] T009 [US1] Change open flow to run analysis instead of immediate indexing in `internal/ui/app.go` and `internal/ui/openfile.go`
- [x] T010 [US1] Show `Analyzing: N%` status and `Esc` cancel during analysis (abort open, close source) in `internal/ui/app.go`
- [x] T011 [US1] Cache `FileAnalysisResult` on `Source` and compute merged `SessionAnalysisSummary` in `internal/explorer/session.go`

**Checkpoint**: Opening a binlog runs analysis with progress; metadata cached; cancel works; list still empty until scope (US2)

---

## Phase 4: User Story 2 - Choose Investigation Scope at Open (Priority: P1)

**Goal**: Mandatory scope dialog with presets, custom range, and large-file warning before indexing begins

**Independent Validation**: VS-2 and VS-3 in `specs/003-scoped-binlog-investigation/quickstart.md` — scope dialog blocks list; presets work; Entire file shows warning on large binlog

### Implementation for User Story 2

- [x] T012 [P] [US2] Implement `ScopeDialogModel` with options 1–4 and custom From/To inputs in `internal/ui/scopedialog.go` per `specs/003-scoped-binlog-investigation/contracts/scope-dialog.md`
- [x] T013 [US2] Implement large-file secondary warning modal (Y/F) in `internal/ui/scopedialog.go` using thresholds from `internal/sources/mysql/limits.go`
- [x] T014 [US2] Gate browse list population on confirmed scope (`AwaitingScope` / `InvestigationScope` set) in `internal/ui/app.go`
- [x] T015 [US2] Wire scope dialog display after analysis completes (when no CLI launch scope) in `internal/ui/app.go`
- [x] T016 [US2] Validate custom range ⊆ merged analysis min/max with clear error in `internal/ui/scopedialog.go`

**Checkpoint**: After analysis, scope dialog appears; list remains empty until scope confirmed; large-file warning works

---

## Phase 5: User Story 3 - Scoped Indexing with Early Stop (Priority: P1)

**Goal**: Index only events within `[From, To]` inclusive; skip before `From`; stop after `To` (best-effort)

**Independent Validation**: VS-4 in `specs/003-scoped-binlog-investigation/quickstart.md` — narrow scope faster than full file; first events within SC-001 targets

### Implementation for User Story 3

- [x] T017 [US3] Extend `IndexStream` to accept `InvestigationScope` with skip-before-From and stop-after-To logic in `internal/sources/mysql/indexer.go`
- [x] T018 [US3] For **Entire file** scope, index using cached `[MergedMin, MergedMax]` from analysis with early stop disabled — **one index pass after analysis** (second read only; FR-019); do not add a third combined analyze+index pass in `internal/sources/mysql/indexer.go`
- [x] T019 [US3] Start scoped indexer after scope confirmation in `internal/ui/app.go` (replace full-file indexer trigger)
- [x] T020 [US3] Update status bar for scoped indexing (`Indexing scope: From..To | N%`) in `internal/ui/app.go`
- [x] T021 [US3] Preserve on-demand detail load semantics (no change to payload storage) verifying `internal/sources/mysql/detail.go` still works with scoped index offsets

**Checkpoint**: Scoped index populates list; early stop on narrow range; detail inspection works on scoped events

---

## Phase 6: User Story 5 - Launch with Scope from CLI (Priority: P2)

**Goal**: `--from` and `--to` flags skip scope dialog when both provided; reject invalid partial flags and out-of-bounds ranges

**Independent Validation**: VS-5 in `specs/003-scoped-binlog-investigation/quickstart.md` — CLI scope launch; error on single flag; error on out-of-bounds range

### Implementation for User Story 5

- [x] T022 [P] [US5] Parse `--from` and `--to` with formats `2006-01-02 15:04:05` and date-only in `cmd/binlog-explorer/main.go` per `specs/003-scoped-binlog-investigation/contracts/cli.md`
- [x] T023 [US5] Enforce both-or-neither CLI boundary rule with exit code 1 in `cmd/binlog-explorer/main.go`
- [x] T024 [US5] Pass parsed launch scope to session as `LaunchScope` in `cmd/binlog-explorer/main.go` and `internal/explorer/session.go`
- [x] T025 [US5] Skip scope dialog when `LaunchScope` set; begin scoped indexing after analysis in `internal/ui/app.go`
- [x] T026 [US5] Validate CLI scope against merged analysis min/max after analysis; show TUI error and do not index if out of bounds (FR-011a) in `internal/explorer/session.go` and `internal/ui/app.go` — no exit `1` at flag parse for this case

**Checkpoint**: CLI launch with valid range skips dialog and indexes; invalid flags and out-of-bounds ranges fail clearly

---

## Phase 7: User Story 4 - Change Investigation Scope During Session (Priority: P2)

**Goal**: Key `s` reopens scope dialog; scope change replaces index; mid-session `o` re-analyzes and re-indexes all sources

**Independent Validation**: VS-6 in `specs/003-scoped-binlog-investigation/quickstart.md` — scope change replaces list; cached analysis reused; mid-session open triggers re-scope

### Implementation for User Story 4

- [x] T027 [P] [US4] Add key `s` and help text for scope dialog in `internal/ui/keys.go` per `specs/003-scoped-binlog-investigation/contracts/keybindings.md`
- [x] T028 [US4] Implement `ApplyScopeChange` (clear index, reset selection/detail, clear secondary filters) in `internal/explorer/session.go`
- [x] T029 [US4] Open scope dialog on `s` reusing cached analysis (no re-analyze) in `internal/ui/app.go`
- [x] T030 [US4] Implement in-memory narrow optimization when new scope ⊆ current scope in `internal/explorer/session.go`; re-index fallback otherwise
- [x] T031 [US4] Wire mid-session `o` flow: analyze new source, merge metadata, scope dialog, re-index all sources (FR-013a) in `internal/ui/openfile.go` and `internal/ui/app.go`

**Checkpoint**: Scope change via `s` works; widening re-indexes; additional binlog in session triggers full re-scope flow

---

## Phase 8: User Story 6 - Secondary Filters on Scoped Index (Priority: P3)

**Goal**: Remove time from filter editor; table/schema/op filters apply in memory over scoped index; status shows `shown / scoped-total`

**Independent Validation**: VS-7 and VS-8 in `specs/003-scoped-binlog-investigation/quickstart.md` — instant secondary filter; scope label in top bar

### Implementation for User Story 6

- [x] T032 [P] [US6] Remove `TimeStart`/`TimeEnd` from `internal/filters/filter.go` and update `Apply`/`Summary` for schema/table/op only
- [x] T033 [US6] Remove time range fields from filter editor in `internal/ui/filterbar.go` (Tab cycle: operation → schema → table)
- [x] T034 [US6] Update `FilteredLabel` and status bar to `shown / scoped-total` in `internal/explorer/session.go` and `internal/ui/app.go`
- [x] T035 [US6] Show active investigation scope in top bar in `internal/ui/app.go` per `specs/003-scoped-binlog-investigation/contracts/ui-layout.md`
- [x] T036 [US6] Verify inspect-on-filtered and clear-filters restore scoped index (not entire file) in `internal/ui/app.go`

**Checkpoint**: Secondary filters instant; no time fields in `f` modal; counts reflect scoped total

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, validation, and regression checks across all stories

- [x] T037 [P] Align `specs/003-scoped-binlog-investigation/contracts/cli.md` exit code notes with FR-011a out-of-bounds behavior (TUI error after analysis, not startup exit `1`)
- [x] T038 Update help overlay text in `internal/ui/keys.go` for analysis, scope (`s`), and revised filter (`f`) workflows
- [x] T039 Run `go build ./cmd/binlog-explorer` and fix compile errors across modified packages
- [ ] T040 Manual validation pass VS-1 through VS-8 documenting results in `specs/003-scoped-binlog-investigation/quickstart.md`
- [ ] T041 Regression check: feature 002 browse, detail, multi-source merge, and partial CLI open failures still work within scoped workflow

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies — start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 — **blocks all user stories**
- **Phase 3 (US1)**: Depends on Phase 2
- **Phase 4 (US2)**: Depends on US1 (analysis must exist before scope dialog)
- **Phase 5 (US3)**: Depends on US2 (scope must be confirmable before scoped index)
- **Phase 6 (US5)**: Depends on US3 (scoped indexing must work); can parallelize with US4 after US3
- **Phase 7 (US4)**: Depends on US2 + US3 (scope dialog and scoped indexer)
- **Phase 8 (US6)**: Depends on US3 (scoped index exists); best after US4
- **Phase 9 (Polish)**: Depends on desired user stories complete

### User Story Dependencies

| Story | Depends on | Independent test |
|-------|------------|------------------|
| US1 | Foundational | VS-1 analysis only |
| US2 | US1 | VS-2, VS-3 scope dialog |
| US3 | US2 | VS-4 scoped index |
| US5 | US3 | VS-5 CLI |
| US4 | US2, US3 | VS-6 scope change |
| US6 | US3 | VS-7, VS-8 filters |

### Parallel Opportunities

**Phase 2** (after T003):

```text
T003 scope.go  ||  T004 classify.go
```

**Phase 4**:

```text
T012 scopedialog.go (can start UI model while T014 gates app.go)
```

**Phase 6 + 7** (after US3 complete):

```text
Developer A: US5 (T022–T026 CLI)
Developer B: US4 (T027–T031 scope change)
```

**Phase 8**:

```text
T032 filter.go  ||  T033 filterbar.go
```

---

## Parallel Example: User Story 1

```bash
# After Phase 2 complete:
# Sequential core: T007 analyzer.go → T008–T011 app.go + session.go
# T007 and T003/T004 can be prepared in parallel during Phase 2
```

---

## Implementation Strategy

### MVP First (US1 + US2 + US3)

1. Complete Phase 1–2 (foundation)
2. Complete US1 → validate VS-1 (analysis)
3. Complete US2 → validate VS-2/VS-3 (scope dialog)
4. Complete US3 → validate VS-4 (scoped indexing) — **first usable scoped explorer**
5. STOP and demo before CLI/scope-change/filter split

### Incremental Delivery

1. Foundation → US1 → US2 → US3 = **core scoped workflow (MVP)**
2. Add US5 (CLI) → validate VS-5
3. Add US4 (scope change + mid-session open) → validate VS-6
4. Add US6 (secondary filter split) → validate VS-7/VS-8
5. Polish → T040 full validation

### Suggested MVP Scope

**User Stories 1–3 (P1)**: Analysis + scope dialog + scoped indexing deliver the primary performance and UX value. US5/US4/US6 enhance workflow but are not required for first scoped investigation demo.

---

## Notes

- Feature 002 `internal/filters/filter.go` time-range fields are **removed** in US6; do not use post-index date filter for scope
- Mid-session open (`o`) always triggers re-scope and re-index of all sources per clarified spec (FR-013a)
- Entire-file path: analysis (pass 1) + one index pass using cached min/max (pass 2) — never a third read (FR-019 / T018)
- Analysis mid-read errors: mark source error, session continues (T007/T008)
- Manual validation only; record pass/fail in quickstart.md after T040
