# Quickstart: Dashboard de guardias

**Branch**: `001-guardia-dashboard` | **Plan**: [plan.md](./plan.md)

Manual verification only (no automated tests).

## Prerequisites

- Node.js 20 LTS (or 18.18+)
- Planilla `GuardiaWeb2026.xlsx` (layout según [spec.md](./spec.md))
- Excel **cerrado** (evita bloqueo de archivo en Windows)

## Environment

Create `.env.local` at repo root:

```env
GUARDIA_EXCEL_PATH=./GuardiaWeb2026.xlsx
```

Or copy to `./data/GuardiaWeb2026.xlsx` and point accordingly.

## Run (after implementation)

```bash
npm install
npm run dev
```

Open `http://localhost:3000`.

## Manual Test Checklist

### P1 — Guardia hoy

- [ ] Abre la app: vista **mensual** del mes actual
- [ ] Banner o evento destacado muestra persona de guardia si hoy cae en un período
- [ ] Muestra fecha inicio y fin del período
- [ ] Si hoy no tiene guardia: mensaje claro, sin pantalla en blanco

### P2 — Navegación

- [ ] Cambiar a vista día, semana, mes, año
- [ ] Semana comienza en **miércoles** (alineado con planilla)
- [ ] Clic en fecha muestra guardia correcta para ese día

### P3 — Filtro

- [ ] Desplegable lista todos los nombres + «Todos»
- [ ] Al elegir una persona, solo sus períodos visibles
- [ ] Colores consistentes por persona

### Refresh & errors

- [ ] Editar planilla (Excel cerrado), pulsar «Actualizar» o F5 → datos nuevos
- [ ] Fila inválida de prueba → calendario parcial + aviso «N filas no cargadas»

### Sample verification rows

Compare parser output against these rows from the spec example:

| Persona | Inicio (Mié) | Fin (Mar) |
|---------|--------------|-----------|
| Gerardo | 24-12-25 | 30-12-25 |
| Diego | 31-12-25 | 06-01-26 |
| Pablo | 04-02-26 | 10-02-26 |

## Troubleshooting

| Symptom | Check |
|---------|-------|
| Permission denied on xlsx | Close Excel; avoid `~$` lock files |
| Empty calendar | `GUARDIA_EXCEL_PATH`; headers row B1–I1 |
| Wrong year on dates | Parser century rule (see research.md R4) |
| Week view starts Monday | FullCalendar `firstDay: 3` |

## Next Steps

1. `/speckit-tasks` — generate `tasks.md`
2. `/speckit-implement` — execute phases from [plan.md](./plan.md)
