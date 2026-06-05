# Implementation Plan: Dashboard de guardias

**Branch**: `001-guardia-dashboard` | **Date**: 2026-05-15 | **Spec**: [spec.md](./spec.md)

**Input**: Planilla Excel de guardias semanales (miércoles–martes), panel interno de solo
lectura con calendario (día/semana/mes/año), filtro por persona y destacado de guardia
activa hoy.

## Summary

GuardiaWeb es una aplicación **Next.js (App Router)** que lee un `.xlsx` en el servidor
con **SheetJS**, normaliza cada fila a un **período de guardia** (persona + inicio + fin)
y expone los datos como JSON a la UI. **FullCalendar** renderiza las vistas temporales;
**shadcn/ui** aporta controles (filtro, actualizar, avisos). Sin base de datos, sin auth,
sin CRUD. Estado de UI en React local; relectura de Excel solo al cargar/recargar o al
pulsar «Actualizar».

## Technical Context

**Language/Version**: TypeScript 5.x, Node 20 LTS (mínimo 18.18)

**Primary Dependencies**:

| Paquete | Rol |
|---------|-----|
| `next`, `react`, `react-dom` | App web, SSR/RSC, Route Handlers |
| `xlsx` (SheetJS) | Parseo del Excel en servidor |
| `@fullcalendar/react` + plugins `daygrid`, `multimonth`, `interaction` | Vistas día/semana/mes/año (`dayGridDay`, `dayGridWeek`, `dayGridMonth`, `multiMonthYear`; sin `timegrid`) |
| `tailwindcss`, `class-variance-authority`, `clsx`, `tailwind-merge` | Estilos (base shadcn) |
| Componentes shadcn: `button`, `select`, `alert`, `card` (según necesidad) | UI mínima |

**Storage**: Archivo Excel en disco; ruta vía variable de entorno `GUARDIA_EXCEL_PATH`
(ej. `./GuardiaWeb2026.xlsx` en la raíz del repo). Sin DB ni caché persistente en v1.

**Testing**: Verificación manual según [quickstart.md](./quickstart.md) y escenarios del
spec. Sin suites automatizadas (constitución).

**Target Platform**: Navegador desktop interno (Chrome/Edge); despliegue Node (Vercel,
servidor Windows/Linux o `next start` local).

**Project Type**: Web app monolito (frontend + Route Handler en un solo proyecto Next.js).

**Performance Goals**: Carga inicial y cambio de vista < 1 s percibido con ~52 filas
anuales; parseo Excel < 200 ms en servidor de oficina.

**Constraints**: Solo lectura; layout Excel fijo; fechas `DD-MM-YY`; semana inicia
miércoles; vista por defecto mensual; sin polling de archivo.

**Scale/Scope**: Un equipo técnico interno, una planilla por entorno, decenas de filas.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Estado |
|------|--------|
| Simplicity (sin microservicios/DB/auth/CRUD) | ✅ Pass |
| Stack Next + TS + shadcn + FullCalendar + SheetJS | ✅ Pass |
| Parseo en servidor (Route Handler) | ✅ Pass |
| Entrega incremental + verificación manual | ✅ Pass |
| UI clara, sin sobrecarga | ✅ Pass (diseño en componentes acotados) |

**Post-design (Phase 1)**: Sin violaciones. No se requiere tabla en Complexity Tracking.

## Architecture Overview

```text
┌─────────────────────────────────────────────────────────────┐
│  Browser                                                     │
│  ┌──────────────┐  ┌─────────────────┐  ┌────────────────┐ │
│  │ shadcn/ui    │  │ FullCalendar    │  │ React state    │ │
│  │ (filter,     │  │ (day/week/month │  │ view, person,  │ │
│  │  refresh,    │  │  /year)         │  │ selected date) │ │
│  │  alerts)     │  │                 │  │                │ │
│  └──────┬───────┘  └────────▲────────┘  └───────┬────────┘ │
│         │                  │ events JSON        │          │
│         │         fetch / router.refresh         │          │
└─────────┼──────────────────┼────────────────────┼──────────┘
          │                  │                    │
          ▼                  │                    │
┌─────────────────────────────────────────────────────────────┐
│  Next.js App Router (Node)                                   │
│  app/page.tsx (Server Component) ──► initial data optional   │
│  app/api/schedule/route.ts ◄── GET (refresh explícito)       │
│         │                                                    │
│         ▼                                                    │
│  lib/excel/parse-schedule.ts (SheetJS read file)               │
│  lib/calendar/to-fullcalendar-events.ts                      │
└──────────────────────────────┬──────────────────────────────┘
                               │ fs.readFile
                               ▼
                    GUARDIA_EXCEL_PATH (.xlsx)
```

**Decisiones clave**:

1. **Parseo solo en servidor** — `xlsx` no va al bundle del cliente; evita peso y expone
   menos superficie.
2. **Un Route Handler** `GET /api/schedule` — contrato JSON estable; la página puede
   cargar datos en el servidor (RSC) o el botón «Actualizar» hace `fetch` + actualiza
   estado (patrón cliente ligero). MVP: RSC en `page.tsx` que llama al parser directamente;
   «Actualizar» usa `router.refresh()` para re-ejecutar el servidor (simple, sin estado
   duplicado).
3. **Sin capa “repository”** — `parse-schedule.ts` devuelve `ScheduleData`; el mapper a
   FullCalendar es una función pura aparte.
4. **Sin Zustand/Redux** — filtro de persona y vista activa en `useState` del layout
   del dashboard; datos derivados con `useMemo`.

## Data Flow (Excel → UI)

1. **Lectura**: `fs.readFileSync` / `readFile` de `GUARDIA_EXCEL_PATH` (validar existe).
2. **Parseo SheetJS**: primera hoja; fila 1 cols B–I = encabezados esperados; filas 2+.
3. **Por fila**: leer 7 celdas fecha + `Persona`; parsear `DD-MM-YY` → `Date` (año 20xx);
   construir `GuardPeriod { id, person, start, end }`; acumular errores de fila.
4. **Post-proceso**: lista única de personas (trim + normalizar espacios); detectar
   solapes opcional (flag en metadata); asignar color estable por persona.
5. **Respuesta API / props**: `{ periods, people, skippedRowCount, warnings? }`.
6. **UI**: mapper → eventos FullCalendar (`title`, `start`, `end`, `backgroundColor`,
   `extendedProps`); filtro cliente oculta eventos por `person`; hoy → `eventClassNames`
   o banner superior con período activo.

## State Management

| Estado | Dónde | Notas |
|--------|-------|-------|
| Períodos parseados | Servidor (origen) | Releídos en cada request tras refresh |
| Vista calendario (day/week/month/year) | Cliente `useState` | Sincronizada con `initialView` de FullCalendar |
| Persona filtro (`Todos` \| nombre) | Cliente `useState` | Filtra eventos en memoria, no re-parsea |
| Fecha visible / seleccionada | FullCalendar API | `datesSet` callback |
| Errores parciales (`skippedRowCount`) | Props desde servidor | Alert shadcn discreto |

No hay estado global ni persistencia en `localStorage` en v1.

## Main Components

| Componente | Responsabilidad |
|------------|-----------------|
| `app/layout.tsx` | Shell, fuentes, tema claro |
| `app/page.tsx` | Server: invoca parser, pasa datos a `ScheduleDashboard` |
| `ScheduleDashboard` (client) | Orquesta toolbar + calendario + alertas |
| `TodayGuardBanner` | Muestra guardia activa hoy (FR-006, P1) |
| `ScheduleToolbar` | Select persona, botones vista, «Actualizar» |
| `ScheduleCalendar` | Wrapper FullCalendar + mapping eventos |
| `SkippedRowsAlert` | Aviso «N filas no cargadas» |
| `lib/excel/parse-schedule.ts` | Lógica de grilla Excel |
| `lib/calendar/map-to-events.ts` | `GuardPeriod[]` → FullCalendar `EventInput[]` |
| `lib/person-colors.ts` | Hash nombre → color HSL consistente |

## Responsive UI Strategy

- **Desktop (≥1024px)**: toolbar horizontal; calendario altura `calc(100vh - header)`.
- **Tablet (768–1023px)**: toolbar en dos filas; vistas mes/semana prioritarias.
- **Mobile (<768px)**: fuera de MVP crítico; vista **día** por defecto si ancho bajo
  (simplificación pragmática) o scroll horizontal en semana — documentado en Phase 5.
- FullCalendar `height="auto"` + contenedor con scroll; tipografía shadcn legible.

## Dependency Choices (rationale)

- **Next.js App Router**: un solo despliegue, Route Handlers para Excel, RSC para carga
  inicial rápida sin API expuesta si no hace falta.
- **FullCalendar**: vistas día/semana/mes/año ya resueltas; evita calendario custom.
- **SheetJS**: estándar de facto para `.xlsx`; solo servidor.
- **shadcn/ui**: copiar componentes, sin runtime pesado; alinea con constitución.
- **No** `date-fns` obligatorio en v1 — `Date` nativo + parser propio acotado; añadir
  `date-fns` solo si el parseo se vuelve frágil (ver research.md).

## Implementation Phases

### Phase 0 — Scaffold y parseo (MVP base)

- Scaffold Next.js en la raíz del repo **sin sobrescribir** `.specify/`, `specs/`,
  `.gitignore` ni la planilla local. Usar `create-next-app` con cuidado o archivos
  manuales (`app/`, `package.json`, configs). TypeScript, Tailwind, App Router; sin
  carpeta `src/`.
- Instalar `xlsx`, configurar `GUARDIA_EXCEL_PATH` en `.env.local`.
- Implementar `parse-schedule.ts` contra `GuardiaWeb2026.xlsx` de ejemplo.
- Prueba manual: script o página temporal que liste períodos en consola/log.

**Checkpoint**: JSON correcto para 3–5 filas conocidas del spec de ejemplo.

### Phase 1 — MVP visual (P1)

- shadcn init + `button`, `select`, `alert`, `card`.
- `TodayGuardBanner` + vista **mensual** FullCalendar + eventos coloreados.
- Destacar período que incluye hoy.
- Estados vacío: sin guardia hoy / planilla ilegible.

**Checkpoint**: US1 verificable manualmente.

### Phase 2 — Vistas y navegación (P2)

- Toolbar: conmutar day / week / month / year en FullCalendar.
- Navegación prev/next; fecha visible coherente.
- Manejo básico de solapes (ambos eventos visibles).

**Checkpoint**: US2 verificable.

### Phase 3 — Filtro por persona (P3)

- `Select` con personas + «Todos».
- Filtrado client-side de eventos; leyenda de colores por persona.

**Checkpoint**: US3 verificable.

### Phase 4 — Actualizar y errores parciales

- Botón «Actualizar» → `router.refresh()`.
- `SkippedRowsAlert` cuando `skippedRowCount > 0`.

**Checkpoint**: FR-010, FR-011.

### Phase 5 — Pulido

- Ajustes responsive básicos; mensajes en español; README/quickstart.
- Revisión de rendimiento con planilla anual completa.

## MVP Scope (recomendado para primer entregable)

Incluir: **Phase 0 + Phase 1** (parseo + mes + hoy + carga inicial).

Diferir a iteración siguiente: filtro persona, vistas extra, botón actualizar (el usuario
puede F5 hasta Phase 4).

## Technical Risks & Mitigations

| Riesgo | Impacto | Mitigación |
|--------|---------|------------|
| Fechas Excel como número serial vs texto `DD-MM-YY` | Filas inválidas | Normalizar en parser: si `cell.t === 'n'` usar conversión SheetJS; si string, regex `DD-MM-YY` |
| Archivo abierto en Excel (`~$*.xlsx`) bloquea lectura | Error en servidor | Documentar cerrar Excel; `.gitignore` ya ignora `~$*` |
| Semana miércoles vs locale FullCalendar (`firstDay`) | Vista semana desalineada | `firstDay: 3` (miércoles) en config FullCalendar |
| Solapes de fechas entre filas | Confusión UX | Mostrar ambos eventos + `warnings` opcional en API |
| Bundle grande si `xlsx` entra al cliente | Perf | Import dinámico solo en `route.ts` / server modules |
| Año 2 dígitos (`25` → 1925 vs 2025) | Fechas erróneas | Regla fija: 00–69 → 2000–2069, 70–99 → 1970–1999 (documentar en parser) |

## Pragmatic Simplifications

- **Una sola hoja** (índice 0); ignorar resto.
- **Sin upload** de Excel por UI; solo ruta en env.
- **Sin internacionalización** i18n framework; strings en español inline.
- **Sin SSR de FullCalendar** — componente `'use client'` único para el calendario.
- **Actualizar** = `router.refresh()` antes que re-fetch API duplicado.

## Development Workflow

1. Rama `001-guardia-dashboard` (activa).
2. Implementar por fases con checkpoint manual (quickstart).
3. Commits pequeños por fase; hooks Spec Kit opcionales.
4. Tras plan: `/speckit-tasks` para tareas atómicas; luego `/speckit-implement`.
5. No generar tests automatizados salvo cambio de constitución.

## Project Structure

### Documentation (this feature)

```text
specs/001-guardia-dashboard/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── schedule-api.md
└── tasks.md          # (/speckit-tasks — not yet)
```

### Source Code (repository root)

```text
guardiaWeb/
├── app/
│   ├── layout.tsx
│   ├── page.tsx
│   └── api/
│       └── schedule/
│           └── route.ts          # GET JSON (opcional si solo RSC)
├── components/
│   ├── schedule-dashboard.tsx    # client orchestrator
│   ├── schedule-calendar.tsx
│   ├── schedule-toolbar.tsx
│   ├── today-guard-banner.tsx
│   └── skipped-rows-alert.tsx
├── lib/
│   ├── excel/
│   │   ├── parse-schedule.ts
│   │   └── parse-date.ts
│   ├── calendar/
│   │   └── map-to-events.ts
│   └── person-colors.ts
├── GuardiaWeb2026.xlsx           # planilla local (ignorada por .gitignore si aplica)
├── .env.local                    # GUARDIA_EXCEL_PATH=./GuardiaWeb2026.xlsx
├── components.json               # shadcn
└── package.json
```

**Structure Decision**: Monolito Next.js en raíz del repo (sin carpetas `frontend/`
ni `backend/`). Toda la lógica de negocio en `lib/`; UI en `components/`; entrada en
`app/`.

## Complexity Tracking

> No violations requiring justification.
