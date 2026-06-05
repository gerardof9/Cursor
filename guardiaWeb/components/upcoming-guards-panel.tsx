"use client";

import { ChevronRight } from "lucide-react";
import type { GuardPeriod } from "@/lib/excel/types";
import { formatDateRange } from "@/lib/format-date";
import { personLegendColor } from "@/lib/person-colors";

export function UpcomingGuardsPanel({
  periods,
  onViewAll,
}: {
  periods: GuardPeriod[];
  onViewAll?: () => void;
}) {
  return (
    <section className="app-card-sidebar">
      <h2 className="section-title mb-5">Próximas guardias</h2>
      {periods.length === 0 ? (
        <p className="text-sm text-slate-500">No hay guardias futuras.</p>
      ) : (
        <ul className="space-y-5">
          {periods.map((p) => (
            <li key={p.id} className="flex gap-3">
              <span
                className="mt-2 h-2 w-2 shrink-0 rounded-full opacity-80"
                style={{ backgroundColor: personLegendColor(p.person) }}
              />
              <div>
                <p className="text-sm font-semibold text-slate-900 dark:text-slate-100">
                  {p.person}
                </p>
                <p className="mt-0.5 text-xs text-slate-500">
                  {formatDateRange(p.start, p.end)}
                </p>
              </div>
            </li>
          ))}
        </ul>
      )}
      {periods.length > 0 && onViewAll && (
        <button
          type="button"
          onClick={onViewAll}
          className="mt-5 flex items-center gap-1 text-xs font-semibold text-[var(--brand)] transition-all duration-200 hover:gap-1.5 hover:text-[var(--brand-hover)] dark:text-blue-400"
        >
          Ver todas
          <ChevronRight className="h-3.5 w-3.5" />
        </button>
      )}
    </section>
  );
}
