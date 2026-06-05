"use client";

import { CalendarDays, Filter, Search } from "lucide-react";
import { useEffect, useState } from "react";
import { Accordion, AccordionItem } from "@/components/ui/accordion";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { GuardPeriod } from "@/lib/excel/types";
import type { ScheduleFilter } from "@/lib/schedule/filter-periods";
import { formatDateRange } from "@/lib/format-date";

export type RangeFilter = ScheduleFilter;

export function RangeSearchAccordion({
  people,
  onApply,
  onClear,
  dayLookup,
  resetKey,
}: {
  people: string[];
  onApply: (f: ScheduleFilter) => void;
  onClear: () => void;
  dayLookup: { date: Date; periods: GuardPeriod[] } | null;
  resetKey?: number;
}) {
  const [onDate, setOnDate] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [person, setPerson] = useState("all");

  const personValue = person === "all" ? null : person;

  useEffect(() => {
    setOnDate("");
    setFrom("");
    setTo("");
    setPerson("all");
  }, [resetKey]);

  const handleApplyDay = () => {
    if (!onDate) return;
    onApply({
      onDate,
      from: null,
      to: null,
      person: personValue,
    });
  };

  const handleApplyRange = () => {
    onApply({
      onDate: null,
      from: from || null,
      to: to || null,
      person: personValue,
    });
  };

  const handleClear = () => {
    setOnDate("");
    setFrom("");
    setTo("");
    setPerson("all");
    onClear();
  };

  const dayLabel = dayLookup
    ? dayLookup.date.toLocaleDateString("es-AR", {
        day: "2-digit",
        month: "2-digit",
        year: "numeric",
      })
    : null;

  return (
    <Accordion defaultOpenId="search">
      <AccordionItem
        id="search"
        title={
          <span className="section-title flex items-center gap-2.5">
            <Search className="h-4 w-4 text-slate-400" />
            Buscar guardias
          </span>
        }
      >
        <label className="mb-5 grid gap-2 text-sm">
          <span className="text-slate-500 dark:text-slate-400">Persona (opcional)</span>
          <Select value={person} onValueChange={setPerson}>
            <SelectTrigger className="input-field h-10">
              <SelectValue />
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
        </label>

        <div className="mb-5 rounded-lg border border-slate-200/80 bg-slate-50/80 p-4 dark:border-border dark:bg-muted/40">
          <p className="mb-3 flex items-center gap-2 text-sm font-semibold text-[color:var(--brand-fg)]">
            <CalendarDays className="h-4 w-4 shrink-0" />
            Por día
          </p>
          <p className="mb-3 text-xs text-slate-500 dark:text-slate-400">
            Consultá quién está de guardia en una fecha concreta.
          </p>
          <label className="mb-3 grid gap-2 text-sm">
            <span className="text-slate-500 dark:text-slate-400">Fecha</span>
            <input
              type="date"
              value={onDate}
              onChange={(e) => setOnDate(e.target.value)}
              className="input-field"
            />
          </label>
          <button
            type="button"
            className="btn-primary-soft h-9 w-full text-sm"
            disabled={!onDate}
            onClick={handleApplyDay}
          >
            Buscar día
          </button>

          {dayLookup && (
            <div
              className="mt-3 border-t border-slate-200/80 pt-3 dark:border-border"
              role="status"
            >
              <p className="text-xs font-medium text-slate-500">
                Guardia el {dayLabel}
              </p>
              {dayLookup.periods.length === 0 ? (
                <p className="mt-1.5 text-sm text-slate-600 dark:text-slate-300">
                  Sin guardia asignada en esa fecha.
                </p>
              ) : (
                <ul className="mt-2 space-y-2">
                  {dayLookup.periods.map((p) => (
                    <li key={p.id} className="text-sm">
                      <span className="font-semibold text-[color:var(--brand-fg)]">
                        {p.person}
                      </span>
                      <span className="mt-0.5 block text-xs text-slate-500">
                        {formatDateRange(p.start, p.end)}
                      </span>
                    </li>
                  ))}
                </ul>
              )}
              {dayLookup.periods.length > 1 && (
                <p className="mt-2 text-xs text-amber-700">
                  Hay más de una guardia en esa fecha (solape en la planilla).
                </p>
              )}
            </div>
          )}
        </div>

        <div className="border-t border-slate-200/80 pt-4 dark:border-border">
          <p className="mb-3 text-sm font-semibold text-[color:var(--brand-fg)]">
            Por rango
          </p>
          <label className="mb-4 grid gap-2 text-sm">
            <span className="text-slate-500 dark:text-slate-400">Desde</span>
            <input
              type="date"
              value={from}
              onChange={(e) => setFrom(e.target.value)}
              className="input-field"
            />
          </label>
          <label className="mb-4 grid gap-2 text-sm">
            <span className="text-slate-500 dark:text-slate-400">Hasta</span>
            <input
              type="date"
              value={to}
              onChange={(e) => setTo(e.target.value)}
              className="input-field"
            />
          </label>
          <button
            type="button"
            className="btn-primary-soft h-10 w-full"
            onClick={handleApplyRange}
          >
            <Filter className="h-4 w-4" />
            Aplicar rango
          </button>
        </div>

        <Button
          type="button"
          variant="outline"
          className="mt-3 h-10 w-full rounded-lg border-slate-200 dark:border-border dark:bg-card dark:text-slate-100 dark:hover:bg-accent"
          onClick={handleClear}
        >
          Limpiar filtros
        </Button>
      </AccordionItem>
    </Accordion>
  );
}
