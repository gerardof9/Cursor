# Tasks: Open and Explore Database Change Events

**Input**: Design documents from `specs/002-explore-binlog-events/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Per constitution (Principle XII), automated tests are NOT used. Manual validation steps reference `quickstart.md`.

**Organization**: Tasks are grouped by user story to enable independent implementation and manual validation of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Go TUI application**: `cmd/binlog-explorer/`, `internal/` at repository root
- Package layout: `internal/events`, `internal/explorer`, `internal/filters`, `internal/sources/mysql`, `internal/ui`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and dependency setup

- [ ] T001 Create project directory structure (`cmd/binlog-explorer/`, `internal/events/`, `internal/explorer/`, `internal/filters/`, `internal/sources/mysql/`, `internal/ui/`) per `specs/002-explore-binlog-events/plan.md`
- [ ] T002 Initialize `go.mod` (Go 1.22+) with dependencies: `bubbletea`, `bubbles`, `lipgloss`, `go-mysql-org/go-mysql` in repository root
- [ ] T003 [P] Add `.gitignore` for `bin/`, `binlog-explorer`, and OS artifacts at repository root

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core types, session shell, and TUI scaffold that all user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T004 [P] Define `Operation` and `ReplicationFormat` types in `internal/events/types.go` per `specs/002-explore-binlog-events/data-model.md`
- [ ] T005 [P] Define `EventSummary`, `EventDetail`, and `RowChange` structs in `internal/events/event.go` per `specs/002-explore-binlog-events/data-model.md`
- [ ] T006 Create `ExplorerSession` struct with source registry and empty index in `internal/explorer/session.go`
- [ ] T007 [P] Implement key map and help metadata in `internal/ui/keys.go` per `specs/002-explore-binlog-events/contracts/keybindings.md`
- [ ] T008 Create root `tea.Model` with split-pane layout shell in `internal/ui/app.go` per `specs/002-explore-binlog-events/contracts/ui-layout.md`
- [ ] T009 Create application entry point launching empty TUI in `cmd/binlog-explorer/main.go`

**Checkpoint**: `go build ./cmd/binlog-explorer` succeeds; empty split-pane TUI launches and quits with `q`

---

## Phase 3: User Story 1 - Open Binlogs (Priority: P1) 🎯 MVP

**Goal**: DBA can open binlog files via CLI at launch and in-session from the TUI; errors surface without crashing

**Independent Validation**: VS-1 in `specs/002-explore-binlog-events/quickstart.md` — CLI open, in-session `o` open, multi-file session, invalid file handling

### Implementation for User Story 1

- [ ] T010 [US1] Implement `BinlogSource` open, validation, and state tracking in `internal/sources/mysql/source.go`
- [ ] T011 [US1] Add `OpenSource(path)` and source list accessors to `internal/explorer/session.go`
- [ ] T012 [US1] Implement CLI argument parsing, TTY check, and launch-time file open in `cmd/binlog-explorer/main.go` per `specs/002-explore-binlog-events/contracts/cli.md`
- [ ] T013 [US1] Implement in-session path input modal in `internal/ui/openfile.go` (triggered by `o`)
- [ ] T014 [US1] Wire open-file flow, per-file errors, and status bar messages in `internal/ui/app.go`
- [ ] T015 [US1] Document VS-1 validation results in `specs/002-explore-binlog-events/quickstart.md` (append pass/fail notes section)

**Checkpoint**: Open one or two binlog files via CLI and in-session; sources visible in status; invalid paths show status error; app remains usable. Event list not required for US1 completion.

---

## Phase 4: User Story 2 - Browse Events (Priority: P2)

**Goal**: Chronological change-event list with progressive indexing and keyboard navigation

**Independent Validation**: VS-2 in `specs/002-explore-binlog-events/quickstart.md` — ordered list, no housekeeping events, responsive navigation on large sample

### Implementation for User Story 2

- [ ] T016 [P] [US2] Implement stream indexer mapping `RowsEvent`/`QueryEvent` to `EventSummary` in `internal/sources/mysql/indexer.go` (skip housekeeping per `research.md`; propagate mid-read errors without crashing session)
- [ ] T017 [US2] Add index merge, chronological sort, `tea.Msg` batch updates, and source `error` state on indexer failure in `internal/explorer/session.go`
- [ ] T018 [US2] Implement event list view with TIME/OP/SCHEMA/TABLE/SRC columns in `internal/ui/list.go`
- [ ] T019 [US2] Wire list navigation (↑/↓/PgUp/PgDn/Home/End), indexing progress, empty-source state, empty-index message (no user-data events), and per-source error status in `internal/ui/app.go` (target < 100ms perceived nav per `plan.md`)
- [ ] T020 [US2] Document VS-2 validation results in `specs/002-explore-binlog-events/quickstart.md`

**Checkpoint**: Events appear progressively; list is chronological across multiple sources; keyboard navigation works

---

## Phase 5: User Story 3 - Inspect Event Details (Priority: P3)

**Goal**: Select an event to load and display full detail on demand

**Independent Validation**: VS-3 in `specs/002-explore-binlog-events/quickstart.md` — row and statement events, detail updates on selection change, incomplete metadata notes

### Implementation for User Story 3

- [ ] T021 [P] [US3] Implement on-demand `EventDetail` loader from file offset in `internal/sources/mysql/detail.go`
- [ ] T022 [US3] Add selection state and single-entry detail cache to `internal/explorer/session.go`
- [ ] T023 [US3] Implement detail viewport (header, SQL/rows, completeness notes) in `internal/ui/detail.go`
- [ ] T024 [US3] Wire async detail load on selection with loading indicator in `internal/ui/app.go`
- [ ] T025 [US3] Document VS-3 validation results in `specs/002-explore-binlog-events/quickstart.md`

**Checkpoint**: Selecting events populates detail pane; statement events note missing row images when applicable

---

## Phase 6: User Story 4 - Filter Events (Priority: P4)

**Goal**: Filter events by operation, schema/table, and inclusive time range; clear restores full list

**Independent Validation**: VS-4 in `specs/002-explore-binlog-events/quickstart.md` — table, operation, time boundary, and clear filters

### Implementation for User Story 4

- [ ] T026 [P] [US4] Implement `FilterCriteria` and AND-apply logic with inclusive time bounds in `internal/filters/filter.go`
- [ ] T027 [US4] Add `ApplyFilter`, `ClearFilter`, and `Filtered` index computation to `internal/explorer/session.go`
- [ ] T028 [US4] Implement filter editor modal (operation, schema, table, time range) in `internal/ui/filterbar.go`
- [ ] T029 [US4] Wire `f`/`c` keybindings and filtered list rendering in `internal/ui/app.go` and `internal/ui/list.go`
- [ ] T030 [US4] Document VS-4 validation results in `specs/002-explore-binlog-events/quickstart.md`

**Checkpoint**: Filters narrow the list correctly; `c` restores full index; zero-match shows empty state message

---

## Phase 7: User Story 5 - Investigate Activity (Priority: P5)

**Goal**: Integrated filter → browse → inspect workflow for real troubleshooting scenarios

**Independent Validation**: VS-5 and VS-6 in `specs/002-explore-binlog-events/quickstart.md` — end-to-end investigation under 5 minutes; performance spot checks

### Implementation for User Story 5

- [ ] T031 [US5] Implement selection reconciliation on filter change (nearest match or clear with status) in `internal/explorer/session.go`
- [ ] T032 [US5] Enhance status bar with filtered/total counts, active filter summary, and indexing % in `internal/ui/app.go`
- [ ] T033 [US5] Ensure detail inspection works on filtered results and empty-filter UX in `internal/ui/list.go` and `internal/ui/detail.go`
- [ ] T034 [US5] Document VS-5 and VS-6 validation results in `specs/002-explore-binlog-events/quickstart.md`

**Checkpoint**: Complete investigation scenario ("what happened to table X between A and B?") without external tools

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Help, terminal guards, exit codes, and full manual validation pass

- [ ] T035 [P] Implement help overlay (`?`) and minimum 80×24 terminal guard in `internal/ui/app.go`
- [ ] T036 [P] Implement exit codes and `--help`/`--version` flags in `cmd/binlog-explorer/main.go` per `specs/002-explore-binlog-events/contracts/cli.md`
- [ ] T037 Run full manual validation checklist (VS-1 through VS-6) and record outcomes in `specs/002-explore-binlog-events/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup — **BLOCKS all user stories**
- **User Stories (Phases 3–7)**: Depend on Foundational; execute sequentially P1 → P5 (each builds on prior)
- **Polish (Phase 8)**: Depends on User Story 5 completion

### User Story Dependencies

- **US1 (P1)**: After Foundational — no prior story dependencies
- **US2 (P2)**: After US1 — requires open sources and session wiring
- **US3 (P3)**: After US2 — requires indexed list and selection
- **US4 (P4)**: After US2 — requires index (detail from US3 not required for filtering but recommended before US5)
- **US5 (P5)**: After US3 and US4 — integrates browse, inspect, and filter

### Within Each User Story

- Parser/source logic before explorer session integration
- Explorer session before UI wiring
- UI components before app.go integration
- Manual validation task last in each story phase

### Parallel Opportunities

- **Phase 1**: T003 parallel with T002 after T001
- **Phase 2**: T004, T005, T007 parallel after T001; T008 after T006
- **US2**: T016 parallel with prep work before T017
- **US3**: T021 parallel before T022
- **US4**: T026 parallel before T027
- **Phase 8**: T035 and T036 parallel

---

## Parallel Example: Foundational Phase

```bash
# After T001–T003, launch in parallel:
Task T004: "Define Operation and ReplicationFormat in internal/events/types.go"
Task T005: "Define EventSummary, EventDetail, RowChange in internal/events/event.go"
Task T007: "Implement key map in internal/ui/keys.go"

# Then sequentially:
Task T006 → T008 → T009
```

---

## Parallel Example: User Story 2

```bash
# Launch indexer while session merge is being prepared:
Task T016: "Implement stream indexer in internal/sources/mysql/indexer.go"

# After T016 completes:
Task T017 → T018 → T019 → T020
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: User Story 1 (T010–T015)
4. **STOP and VALIDATE**: VS-1 with real binlog files
5. Demo file-open workflow if ready

### Incremental Delivery

1. Setup + Foundational → empty TUI shell
2. US1 → open files → validate VS-1
3. US2 → browse indexed events → validate VS-2
4. US3 → inspect details → validate VS-3
5. US4 → filter events → validate VS-4
6. US5 → full investigation workflow → validate VS-5/VS-6
7. Polish → help, exit codes, full checklist

### Suggested MVP Scope

**User Story 1 (Phase 3)** — proves file loading from CLI and TUI is viable before investing in parser/index complexity.

---

## Notes

- [P] tasks = different files, no incomplete-task dependencies
- [Story] label maps task to user story for traceability
- Each user story ends with a quickstart validation documentation task
- Commit after each task or logical group
- Stop at any checkpoint to validate before proceeding
- Do NOT create automated test files or `*_test.go` packages (constitution Principle XII)
