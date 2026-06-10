# Feature Specification: Open and Explore Database Change Events

**Feature Branch**: `002-explore-binlog-events`

**Created**: 2025-06-10

**Status**: Clarified

**Input**: User description: "Provide an interactive exploration experience that allows a DBA to open one or more MySQL binary logs and investigate events quickly through browsing, inspection, and filtering."

## Clarifications

### Session 2026-06-10

- Q: Which binlog event types should appear in the primary browsable event list? → A: User-data change events only (INSERT, UPDATE, DELETE, DDL); exclude housekeeping events (Rotate, Format_description, GTID, etc.)
- Q: Which MySQL binlog replication event formats must be supported in this iteration? → A: Row-based and statement-based binlog events
- Q: How should the DBA provide binlog file paths to the application? → A: Both CLI paths at launch and in-session file open from the TUI
- Q: How should the application handle large binlog files for browsing and inspection? → A: Stream-parse with lightweight index; full detail loaded on event selection
- Q: How should time-range filter boundaries treat events at the exact start and end timestamps? → A: Inclusive start and end (events at boundary timestamps are included)

## User Scenarios & Manual Validation *(mandatory)*

### User Story 1 - Open Binlogs (Priority: P1)

As a DBA, I want to open one or more MySQL binary log files so that I can inspect their contents from a single interface without chaining multiple command-line tools.

**Why this priority**: No exploration is possible until binlog files are loaded. This is the entry point for every investigation workflow.

**Independent Validation**: Launch the application, open two known binlog files from a test environment, and confirm both are accepted and represented as available sources for exploration.

**Acceptance Scenarios**:

1. **Given** the application is started with one or more binlog file paths as CLI arguments, **When** launch completes, **Then** those files are loaded and their events are available for exploration.
2. **Given** the application is running, **When** the DBA opens an additional valid binlog file from within the TUI, **Then** the file is loaded and its events join the current exploration session.
3. **Given** one binlog file is already open, **When** the DBA adds a second valid binlog file (via CLI at launch or in-session open), **Then** events from both files are available within the same exploration session.
4. **Given** the DBA provides an invalid or unreadable file (via CLI or in-session open), **When** open is attempted, **Then** the application reports a clear error without crashing and allows the DBA to try another file.

---

### User Story 2 - Browse Events (Priority: P2)

As a DBA, I want to see events presented in chronological order so that I can understand the sequence of database activity.

**Why this priority**: Chronological browsing is the core exploration experience and delivers immediate value over raw mysqlbinlog output.

**Independent Validation**: Open a binlog with known event ordering, scroll through the event list, and confirm events appear in time sequence with enough summary information to distinguish event types without reading raw dump text.

**Acceptance Scenarios**:

1. **Given** one or more binlog files are open, **When** the DBA views the event list, **Then** user-data change events (INSERT, UPDATE, DELETE, DDL) are shown in chronological order across all open sources; housekeeping events are excluded from the list.
2. **Given** a long event list backed by a lightweight index, **When** the DBA navigates forward and backward, **Then** navigation remains responsive and the current position in the list is always visible without waiting for full event payloads.
3. **Given** events from multiple open binlog files, **When** displayed together, **Then** the combined list preserves chronological ordering.

---

### User Story 3 - Inspect Event Details (Priority: P3)

As a DBA, I want to select an event and view additional information so that I can better understand what happened during that change.

**Why this priority**: Summary lists answer "what happened when"; detail inspection answers "what exactly changed" for a specific event.

**Independent Validation**: Select a known row-modification event in a sample binlog and verify the detail view shows identifying information (timestamp, operation type, affected object) beyond the list summary.

**Acceptance Scenarios**:

1. **Given** an event list is visible, **When** the DBA selects an event, **Then** the application loads and displays full detail for that event on demand in a detail view.
2. **Given** the detail view is open, **When** the DBA selects a different event in the list, **Then** the detail view updates to reflect the newly selected event.
3. **Given** an event with limited parseable metadata, **When** inspected, **Then** the application shows all available detail and indicates when information is incomplete rather than failing silently.

---

### User Story 4 - Filter Events (Priority: P4)

As a DBA, I want to filter events by common criteria so that I can focus on relevant activity and ignore noise during an investigation.

**Why this priority**: Filtering reduces cognitive load on large binlogs but depends on browse and inspect capabilities already being in place.

**Independent Validation**: Open a binlog with mixed table activity, apply a filter for a specific table name, and confirm only matching events remain visible in the list.

**Acceptance Scenarios**:

1. **Given** an event list is loaded, **When** the DBA applies a time-range filter, **Then** only events with timestamps within that period are shown, inclusive of events occurring exactly at the start and end boundaries.
2. **Given** an event list is loaded, **When** the DBA filters by affected table (schema and/or table name), **Then** only events referencing that object are shown.
3. **Given** an event list is loaded, **When** the DBA filters by operation type (e.g., insert, update, delete, DDL), **Then** only events of the selected types are shown.
4. **Given** one or more filters are active, **When** the DBA clears filters, **Then** the full unfiltered event list is restored.

---

### User Story 5 - Investigate Activity (Priority: P5)

As a DBA, I want to move quickly between filtered results and event details so that I can answer operational questions efficiently during troubleshooting.

**Why this priority**: This story integrates browse, inspect, and filter into a cohesive investigation workflow—the primary value proposition of the product.

**Independent Validation**: Perform a realistic investigation: open a binlog, filter to a suspect time window and table, step through results, inspect details for each candidate event, and answer "what happened during this period on this table?" without leaving the application or using external tools.

**Acceptance Scenarios**:

1. **Given** filters are applied, **When** the DBA navigates the filtered list and selects events, **Then** detail inspection works on filtered results without requiring filter removal.
2. **Given** an active investigation, **When** the DBA adjusts filters, **Then** the event list and selection state update predictably (current selection moves to the nearest valid event or clears with clear indication).
3. **Given** a typical troubleshooting question ("What happened to table X between time A and B?"), **When** the DBA uses filter, browse, and inspect together, **Then** the question can be answered within a single session.

---

### Edge Cases

- What happens when a binlog file is empty or contains no user-data change events (only housekeeping events)?
- How does the system handle very large binlog files (millions of events) without becoming unusable? *(Resolved: stream-parse with lightweight index; detail on selection.)*
- What happens when binlog files use different format versions or character sets?
- How does the application present statement-based events where row-level before/after data is not available?
- What happens when a single binlog file contains a mix of row-based and statement-based events?
- How does the system behave when one of multiple open files fails mid-read?
- What happens when filters match zero events?
- How are events at the boundary of a time-range filter handled (inclusive vs exclusive)? *(Resolved: inclusive on both start and end timestamps.)*
- What happens when the DBA opens duplicate or overlapping binlog files from the same server sequence?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept one or more binlog file paths as CLI arguments at application launch.
- **FR-002**: System MUST allow the DBA to open additional MySQL binary log files from within the TUI during an active session.
- **FR-002a**: System MUST allow multiple binlog files to be open within the same session (via launch arguments, in-session open, or both).
- **FR-003**: System MUST present user-data change events (INSERT, UPDATE, DELETE, DDL) from all open sources in a single chronologically ordered list.
- **FR-003a**: System MUST exclude binlog housekeeping events (e.g., Rotate, Format_description, GTID, Previous_gtids, Xid) from the browsable event list.
- **FR-004**: System MUST display, for each event in the list, enough summary information to distinguish event type, approximate timing, and affected objects where available.
- **FR-005**: System MUST support efficient keyboard-driven navigation through the event list.
- **FR-006**: System MUST allow the DBA to select an event and view expanded details in a dedicated detail area.
- **FR-007**: System MUST update the detail view when the DBA selects a different event.
- **FR-008**: System MUST support filtering events by time range, inclusive of events at the exact start and end boundary timestamps.
- **FR-009**: System MUST support filtering events by affected table (schema and/or table identifier).
- **FR-010**: System MUST support filtering events by operation type.
- **FR-011**: System MUST allow the DBA to clear active filters and return to the full event list.
- **FR-012**: System MUST support applying filters while browsing and inspecting without restarting the session.
- **FR-013**: System MUST report clear, actionable errors when a file cannot be opened or read.
- **FR-014**: System MUST indicate when event metadata is partial or unavailable rather than omitting the event without explanation.
- **FR-015**: System MUST NOT require the DBA to use external command-line tools to complete basic open-browse-filter-inspect workflows defined in this specification.
- **FR-016**: System MUST parse and present user-data change events from both row-based and statement-based binlog formats.
- **FR-017**: System MUST indicate in event detail when row-level data is unavailable (e.g., statement-based events without row images).
- **FR-018**: System MUST stream-parse binlog files and maintain a lightweight index (timestamp, operation type, affected objects, source reference) sufficient for browse and filter without loading full event payloads into memory.
- **FR-019**: System MUST load full event detail on demand when the DBA selects an event from the list.

### Key Entities

- **Binlog Source**: A single MySQL binary log file opened for exploration; identified by its file path and contributing events to the session.
- **Change Event**: A user-data change record (INSERT, UPDATE, DELETE, or DDL) extracted from a binlog source in row-based or statement-based format; has a timestamp, operation type, optional affected schema/table, replication format indicator, and source reference. Housekeeping binlog events are parsed internally but not surfaced in the browsable list.
- **Event Summary**: The compact, index-backed representation of a change event shown in the browsable list (metadata only).
- **Event Detail**: The expanded representation of a selected change event, loaded on demand from the binlog source, including all parseable attributes relevant to investigation.
- **Event Index**: Lightweight per-event metadata built during stream-parse to support chronological browse and filtering without retaining full payloads in memory.
- **Filter Criteria**: User-specified constraints (time range, table, operation type) applied to narrow the visible event set. Time-range filters use inclusive start and end boundaries.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A DBA can open one or more binlog files and see the first indexed events within 30 seconds for binlog files up to 1 GB on typical DBA workstation hardware, without requiring a full in-memory parse to complete first.
- **SC-002**: A DBA can answer "what events occurred on table X between time A and B?" using only in-application browse, filter, and inspect—without running mysqlbinlog or grep— in under 5 minutes on a familiar binlog sample.
- **SC-003**: A DBA can reduce a list of 10,000+ events to a focused subset using at least one filter criterion in under 1 minute.
- **SC-004**: In a side-by-side comparison with a traditional mysqlbinlog plus shell-filtering workflow, the DBA completes a standard investigation scenario (locate modifications to a specific table in a time window) at least 40% faster using this application.
- **SC-005**: 100% of acceptance scenarios in User Stories 1–5 pass when validated manually against at least two real-world binlog samples (one small, one large).

## Assumptions

- Target users are experienced DBAs already familiar with MySQL binary logs, transactions, and troubleshooting workflows.
- Binlog files are locally accessible files on the DBA's machine; remote or streaming binlog sources are out of scope for this iteration.
- Binlog files may be provided at launch via CLI arguments and/or opened in-session from the TUI; both mechanisms are required in this iteration.
- Input files are standard MySQL binary log format produced by supported MySQL versions in the DBA's environment.
- Both row-based and statement-based replication events are in scope for INSERT, UPDATE, DELETE, and DDL; mixed-format binlogs are supported by parsing whichever event type appears in the file.
- Validation is performed manually with real binlog datasets per project constitution; no automated test suites are required for this feature.
- SQL reconstruction, before/after visual diff, transaction grouping, statistics dashboards, activity timelines, export, alerting, reporting, and non-MySQL data sources are explicitly out of scope.
- "Basic filtering" in this iteration means time range, table/schema, and operation type; free-text search and advanced compound filter expressions are deferred to future specifications unless needed during planning.
- Multiple open binlog files are merged into one chronological view; per-file isolation views are not required in this iteration.
- The browsable event list includes only user-data change events (INSERT, UPDATE, DELETE, DDL); housekeeping binlog events are excluded from exploration views in this iteration.
