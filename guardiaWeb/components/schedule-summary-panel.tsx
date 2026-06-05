import type { ScheduleStats } from "@/lib/schedule/stats";

const ROWS: { key: keyof ScheduleStats; label: string }[] = [
  { key: "totalWeeks", label: "Total semanas" },
  { key: "totalPeople", label: "Personas" },
  { key: "completed", label: "Guardias completadas" },
  { key: "inProgress", label: "En curso" },
  { key: "upcoming", label: "Próximas" },
];

export function ScheduleSummaryPanel({ stats }: { stats: ScheduleStats }) {
  return (
    <section className="app-card-sidebar">
      <h2 className="section-title mb-5">Resumen anual</h2>
      <dl className="space-y-4">
        {ROWS.map(({ key, label }) => (
          <div key={key} className="flex items-center justify-between gap-4">
            <dt className="text-sm text-slate-600 dark:text-slate-400">{label}</dt>
            <dd className="text-sm font-bold tabular-nums text-slate-900 dark:text-slate-100">
              {stats[key]}
            </dd>
          </div>
        ))}
      </dl>
    </section>
  );
}
