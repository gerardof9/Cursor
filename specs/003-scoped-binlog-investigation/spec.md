# Feature Specification: Scoped Binlog Investigation with Pre-Open Analysis

**Feature Branch**: `003-scoped-binlog-investigation`

**Created**: 2026-06-10

**Status**: Clarified

**Input**: User description: "Introduce a two-phase open workflow: file analysis pass, mandatory investigation scope selection, and scoped indexing with early stop. Preserve browse, inspect, and secondary filters from feature 002 while making time range a load criterion rather than only a post-index view filter."

## Clarifications

### Session 2026-06-10

- Q: What happens when only one CLI date boundary (`--from` or `--to`) is provided? → A: Both boundaries are required when using CLI scope; if only one is provided, the application reports a clear error and does not begin indexing.
- Q: How should narrowing the investigation scope behave in v1? → A: Any scope change replaces the in-memory index; narrowing within an already-indexed range may filter in memory as an optimization, but re-indexing on scope change is always acceptable and is the required fallback when in-memory narrowing cannot be proven safe.
- Q: How does the DBA change investigation scope during a session? → A: Via key `s` opening the investigation scope dialog (separate from table/operation filters opened with `f`).
- Q: When the DBA opens an additional binlog file mid-session (`o`) while a scoped index already exists, what should happen? → A: Analyze the new file, merge analysis metadata into the session summary, show the scope dialog to confirm or adjust the investigation scope, then re-index all open sources with the confirmed scope.
- Q: When CLI `--from`/`--to` falls outside the detected min/max timestamp span of opened binlog(s), what should happen? → A: Reject with a clear error; do not begin scoped indexing.

## User Scenarios & Manual Validation *(mandatory)*

### User Story 1 - Analyze Binlog Before Indexing (Priority: P1)

As a DBA, when I open one or more binlog files, I want the application to perform an analysis pass so I understand file size and temporal coverage before committing to a heavy index.

**Why this priority**: Without temporal bounds and size context, the DBA cannot make an informed scope choice. This is the entry point for the new open workflow.

**Independent Validation**: Open a known large binlog from a test environment. Confirm analysis progress is shown, then verify the application reports file size, minimum timestamp, maximum timestamp, and an approximate change-event count before any browse index is built.

**Acceptance Scenarios**:

1. **Given** the DBA opens one or more valid binlog files (CLI or in-session), **When** open completes validation, **Then** a file analysis pass runs per source before scoped indexing begins.
2. **Given** analysis is running, **When** the DBA views status, **Then** progress is shown for large files and the DBA can cancel to abort the open without building a browse index.
3. **Given** analysis completes, **When** results are presented, **Then** each source shows human-readable file size, detected minimum event timestamp, detected maximum event timestamp, and approximate user-data change-event count (prefixed as approximate, e.g. `~`).
4. **Given** multiple binlog files are opened together, **When** analysis completes, **Then** a merged session summary shows aggregate size and merged min/max time span across sources.
5. **Given** analysis results are cached for a source in the session, **When** the DBA changes investigation scope later without closing the file, **Then** analysis is not repeated unless the source is closed and reopened.

---

### User Story 2 - Choose Investigation Scope at Open (Priority: P1)

As a DBA, I want to choose what portion of the binlog timeline to index so I do not accidentally index an entire multi-GB file without meaning to.

**Why this priority**: Scope selection prevents blind full-file indexing—the most common pain point on large binlogs—while still allowing full-file investigation after explicit confirmation.

**Independent Validation**: Open a multi-GB binlog without CLI date flags. Confirm the scope dialog appears with file metadata, select "Last day (of file)", and verify indexing begins only after scope confirmation.

**Acceptance Scenarios**:

1. **Given** file analysis has completed, **When** no investigation scope is yet defined, **Then** a mandatory scope dialog is shown with message "No investigation scope defined" and cannot be skipped except by cancel/quit.
2. **Given** the scope dialog is visible, **When** the DBA views it, **Then** it shows selected file(s), total size, detected time span, and approximate change-event count.
3. **Given** the scope dialog is visible, **When** the DBA selects a scope option, **Then** one of the following applies before indexing starts:
   - **Entire file** — index all user-data change events in the open source(s).
   - **Last hour (of file)** — index events from `(max_timestamp − 1 hour)` through `max_timestamp`, inclusive.
   - **Last day (of file)** — index events from `(max_timestamp − 24 hours)` through `max_timestamp`, inclusive.
   - **Custom range** — DBA enters From and To; fields are pre-filled with detected min/max; out-of-range values show a clear error.
4. **Given** the DBA selects **Entire file** on a file exceeding configurable large-file thresholds (size and/or approximate event count), **When** confirming scope, **Then** a secondary warning appears stating that full indexing may take several minutes and consume significant memory, with options to continue or return to define a date range.
5. **Given** the DBA has not confirmed a scope, **When** they attempt to browse events, **Then** the browse list does not populate.

---

### User Story 3 - Scoped Indexing with Early Stop (Priority: P1)

As a DBA, I want indexing to include only events within my chosen time scope so investigation starts faster and uses less memory.

**Why this priority**: This delivers the core performance and memory benefit of the feature; scope selection alone has no value without scoped indexing.

**Independent Validation**: Open a large binlog, choose a one-day custom range known to cover a small fraction of the file. Confirm first indexed events appear quickly, indexing completes sooner than a full-file index of the same file, and memory use stays proportional to the scoped event count.

**Acceptance Scenarios**:

1. **Given** an investigation scope `[From, To]` is confirmed, **When** scoped indexing runs, **Then** only user-data change events with timestamps within `[From, To]` inclusive are added to the browse index.
2. **Given** scoped indexing encounters events before `From`, **When** parsing continues, **Then** those events are skipped without being indexed.
3. **Given** scoped indexing encounters an event after `To` and timestamps are monotonic within the file, **When** parsing continues, **Then** parsing of that source stops (early termination).
4. **Given** multiple sources are open with the same scope, **When** indexing runs, **Then** each source is indexed in the background and results merge into one chronological browse list.
5. **Given** scoped indexing is in progress, **When** the DBA views status, **Then** progress shows bytes read and/or the current timestamp being processed.
6. **Given** an event is selected from the scoped list, **When** detail is requested, **Then** full detail loads on demand by file position (same semantics as feature 002).

---

### User Story 4 - Change Investigation Scope During Session (Priority: P2)

As a DBA, after working on one time range, I want to investigate a different range in the same session without restarting the application.

**Why this priority**: Real investigations often pivot to adjacent time windows; this must not require a full application restart.

**Independent Validation**: Index a one-day scope, browse several events, then open scope selection again and choose a different non-overlapping day. Confirm the previous scoped list is replaced, re-indexing progress is shown, and selection/detail reset with a clear status message.

**Acceptance Scenarios**:

1. **Given** a scoped index is loaded, **When** the DBA opens the in-session scope selection flow, **Then** cached analysis metadata (min/max/count) is reused without a full re-analysis pass.
2. **Given** the DBA selects a new investigation scope, **When** scope is applied, **Then** the previous in-memory browse index for affected sources is replaced (previous scoped results are not retained in the list).
3. **Given** the new scope is wider than the current index or switches to Entire file, **When** scope is applied, **Then** binlog(s) are re-parsed with the new scope and progress is shown.
4. **Given** the new scope is narrower and provably contained within the already-indexed range, **When** scope is applied, **Then** the application may narrow by filtering the existing index in memory without re-parse (optimization); otherwise it re-indexes.
5. **Given** scope changes, **When** indexing completes or begins, **Then** list selection and detail pane reset predictably and the status bar explains that scope changed.

---

### User Story 5 - Launch with Scope from CLI (Priority: P2)

As a DBA working over SSH on the database server, I want to pass a date range on the command line so I skip the scope dialog when I already know my window.

**Why this priority**: Matches the common server-side workflow where the time window is known before launch.

**Independent Validation**: Launch with `binlog-explorer --from <ts> --to <ts> sample.bin`. Confirm analysis runs, scope is applied automatically, scope dialog is skipped, and scoped indexing begins.

**Acceptance Scenarios**:

1. **Given** binlog path(s) and both `--from` and `--to` are provided at launch, **When** the application starts, **Then** analysis runs, the provided range is applied as investigation scope, the scope dialog is skipped, and scoped indexing begins.
2. **Given** binlog path(s) are provided without date flags, **When** the application starts, **Then** analysis runs and the scope dialog is shown before indexing (User Story 2).
3. **Given** only one of `--from` or `--to` is provided, **When** launch is attempted, **Then** the application reports a clear error requiring both boundaries or neither.
4. **Given** both `--from` and `--to` are provided but fall outside the merged detected min/max timestamp span after analysis, **When** launch completes analysis, **Then** the application reports a clear error and does not begin scoped indexing.
5. **Given** multiple binlog paths are provided at launch with partial open failures, **When** launch completes, **Then** behavior matches feature 002 (warnings for failures, continue with successful sources).

---

### User Story 6 - Secondary Filters on Scoped Index (Priority: P3)

As a DBA, within my chosen time scope, I want to filter by table, schema, and operation type without re-reading the entire binlog.

**Why this priority**: Table and operation narrowing is the second most common investigation step after choosing a time window.

**Independent Validation**: After scoped indexing completes, apply a table filter and confirm results narrow instantly with no re-index progress shown. Change investigation scope and confirm re-indexing occurs.

**Acceptance Scenarios**:

1. **Given** a scoped browse index is loaded, **When** the DBA applies schema, table, or operation filters, **Then** filtering occurs in memory over the scoped index without re-parsing binlog files.
2. **Given** secondary filters are active, **When** the DBA inspects events in the filtered list, **Then** detail inspection works without removing filters.
3. **Given** secondary filters are active, **When** the DBA clears filters, **Then** the full scoped index is restored (not the entire file unless scope is Entire file).
4. **Given** the status bar shows event counts, **When** filters are active, **Then** counts display as `shown / scoped-total` (not entire-file total unless scope is Entire file).
5. **Given** the DBA changes investigation scope, **When** scope is applied, **Then** secondary filters are cleared; table/operation filter changes never trigger binlog re-parse.
6. **Given** a scoped index is already loaded, **When** the DBA opens an additional binlog file via `o`, **Then** the new file is analyzed, session analysis metadata is merged, the scope dialog is shown to confirm or adjust scope, and all open sources are re-indexed with the confirmed scope (replacing the prior index).

---

### Edge Cases

- What happens when a binlog file is empty or contains no user-data change events? *(Resolved: analysis completes with zero approximate count; scope dialog still shown; after scoped index, show empty-index message; source remains registered.)*
- What happens when analysis or scoped indexing fails mid-read? *(Resolved: mark affected source as error; other sources remain usable; session does not crash.)*
- What happens on very large files where analysis itself takes minutes? *(Resolved: show progress with user-facing copy such as "Analyzing file…"; allow cancel.)*
- What happens when event timestamps within a file are not strictly monotonic? *(Resolved: early stop after `To` is best-effort; document limitation in quickstart.)*
- What happens when multiple binlog files are opened? *(Resolved: single investigation scope applies to all; dialog shows aggregate metadata.)*
- What happens when approximate event count differs from final scoped count? *(Resolved: analysis count remains labeled approximate; final scoped count shown after indexing.)*
- What happens when the DBA selects Entire file on a large file? *(Resolved: secondary warning with explicit confirmation required before indexing.)*
- What happens when custom range extends beyond detected min/max? *(Resolved: validation error with clear message; scope not applied until corrected.)*
- What happens when the DBA opens an additional binlog mid-session while a scoped index exists? *(Resolved: analyze new file, merge metadata, scope dialog to confirm/adjust, re-index all sources with confirmed scope.)*
- What happens when CLI `--from`/`--to` is outside the detected file time span? *(Resolved: clear error; scoped indexing does not start.)*

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST run a file analysis pass on each opened binlog source before building the browse index.
- **FR-002**: Analysis MUST report per-source file size, minimum detected event timestamp, maximum detected event timestamp, and approximate user-data change-event count using the same inclusion rules as feature 002 (INSERT, UPDATE, DELETE, DDL; exclude housekeeping).
- **FR-003**: Analysis MUST NOT build browse summaries, decode row images, or retain event payloads in memory.
- **FR-004**: System MUST show analysis progress on large files and allow the DBA to cancel analysis to abort open.
- **FR-005**: System MUST cache analysis metadata per source for the session and reuse it when investigation scope changes without closing the source.
- **FR-006**: System MUST present a mandatory investigation scope dialog after analysis when scope is not provided via CLI.
- **FR-007**: Investigation scope options MUST include Entire file, Last hour (of file), Last day (of file), and Custom range with inclusive From/To boundaries.
- **FR-008**: System MUST require explicit secondary confirmation before indexing Entire file when configurable large-file thresholds are exceeded.
- **FR-009**: System MUST apply investigation scope at parse time: index only events within `[From, To]` inclusive; skip events before `From`; stop parsing a source after `To` when timestamps are monotonic (best-effort).
- **FR-010**: System MUST NOT populate the browse list until investigation scope is confirmed.
- **FR-011**: System MUST support `--from` and `--to` CLI flags with inclusive boundaries; both flags MUST be provided together when using CLI scope, or neither for dialog-based scope.
- **FR-011a**: When CLI `--from`/`--to` falls outside the merged detected min/max timestamp span after analysis, System MUST report a clear error and MUST NOT begin scoped indexing.
- **FR-012**: System MUST skip the scope dialog when both CLI date boundaries are provided at launch with valid binlog path(s).
- **FR-013**: System MUST allow changing investigation scope during an active session via key `s` (investigation scope dialog), separate from secondary filters (key `f`).
- **FR-013a**: When an additional binlog is opened in-session while a scoped index exists, System MUST analyze the new source, merge analysis metadata, present the scope dialog to confirm or adjust scope, and re-index all open sources with the confirmed scope (replacing the prior index).
- **FR-014**: Changing investigation scope MUST replace the in-memory browse index for affected sources; widening scope or selecting Entire file MUST trigger re-parse.
- **FR-015**: System MUST preserve feature 002 behaviors for browse, on-demand detail, multi-source chronological merge, housekeeping exclusion, row/statement format support, and inclusive time boundaries within the active scope.
- **FR-016**: Secondary filters (schema, table, operation type) MUST apply in memory over the scoped index without binlog re-parse.
- **FR-017**: Status display MUST distinguish scoped totals from secondary-filtered counts (`shown / scoped-total`).
- **FR-018**: System MUST report clear, actionable errors for invalid files, invalid scope input, partial CLI open failures, and mid-read parse failures.
- **FR-019**: When Entire file is selected, System SHOULD avoid a redundant full-file read where analysis and indexing can be combined in a single pass (optimization documented in plan).

### Key Entities

- **Binlog Source**: A single MySQL binary log file opened for exploration; unchanged from feature 002 except for additional analysis and scope state.
- **File Analysis Result**: Per-source metadata from the analysis pass: file size, min timestamp, max timestamp, approximate change-event count, and analysis completion status.
- **Investigation Scope**: The active inclusive time window `[From, To]` governing what events are indexed; set at open (dialog or CLI) or changed in-session; applies to all open sources in the session.
- **Event Summary / Event Index / Event Detail**: Unchanged in role from feature 002; index contains only events within the active investigation scope.
- **Secondary Filter Criteria**: Schema, table, and operation constraints applied in memory over the scoped index; distinct from investigation scope changes.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: On a 1 GB binlog where the DBA selects a one-day scope covering less than 5% of the file's detected time span, first scoped index events appear within 30 seconds and scoped indexing completes at least 50% faster than full-file indexing of the same file on representative hardware.
- **SC-002**: Memory use during scoped indexing for a narrow range remains bounded by scoped event count, not total file event count (no full-file summary accumulation).
- **SC-003**: When opening a multi-GB file without CLI scope, the DBA always sees analysis summary and scope dialog before heavy indexing begins.
- **SC-004**: When selecting Entire file on a file exceeding large-file thresholds, the DBA always sees an explicit warning with size, detected span, and approximate count before indexing proceeds.
- **SC-005**: Changing schema, table, or operation filters within a fixed investigation scope remains responsive (under 1 second for 10,000 scoped events).
- **SC-006**: A DBA can answer "what happened on table X on day D?" using scope selection, browse, secondary filters, and inspect without external tools, in under 5 minutes on a familiar binlog sample.

## Assumptions

- Feature 002 (`002-explore-binlog-events`) is implemented and provides the baseline browse, inspect, and filter TUI.
- Target users are experienced DBAs running the tool locally on database servers (SSH workflow).
- Binlog files are locally accessible; remote/streaming access remains out of scope.
- Event timestamps within a single binlog file are mostly chronological (MySQL/MariaDB binlog v4).
- "Last hour" and "Last day" presets are relative to the maximum timestamp detected in the file, not wall-clock time.
- Large-file warning thresholds (size and/or approximate event count) are defined during planning, not in this specification.
- Persistent on-disk index cache, virtual scroll windows, jump-to-end navigation, and binlog deduplication remain out of scope.
- Analysis pass may require reading the entire file to determine max timestamp and approximate counts; it remains materially lighter than full scoped indexing (no summary accumulation, no row decode).
- Validation is manual with real binlog datasets per project constitution.

## Dependencies

- **Feature 002**: Provides core exploration TUI, event types, housekeeping exclusion rules, and on-demand detail semantics extended by this feature.
- **Project constitution**: Go-native TUI, DBA-first workflows, manual validation, incremental delivery.

## Out of Scope

- SQLite or other persistent index sidecars across sessions.
- Virtual/infinite scroll unrelated to investigation scope.
- Remote or streaming binlog sources.
- Jump-to-end or arbitrary-offset navigation optimizations.
- Exact (non-approximate) event counts during analysis pass.
- Deduplication of overlapping binlog files from the same server sequence.
- Replacing secondary browse, inspect, or keyboard navigation patterns from feature 002 except where investigation scope supersedes post-index date filtering.
