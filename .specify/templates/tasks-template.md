---

description: "Task list template for feature implementation"
---

# Tasks: [FEATURE NAME]

**Input**: Design documents from `/specs/[###-feature-name]/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Per constitution (Principle XII), automated tests are NOT used. Manual validation steps belong in spec acceptance scenarios and quickstart.md.

**Organization**: Tasks are grouped by user story to enable independent implementation and manual validation of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Go TUI application**: `cmd/binlog-explorer/`, `internal/` at repository root
- Package layout follows constitution: `internal/events`, `internal/explorer`, `internal/filters`, `internal/sources/mysql`, `internal/ui`
- Paths shown below assume this layout - adjust based on plan.md structure

<!--
  ============================================================================
  IMPORTANT: The tasks below are SAMPLE TASKS for illustration purposes only.

  The /speckit-tasks command MUST replace these with actual tasks based on:
  - User stories from spec.md (with their priorities P1, P2, P3...)
  - Feature requirements from plan.md
  - Entities from data-model.md
  - Endpoints from contracts/

  Tasks MUST be organized by user story so each story can be:
  - Implemented independently
  - Manually validated independently (real binlog datasets)
  - Delivered as an MVP increment

  DO NOT include automated test tasks unless the constitution is amended.
  ============================================================================
-->

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Create project structure per implementation plan (`cmd/`, `internal/`)
- [ ] T002 Initialize Go module with Bubble Tea, Bubbles, Lip Gloss dependencies
- [ ] T003 [P] Configure Go formatting and linting tools

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

Examples of foundational tasks (adjust based on your project):

- [ ] T004 Implement MySQL binlog source reader in `internal/sources/mysql/`
- [ ] T005 [P] Define core event models in `internal/events/`
- [ ] T006 [P] Setup Bubble Tea application shell in `internal/ui/`
- [ ] T007 Create base explorer navigation in `internal/explorer/`
- [ ] T008 Configure error handling and user-facing messages in TUI
- [ ] T009 Setup configuration (binlog path, display options)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - [Title] (Priority: P1) 🎯 MVP

**Goal**: [Brief description of what this story delivers]

**Independent Validation**: [How a DBA manually verifies this story with real binlog data]

### Implementation for User Story 1

- [ ] T010 [P] [US1] Create event types in `internal/events/`
- [ ] T011 [P] [US1] Add parser logic in `internal/sources/mysql/`
- [ ] T012 [US1] Implement explorer view in `internal/explorer/` (depends on T010, T011)
- [ ] T013 [US1] Wire TUI screens and keyboard navigation in `internal/ui/`
- [ ] T014 [US1] Add filtering/search for user story 1 in `internal/filters/`
- [ ] T015 [US1] Document manual validation steps in `quickstart.md`

**Checkpoint**: At this point, User Story 1 should be fully functional and manually validatable

---

## Phase 4: User Story 2 - [Title] (Priority: P2)

**Goal**: [Brief description of what this story delivers]

**Independent Validation**: [How a DBA manually verifies this story]

### Implementation for User Story 2

- [ ] T016 [P] [US2] Extend event models in `internal/events/`
- [ ] T017 [US2] Implement feature logic in `internal/explorer/`
- [ ] T018 [US2] Add TUI integration in `internal/ui/`
- [ ] T019 [US2] Integrate with User Story 1 components (if needed)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - [Title] (Priority: P3)

**Goal**: [Brief description of what this story delivers]

**Independent Validation**: [How a DBA manually verifies this story]

### Implementation for User Story 3

- [ ] T020 [P] [US3] Extend models or filters in `internal/`
- [ ] T021 [US3] Implement feature in `internal/explorer/`
- [ ] T022 [US3] Add TUI screens and shortcuts in `internal/ui/`

**Checkpoint**: All user stories should now be independently functional

---

[Add more user story phases as needed, following the same pattern]

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] TXXX [P] Documentation updates in docs/ or feature quickstart.md
- [ ] TXXX Code cleanup and refactoring
- [ ] TXXX Performance review for large binlog files
- [ ] TXXX Manual validation pass with real-world binlog datasets
- [ ] TXXX Keyboard shortcut and UX consistency review
- [ ] TXXX Run quickstart.md validation

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - May integrate with US1 but should be independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - May integrate with US1/US2 but should be independently testable

### Within Each User Story

- Parser/model before UI wiring
- Core explorer logic before advanced filters
- Story complete before moving to next priority
- Manual validation documented in quickstart.md after each checkpoint

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- Models within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch parallel model/parser tasks for User Story 1:
Task: "Create event types in internal/events/"
Task: "Add parser logic in internal/sources/mysql/"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Manually validate User Story 1 with real binlog data
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Manually validate → Demo (MVP!)
3. Add User Story 2 → Manually validate → Demo
4. Add User Story 3 → Manually validate → Demo
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1
   - Developer B: User Story 2
   - Developer C: User Story 3
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and manually validatable
- Document validation steps in quickstart.md after each checkpoint
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
