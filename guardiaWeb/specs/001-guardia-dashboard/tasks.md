---
description: "Task list for Dashboard de guardias"
---

# Tasks: Dashboard de guardias

**Input**: Design documents from `/specs/001-guardia-dashboard/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/schedule-api.md

**Tests**: Manual verification only (constitution). Use [quickstart.md](./quickstart.md) at each checkpoint.

**Organization**: Tasks grouped by user story for independent implementation and manual testing.

## Format: `[ID] [P?] [Story] Description`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Initialize Next.js project and dependencies

- [x] T001 Scaffold Next.js at repo root **without overwriting Spec Kit artifacts**: before `npx create-next-app@latest .`, confirm `.specify/`, `specs/`, `.gitignore`, and `GuardiaWeb2026.xlsx` will be preserved (use CLI prompts to skip overwriting, or add `app/`, `package.json`, `next.config.ts`, `tsconfig.json`, Tailwind files manually). TypeScript + App Router + Tailwind + ESLint; no `src/` directory
- [x] T002 Add runtime dependencies in `package.json`: `xlsx`, `@fullcalendar/react`, `@fullcalendar/core`, `@fullcalendar/daygrid`, `@fullcalendar/multimonth`, `@fullcalendar/interaction`
- [x] T003 [P] Initialize shadcn/ui (`components.json`, `lib/utils.ts`, Tailwind config) per plan.md
- [x] T004 [P] Add shadcn components: `button`, `select`, `alert`, `card` under `components/ui/`
- [x] T005 Create `.env.local.example` with `GUARDIA_EXCEL_PATH=./GuardiaWeb2026.xlsx` and document copy to `.env.local` in project README stub

**Checkpoint**: `npm run dev` starts without errors (placeholder page OK); `specs/` and `.specify/` unchanged

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Excel parse pipeline and shared libs — MUST complete before user stories

**⚠️ CRITICAL**: No user story work until this phase is complete

- [x] T007 Define `GuardPeriod` and `ScheduleData` types in `lib/excel/types.ts` per data-model.md
- [x] T008 Implement `DD-MM-YY` and Excel serial date parsing in `lib/excel/parse-date.ts` (century rule 00–69 / 70–99 from research.md)
- [x] T009 Implement `parseScheduleFromWorkbook` in `lib/excel/parse-schedule.ts` (headers B1–I1 normalized: trim + match `M,J,V,S,D,L,M,Persona`; rows 2+, cols B–H dates, col I person, skip invalid rows, `skippedRowCount`, push overlap text into `warnings[]`)
- [x] T010 Implement `loadSchedule()` in `lib/schedule/load-schedule.ts` with `import 'server-only'` at top (read `GUARDIA_EXCEL_PATH`, call parser, handle FILE_NOT_FOUND / bad headers)
- [x] T011 Implement stable person colors in `lib/person-colors.ts`
- [x] T012 Implement `mapPeriodsToFullCalendarEvents` in `lib/calendar/map-to-events.ts` (all-day end +1 day, colors, `guard-active` class when today ∈ period)
- [x] T013 Implement `GET` handler in `app/api/schedule/route.ts` matching `contracts/schedule-api.md`
- [x] T014 Add helper `findActivePeriodsForDate(periods, date)` in `lib/schedule/active-period.ts` for today banner

**Checkpoint**: Call `loadSchedule()` manually or hit `/api/schedule` — verify Gerardo/Diego sample rows from spec (24-12-25 week, etc.)

---

## Phase 3: User Story 1 — Consultar quién está de guardia hoy (Priority: P1) 🎯 MVP

**Goal**: Monthly calendar, today’s guard banner, empty states — core daily use case

**Scope note**: US1 acceptance **scenario 3** (refresh tras editar planilla) se completa en **Phase 6** (T033–T035). Tras esta fase, US1 está **parcial** (escenarios 1–2).

**Independent Test**: Open app on a date with assigned guard → see monthly view, person name, start/end dates, today highlighted; no Excel open required

### Implementation for User Story 1

- [x] T015 [US1] Create `app/layout.tsx` with clean Spanish `lang`, minimal shell, Tailwind globals
- [x] T016 [P] [US1] Create `components/today-guard-banner.tsx` using `findActivePeriodsForDate`: show one or **all** active periods; if 2+, add short overlap note (spec edge case)
- [x] T017 [P] [US1] Create `components/schedule-calendar.tsx` client component with FullCalendar `dayGridMonth` as `initialView`
- [x] T018 [US1] Create `components/schedule-dashboard.tsx` client shell (banner + calendar props, no filter yet)
- [x] T019 [US1] Wire `app/page.tsx` as Server Component: `loadSchedule()`, pass serialized periods/people/skippedRowCount to `ScheduleDashboard`
- [x] T020 [US1] Add error UI in `app/page.tsx` for missing file / parse failure (Spanish messages, no stack traces)
- [x] T021 [US1] Style `.guard-active` events in `app/globals.css` for today’s period highlight (FR-006)
- [x] T022 [US1] Handle empty `periods` array with friendly empty state in `components/schedule-dashboard.tsx`

**Checkpoint (MVP US1 parcial)**: US1 scenarios **1–2** from spec.md pass via [quickstart.md](./quickstart.md) P1; scenario 3 deferred to Phase 6

---

## Phase 4: User Story 2 — Navegar el calendario por período (Priority: P2)

**Goal**: Day, week, month, year views with Wednesday-based weeks

**Independent Test**: Switch views, navigate to known future week, confirm person and date range on events

### Implementation for User Story 2

- [x] T023 [P] [US2] Create `components/schedule-toolbar.tsx` with view buttons: día / semana / mes / año
- [x] T024 [US2] Extend `components/schedule-calendar.tsx` to switch `initialView` / `changeView`: `dayGridDay`, `dayGridWeek`, `dayGridMonth`, `multiMonthYear`
- [x] T025 [US2] Set FullCalendar `firstDay: 3` (Wednesday) and locale-friendly week headers in `components/schedule-calendar.tsx`
- [x] T026 [US2] Integrate `ScheduleToolbar` into `components/schedule-dashboard.tsx` with `calendarView` state
- [x] T027 [US2] Enable date click / `datesSet` to show selected date context (optional subtitle or banner update)
- [x] T028 [US2] Render overlapping periods both visible; overlap detection stays in `lib/excel/parse-schedule.ts` (`warnings[]`)
- [x] T028a [US2] Show discrete shadcn `Alert` in `components/schedule-dashboard.tsx` when `warnings.length > 0` (e.g. “Hay solapes en la planilla”)

**Checkpoint**: US2 acceptance scenarios pass (quickstart P2)

---

## Phase 5: User Story 3 — Filtrar y distinguir personas (Priority: P3)

**Goal**: Person dropdown filter and consistent colors

**Independent Test**: Select person → only their events; select «Todos» → full calendar; colors consistent

### Implementation for User Story 3

- [x] T029 [P] [US3] Add person `Select` to `components/schedule-toolbar.tsx` (options: «Todos» + `people` from props)
- [x] T030 [US3] Filter FullCalendar events client-side in `components/schedule-dashboard.tsx` via `useMemo` when `selectedPerson` set
- [x] T031 [US3] Add person color legend component `components/person-legend.tsx` using `lib/person-colors.ts`
- [x] T032 [US3] Mount legend in `components/schedule-dashboard.tsx` when ≥2 people

**Checkpoint**: US3 acceptance scenarios pass (quickstart P3)

---

## Phase 6: Refresh & partial errors (FR-010, FR-011) — completes US1 scenario 3

**Goal**: Explicit refresh and skipped-row alert; **US1 queda completo** tras esta fase

**Independent Test**: Edit Excel (closed), click «Actualizar» or F5 → new data; invalid row → partial calendar + count alert

- [x] T033 Add «Actualizar» button in `components/schedule-toolbar.tsx` calling `router.refresh()` from `next/navigation`
- [x] T034 Create `components/skipped-rows-alert.tsx` using shadcn `Alert` when `skippedRowCount > 0`
- [x] T035 Wire alert and refresh in `components/schedule-dashboard.tsx` / `app/page.tsx`

**Checkpoint**: FR-010, FR-011; US1 scenario 3; full US1 sign-off (quickstart refresh section)

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Responsive layout, docs, performance sanity check

- [x] T036 [P] Basic responsive layout in `components/schedule-dashboard.tsx` and toolbar (stack on narrow widths per plan.md)
- [x] T037 [P] Add root `README.md` with Node prerequisites, `npm install`, `GUARDIA_EXCEL_PATH`, `npm run dev`
- [ ] T038 Run full [quickstart.md](./quickstart.md) checklist and fix gaps (pending: `npm install` + manual run)
- [x] T039 Verify `xlsx` is not imported from any `'use client'` file (server-only parse)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)** → **Foundational (Phase 2)** → **US1 (Phase 3)** → **US2 (Phase 4)** → **US3 (Phase 5)** → **Phase 6** → **Polish (Phase 7)**
- Phase 6 can start after US1 but needs toolbar from US2 for button placement — **recommended after Phase 4**

### User Story Dependencies

| Story | Depends on | Can test alone after |
|-------|------------|----------------------|
| US1 (P1) | Phase 2 | Phase 3 complete |
| US2 (P2) | US1 calendar shell | Phase 4 complete |
| US3 (P3) | US2 toolbar | Phase 5 complete |

### Within Each User Story

- Types/parser before UI
- Server `page.tsx` before client dashboard
- Calendar before toolbar extensions

### Parallel Opportunities

- **Phase 1**: T003 ∥ T004
- **Phase 3**: T016 ∥ T017
- **Phase 4**: T023 ∥ (after T024 started) partial
- **Phase 5**: T029 ∥ T031
- **Phase 7**: T036 ∥ T037

---

## Parallel Example: User Story 1

```bash
# After T015 starts, in parallel:
# Task T016: components/today-guard-banner.tsx
# Task T017: components/schedule-calendar.tsx
# Then T018 integrates both in schedule-dashboard.tsx
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup  
2. Complete Phase 2: Foundational (**blocking**)  
3. Complete Phase 3: User Story 1  
4. **STOP and VALIDATE** with quickstart P1  
5. Demo to team before P2/P3  

### Incremental Delivery

1. Setup + Foundational → parser works  
2. US1 → daily “who is on call” works (MVP)  
3. US2 → full calendar navigation  
4. US3 → person filter + legend  
5. Phase 6 → refresh + error count  
6. Polish → README + responsive  

### Suggested MVP Scope

**Phases 1–3 (T001–T022)** — parse Excel + monthly view + today banner.

Defer to next iteration: view switcher (US2), person filter (US3), refresh button (Phase 6). US1 scenario 3 also Phase 6 — F5 works until T033.

---

## Task Summary

| Phase | Task IDs | Count |
|-------|----------|-------|
| Setup | T001–T005 | 5 |
| Foundational | T007–T014 | 8 |
| US1 (P1) | T015–T022 | 8 |
| US2 (P2) | T023–T028, T028a | 7 |
| US3 (P3) | T029–T032 | 4 |
| Refresh/errors | T033–T035 | 3 |
| Polish | T036–T039 | 4 |
| **Total** | **39 tasks** (T028a added; T006 merged into T010) | **39** |

**Format validation**: All tasks use `- [ ]`, sequential `T###` IDs, `[P]` only when parallel-safe, `[USn]` on story phases only, and include file paths.
