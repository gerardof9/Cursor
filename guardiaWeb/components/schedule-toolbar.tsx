"use client";

import { ChevronLeft, ChevronRight, RefreshCw, User } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { CalendarView } from "@/lib/calendar/views";
const VIEWS: { id: CalendarView; label: string }[] = [
  { id: "day", label: "Día" },
  { id: "week", label: "Semana" },
  { id: "month", label: "Mes" },
  { id: "year", label: "Año" },
];

type Props = {
  view: CalendarView;
  onViewChange: (v: CalendarView) => void;
  people: string[];
  selectedPerson: string | null;
  onPersonChange: (p: string | null) => void;
  calendarTitle: string;
  onPrev: () => void;
  onNext: () => void;
  onToday: () => void;
  onRefresh: () => void;
};

export function ScheduleToolbar({
  view,
  onViewChange,
  people,
  selectedPerson,
  onPersonChange,
  calendarTitle,
  onPrev,
  onNext,
  onToday,
  onRefresh,
}: Props) {
  return (
    <section className="toolbar-shell">
      <div className="toolbar-section">
        <span className="toolbar-section-label">Vista</span>
        <div className="tab-group">
          {VIEWS.map((v) => (
            <button
              key={v.id}
              type="button"
              onClick={() => onViewChange(v.id)}
              className={view === v.id ? "tab-active" : "tab-inactive"}
            >
              {v.label}
            </button>
          ))}
        </div>
      </div>

      <div className="toolbar-divider" aria-hidden />

      <div className="toolbar-section">
        <span className="toolbar-section-label">Persona</span>
        <Select
          value={selectedPerson ?? "all"}
          onValueChange={(val) => onPersonChange(val === "all" ? null : val)}
        >
          <SelectTrigger className="toolbar-select h-9 w-[11rem] gap-2 border-slate-200 bg-white text-sm font-semibold shadow-none dark:border-border dark:bg-muted dark:text-slate-100">
            <User className="h-4 w-4 shrink-0 text-slate-500 dark:text-slate-400" />
            <SelectValue placeholder="Todos" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Todos</SelectItem>
            {people.map((p) => (
              <SelectItem key={p} value={p}>
                {p}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="toolbar-divider" aria-hidden />

      <div className="toolbar-section">
        <span className="toolbar-section-label">Fecha</span>
        <div className="toolbar-section-controls">
          <div className="date-nav-box">
          <button
            type="button"
            onClick={onPrev}
            aria-label="Anterior"
            className="date-nav-btn"
          >
            <ChevronLeft className="h-4 w-4" />
          </button>
          <span className="date-nav-title">{calendarTitle || "—"}</span>
          <button
            type="button"
            onClick={onNext}
            aria-label="Siguiente"
            className="date-nav-btn"
          >
            <ChevronRight className="h-4 w-4" />
          </button>
          </div>
          <button type="button" onClick={onToday} className="btn-hoy">
            Hoy
          </button>
        </div>
      </div>

      <div className="toolbar-spacer" aria-hidden />

      <button
        type="button"
        className="btn-primary-soft toolbar-refresh shrink-0"
        onClick={onRefresh}
      >
        <RefreshCw className="h-4 w-4" />
        Restaurar
      </button>
    </section>
  );
}
