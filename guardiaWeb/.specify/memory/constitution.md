<!--
Sync Impact Report
- Version change: template placeholders → 1.0.0 (initial project-specific constitution)
- Principles: replaced generic template slots with five GuardiaWeb principles (titles new)
- Added sections: Technical Standards; Scope, Quality & Verification (replacing generic SECTION_2/3 placeholders)
- Removed sections: none (template commentary removed)
- Templates: .specify/templates/plan-template.md ✅ | spec-template.md ✅ | tasks-template.md ✅ | .specify/templates/commands/*.md — not present in this repo (N/A)
- Follow-up TODOs: none
-->

# GuardiaWeb Constitution

## Core Principles

### I. Simplicity Over Ceremony

The codebase MUST favor the smallest solution that works. Contributors MUST NOT add
microservices, premature abstraction, extra wrapper layers, or patterns justified only
by hypothetical future scale. New complexity MUST solve a concrete problem in this
repository or product, and MUST be called out in the feature plan when introduced.

### II. Readability and Focused Modules

Code MUST be written for the next reader: clear names, shallow functions, and small
components with one obvious job. Files SHOULD stay small enough to understand in one
sitting; splits MUST improve clarity rather than scatter related logic.

### III. Incremental Delivery

Work MUST land in small vertical slices that run locally end-to-end. Large one-shot
generations or sweeping refactors are PROHIBITED unless unavoidable; when they are,
the plan MUST explain scope, risk, and rollback before implementation starts.

### IV. Calm, Cohesive UI

The interface MUST be predictable and easy to scan for internal operators. Visual
design MUST follow shadcn/ui conventions. Calendar views MUST use FullCalendar.
Layouts MUST stay uncluttered for a read-only schedule dashboard.

### V. Lean Stack and Honest Scope

Dependencies MUST stay minimal; each addition MUST have a clear payoff. The product
is a read-only internal dashboard: no authentication, no CRUD, no database unless
strictly required to ship, and no multi-service infrastructure.

## Technical Standards

- **Runtime**: Next.js, React, and TypeScript.
- **UI**: shadcn/ui for components and interaction patterns.
- **Calendar**: FullCalendar for schedule visualization.
- **Excel**: SheetJS (`xlsx`) to parse `.xlsx` sources.
- **Data flow**: Prefer server-side parsing (Route Handlers or Server Actions) or
  equivalent simple paths; avoid bespoke backend services for parsing alone.
- **Structure**: Keep folders modular and shallow; colocate UI with its parsing or
  formatting helpers when coupling is natural; avoid empty abstraction layers.

## Scope, Quality, and Verification

- **Product**: Read-only visualization of an on-call schedule from Excel for a trusted
  internal audience.
- **Code style**: Prefer self-explanatory code; document only non-obvious rules such as
  sheet layout assumptions or calendar edge cases.
- **Testing**: Rely on manual and functional checks while building. Automated unit tests
  MUST NOT be generated for this project unless a future constitution amendment
  explicitly enables them.
- **AI-assisted development**: Use short plans, tight prompts, and iterative steps.
  Explain meaningful architectural choices before executing them.

## Governance

This constitution overrides conflicting boilerplate from generic templates in this
repo. Amendments MUST bump `CONSTITUTION_VERSION` using semantic versioning, set
`LAST_AMENDED_DATE` to the change date in ISO `YYYY-MM-DD` form, and refresh the Sync
Impact Report comment. Material rule changes MUST update Spec Kit templates under
`.specify/templates/` so plans, specs, and tasks stay aligned.

**Version**: 1.0.0 | **Ratified**: 2026-05-14 | **Last Amended**: 2026-05-14
