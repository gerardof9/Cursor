# Research: Dashboard de guardias

**Date**: 2026-05-15 | **Plan**: [plan.md](./plan.md)

## R1 — Ubicación del parseo Excel

**Decision**: Parseo exclusivo en módulos servidor (`lib/excel/*` importados solo desde
`app/page.tsx`, `app/api/schedule/route.ts` o Server Actions futuras).

**Rationale**: Constitución exige camino servidor; `xlsx` ~ hundreds of KB — evitar
client bundle. SheetJS en Node usa `readFile` sobre buffer del disco.

**Alternatives considered**:

| Alternativa | Por qué se descartó |
|-------------|---------------------|
| Parseo en cliente con upload | Fuera de alcance; spec fija archivo en servidor |
| Microservicio de parseo | Sobre-ingeniería para ~52 filas |
| CSV export manual | Rompe flujo del equipo con Excel existente |

## R2 — Estrategia de refresco de datos

**Decision**: Carga en Server Component + `router.refresh()` en botón «Actualizar».

**Rationale**: Alineado con clarificación (solo reload explícito). Un solo camino de
datos; no stale client cache.

**Alternatives considered**:

| Alternativa | Por qué se descartó |
|-------------|---------------------|
| Polling cada N minutos | Rechazado en spec |
| `fs.watch` | Rechazado en spec |
| SWR/React Query con revalidate on focus | Comportamiento impredecible vs spec |

## R3 — FullCalendar vistas y semana miércoles

**Decision**:

| Vista spec | Plugin FullCalendar | `initialView` / view type |
|------------|---------------------|---------------------------|
| Día | `@fullcalendar/daygrid` | `dayGridDay` |
| Semana | `dayGrid` + `duration: { weeks: 1 }` | `dayGridWeek` con `firstDay: 3` |
| Mes | `dayGridMonth` | default MVP |
| Año | `@fullcalendar/multimonth` | `multiMonthYear` |

**Rationale**: Paquetes oficiales; `firstDay: 3` alinea semana con planilla (miércoles).

**Alternatives considered**:

| Alternativa | Por qué se descartó |
|-------------|---------------------|
| Calendario custom con CSS grid | Más código; 4 vistas ya resueltas |
| Solo vista mes en v1 | Aceptado para MVP; resto en Phase 2 |

## R4 — Formato de fecha `DD-MM-YY` y celdas Excel

**Decision**:

1. Si celda es **string** matching `\d{2}-\d{2}-\d{2}` → parsear manualmente.
2. Si celda es **número** (serial Excel) → `xlsx.SSF.parse_date_code` o utilidad
   equivalente de SheetJS.
3. Regla siglo: años 00–69 → 2000–2069; 70–99 → 1970–1999.

**Rationale**: Excel puede guardar fechas como serial aunque el usuario vea `DD-MM-YY`;
cubrir ambos reduce filas omitidas.

**Alternatives considered**:

| Alternativa | Por qué se descartó |
|-------------|---------------------|
| Solo string | Riesgo alto de filas fallidas al re-guardar planilla |
| `date-fns/parse` inmediato | Añadible en Phase 5 si frágil; no obligatorio v1 |

## R5 — Colores por persona

**Decision**: Función determinista `person → hsl(h, 65%, 45%)` con hash simple del
nombre normalizado (trim + lower case).

**Rationale**: Sin tabla de configuración; consistente entre vistas y sesiones.

**Alternatives considered**:

| Alternativa | Por qué se descartó |
|-------------|---------------------|
| Colores en Excel | No existen en layout actual |
| Paleta fija en config JSON | Más mantenimiento para equipo pequeño |

## R6 — Next.js versión y estructura

**Decision**: Next.js 15.x (o última estable al crear proyecto), App Router, sin `src/`
directory, TypeScript strict.

**Rationale**: Estándar actual; Route Handlers y RSC maduros.

**Alternatives considered**:

| Alternativa | Por qué se descartó |
|-------------|---------------------|
| Pages Router | Legacy; peor fit RSC |
| Vite + Express separado | Dos despliegues; viola simplicidad |

## R7 — Variables de entorno

**Decision**:

| Variable | Ejemplo | Uso |
|----------|---------|-----|
| `GUARDIA_EXCEL_PATH` | `./data/GuardiaWeb2026.xlsx` | Ruta absoluta o relativa al cwd |

**Rationale**: Diferente ruta por entorno sin UI de configuración.

**Alternatives considered**:

| Alternativa | Por qué se descartó |
|-------------|---------------------|
| Hardcode en repo | Rompe despliegue |
| Base de datos de config | Fuera de alcance |
