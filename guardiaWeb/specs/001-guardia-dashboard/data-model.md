# Data Model: Dashboard de guardias

**Date**: 2026-05-15 | **Spec**: [spec.md](./spec.md)

## Overview

Datos en memoria tras cada lectura del Excel. Sin persistencia en aplicación. Flujo:
**Planilla (filas)** → **GuardPeriod[]** → **FullCalendar EventInput[]** (derivado en UI).

## Entities

### GuardPeriod

Una semana de guardia (miércoles a martes) asignada a una persona.

| Field | Type | Rules |
|-------|------|-------|
| `id` | `string` | Estable: `row-{sheetRowIndex}` o hash de inicio+persona |
| `person` | `string` | No vacío tras `trim()`; columna `Persona` |
| `start` | `Date` | Primera fecha válida de la fila (columna `M` izquierda) |
| `end` | `Date` | Séptima fecha válida (columna `M` derecha); `end >= start` |
| `sourceRow` | `number` | Índice de fila en hoja (≥2); solo diagnóstico interno |

**Derivación**: Para cada fila de datos, leer columnas B–H (fechas) e I (persona).
`start` = fecha celda B; `end` = fecha celda H (último día de la semana en layout).

### Person (logical)

No tabla separada; conjunto único derivado de `GuardPeriod.person`.

| Rule | Description |
|------|-------------|
| Uniqueness | Clave = `normalize(name)` → trim + colapsar espacios + opcional lower para comparación |
| Display | Texto original de la primera aparición (para UI y dropdown) |
| Color | Asignado por `person-colors.ts` |

### ScheduleData (aggregate root)

Resultado del parseo; lo devuelve el servidor a la UI.

| Field | Type | Rules |
|-------|------|-------|
| `periods` | `GuardPeriod[]` | Solo filas válidas |
| `people` | `string[]` | Orden alfabético locale `es` |
| `skippedRowCount` | `number` | ≥ 0; filas rechazadas |
| `warnings` | `string[]` | Opcional; ej. solapes detectados |
| `loadedAt` | `string` (ISO) | Timestamp de lectura |

### ParseRowError (internal)

No expuesto al usuario final en detalle (FR-011).

| Field | Type |
|-------|------|
| `row` | `number` |
| `reason` | `'missing_person' \| 'invalid_dates' \| 'incomplete_week' \| 'bad_headers'` |

## Excel Layout Mapping

| Excel | Maps to |
|-------|---------|
| Row 1, cols B–I | Headers must equal `M,J,V,S,D,L,M,Persona` (case-sensitive o normalizado en parser) |
| Row 2+, col A | Informative month label; **not** used for date math |
| Row 2+, cols B–H | Seven dates Wed→Tue |
| Row 2+, col I | `GuardPeriod.person` |

## Validation Rules

1. Headers row must match expected 8 columns from column B.
2. Each data row needs 7 parseable dates + non-empty `person`.
3. `end` must be ≥ `start` (typically 6 days later).
4. Invalid rows increment `skippedRowCount` only; do not abort entire parse.
5. Overlapping periods: both kept; optional warning `"overlap detected"`.

## UI State (client, not persisted)

| State | Type | Default |
|-------|------|---------|
| `calendarView` | `'day' \| 'week' \| 'month' \| 'year'` | `'month'` |
| `selectedPerson` | `string \| null` | `null` (= Todos) |
| `visibleRange` | managed by FullCalendar | current month |

## Lookup: guardia en fecha D

```text
periods.filter(p => p.start <= D && p.end >= D)
```

- 0 resultados → “Sin guardia asignada”.
- 1 resultado → banner + highlight.
- 2+ → mostrar todos + aviso solape (edge case).

## FullCalendar Event Mapping (derived)

| GuardPeriod | EventInput |
|-------------|------------|
| `person` | `title` |
| `start` | `start` (inclusive) |
| `end` | `end` exclusive (+1 day si FullCalendar all-day) |
| color | `backgroundColor` / `borderColor` from person hash |
| `id` | `id` |
| active today | `classNames: ['guard-active']` if today ∈ [start, end] |

**Note**: FullCalendar all-day events often use exclusive end; mapper adds 1 day to `end`
when required by API version.

## Relationships

```text
Planilla (1) ──► (N) GuardPeriod
GuardPeriod (N) ──► (1) Person [by name]
ScheduleData (1) ──► (N) GuardPeriod
ScheduleData (1) ──► (N) Person [derived]
```
