<!--
Sync Impact Report
Version change: [template/unversioned] → 1.0.0
Modified principles: N/A (initial ratification)
Added sections: Purpose, Architectural Decision Rule, 12 Core Principles, Governance
Removed sections: Template placeholder sections (SECTION_2, SECTION_3)
Templates updated:
  ✅ .specify/templates/plan-template.md
  ✅ .specify/templates/spec-template.md
  ✅ .specify/templates/tasks-template.md
  ✅ .cursor/rules/specify-rules.mdc
Follow-up TODOs: None
-->

# MySQL Binlog Explorer Constitution

## Purpose

MySQL Binlog Explorer is a personal and internal tool designed for database
administrators to analyze, investigate, and explore MySQL binary logs more
efficiently than with mysqlbinlog.

The goal is not to create a graphical wrapper around mysqlbinlog.

The goal is to create a database change explorer focused on:

- Troubleshooting
- Incident investigation
- Auditing
- Forensic analysis
- Change tracking
- Transaction exploration

The initial implementation targets MySQL binary logs exclusively.

Future support for additional database change sources (such as Oracle LogMiner
output) may be considered, but MUST NOT influence current implementation
decisions unless explicitly requested.

The conceptual inspiration is **"Wireshark for database change events"** rather
than **"mysqlbinlog with colors"**.

## Core Principles

### I. DBA First

This application is designed for experienced DBAs and infrastructure engineers.

The project MUST prioritize:

- Productivity over aesthetics
- Keyboard navigation over mouse interactions
- Information density over visual simplicity
- Fast access to data over visual effects

The application is not intended for non-technical users.

**Rationale**: The primary users are skilled operators who need speed and depth
during incidents and audits, not simplified consumer interfaces.

### II. Go Native

The entire project MUST remain in Go.

Approved technologies:

- Go
- Bubble Tea
- Bubbles
- Lip Gloss

The project MUST NOT introduce the following unless explicitly requested:

- React, Angular, Vue
- Electron
- NodeJS runtime dependencies
- Browser-based frontends
- Client/server architectures

The preferred distribution model is a single self-contained executable.

**Rationale**: A single Go binary keeps deployment simple for DBA workflows and
avoids operational overhead from multi-runtime stacks.

### III. TUI First

The application MUST be implemented as a terminal user interface from the
beginning.

The project MUST NOT build a temporary CLI-only application as an intermediate
step.

The project MUST evolve incrementally through working TUI features.

The UI MAY start simple and progressively evolve toward a Wireshark/LazyGit
style experience.

**Rationale**: Exploration workflows benefit from persistent, navigable views
that a TUI provides from day one.

### IV. Explorer, Not Viewer

The application MUST prioritize exploration and investigation capabilities
rather than raw log rendering.

Preferred features answer questions such as:

- What changed?
- Who changed it?
- When did it happen?
- Which objects were affected?
- Which transaction performed the change?

Features that only display raw source output are secondary.

**Rationale**: DBAs need structured answers during investigations, not another
log dump viewer.

### V. Incremental Development

The application MUST be built through small, working vertical slices.

Prefer:

- parser + model + UI + basic interaction

over:

- building large subsystems in isolation

Every iteration MUST result in a usable application.

Advanced features MUST NOT be implemented before the underlying workflow has
been validated.

**Rationale**: Vertical slices prove value early and reduce the risk of building
unused infrastructure.

### VI. Simplicity Over Engineering

The project MUST favor the simplest solution that satisfies the current
requirements.

Avoid:

- Premature optimization
- Unnecessary abstractions
- Enterprise patterns without clear value
- Excessive interfaces
- Architecture designed for hypothetical future requirements

The project is a practical DBA tool, not a framework.

**Rationale**: Complexity slows iteration and makes incident-time debugging
harder for the target audience.

### VII. Performance Awareness

Input files MAY be large. Designs MUST assume large binlog files.

Prefer:

- Efficient memory usage
- Streaming where practical
- Responsive navigation
- Scalable parsing approaches

Avoid designs that unnecessarily load or duplicate large amounts of data.

Simplicity MUST NOT be sacrificed for micro-optimizations unless performance
becomes a proven problem.

**Rationale**: Binlogs can reach gigabytes; responsiveness is part of DBA
productivity.

### VIII. Maintainability

Generated code MUST be:

- Readable
- Explicit
- Easy to debug
- Easy to modify

Prefer clear code over clever code. Avoid unnecessary complexity.

Future contributors MUST be able to understand the codebase quickly.

**Rationale**: A maintainable codebase supports long-lived internal tooling with
minimal bus factor risk.

### IX. User Experience Consistency

Navigation and interaction patterns MUST remain consistent throughout the
application.

Prefer:

- Predictable keyboard shortcuts
- Stable layouts
- Consistent filtering behavior
- Consistent search behavior

New features MUST integrate naturally with existing workflows.

**Rationale**: Consistency reduces cognitive load during high-pressure
investigations.

### X. Pragmatic Extensibility

The current scope is MySQL binary logs.

The project MUST NOT introduce generic frameworks, plugin systems, parser
factories, or speculative abstractions for future data sources.

When naming packages, modules, types, and internal concepts, prefer terminology
that reflects database change exploration rather than MySQL-specific
implementation details when doing so does not increase complexity.

Preferred naming examples:

- `events`, `explorer`, `search`, `filters`, `sources/mysql`

Avoid when unnecessary:

- `mysql_everything`, `binlog_everything`

Future extensibility MUST emerge naturally from good design, not from
speculative architecture.

**Rationale**: Scope discipline prevents over-engineering while sensible naming
preserves optionality without upfront cost.

### XI. Spec-Driven Development

The project MUST use SpecKit as the development methodology.

Major functionality MUST be introduced through specifications.

Specifications MUST remain practical and lightweight.

Unnecessary process overhead MUST be avoided.

**Rationale**: Specs improve clarity and guide AI-assisted development without
replacing working software.

### XII. No Automated Testing

This project does NOT use automated testing frameworks.

The project MUST NOT create:

- Unit tests
- Integration tests
- End-to-end tests
- Test suites
- Testing infrastructure

Validation MUST be performed manually using real-world datasets and real
application workflows.

Development effort MUST focus on delivering functionality rather than
maintaining automated tests.

**Rationale**: Manual validation against real binlogs matches DBA workflows and
keeps focus on exploration features.

## Architectural Decision Rule

When multiple valid solutions exist, prefer the solution that:

1. Keeps the project entirely in Go
2. Preserves the single executable model
3. Minimizes complexity
4. Improves DBA productivity
5. Supports investigation and exploration workflows
6. Fits naturally within a TUI application
7. Can be implemented incrementally

When in doubt, choose the simpler solution.

## Governance

This constitution supersedes conflicting ad-hoc practices for MySQL Binlog
Explorer.

Amendments MUST:

- Be documented in `.specify/memory/constitution.md`
- Include a Sync Impact Report comment describing the change
- Increment `CONSTITUTION_VERSION` according to semantic versioning:
  - **MAJOR**: Backward incompatible governance or principle removals/redefinitions
  - **MINOR**: New principles/sections or materially expanded guidance
  - **PATCH**: Clarifications, wording, typo fixes, non-semantic refinements
- Update dependent templates when principles affect planning, specification, or
  task generation

Compliance review expectations:

- `/speckit-plan` Constitution Check gates MUST reference active principles
- Feature specs MUST align with DBA-first, TUI-first, and explorer-not-viewer goals
- Implementation plans MUST NOT introduce prohibited technologies or automated
  testing unless the constitution is amended first
- Complexity beyond these principles MUST be justified in plan Complexity Tracking

Runtime development guidance: follow the active feature plan at
`specs/[###-feature-name]/plan.md` and this constitution.

**Version**: 1.0.0 | **Ratified**: 2025-06-10 | **Last Amended**: 2025-06-10
