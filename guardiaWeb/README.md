# GuardiaWeb

Panel interno de solo lectura para visualizar turnos de guardia desde una planilla Excel.

## Requisitos

- Node.js 20 LTS (o 18.18+) con **npm** en el PATH
- Planilla `GuardiaWeb2026.xlsx` en la raíz del proyecto (o ruta configurada)

## Configuración

```bash
cp .env.local.example .env.local
```

Edite `.env.local` si la planilla está en otra ruta:

```env
GUARDIA_EXCEL_PATH=./GuardiaWeb2026.xlsx
```

Cierre Excel antes de ejecutar la app (evita bloqueo del archivo).

## Desarrollo

```bash
npm install
npm run dev
```

Abra [http://localhost:3000](http://localhost:3000).

## Documentación de la feature

Ver `specs/001-guardia-dashboard/` (spec, plan, tasks, quickstart).
