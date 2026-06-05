import type { GuardPeriod } from "@/lib/excel/types";
import { startOfDay } from "@/lib/excel/parse-date";

export type ScheduleFilter = {
  from: string | null;
  to: string | null;
  /** ISO date (YYYY-MM-DD): guardias activas ese día */
  onDate: string | null;
  person: string | null;
};

export function parseFilterDate(value: string | null): Date | null {
  if (!value) return null;
  const [y, m, d] = value.split("-").map(Number);
  if (!y || !m || !d) return null;
  return startOfDay(new Date(y, m - 1, d));
}

export function applyPeriodFilter(
  periods: GuardPeriod[],
  filter: ScheduleFilter | null,
): GuardPeriod[] {
  if (!filter) return periods;

  let list = periods;
  if (filter.person) {
    list = list.filter((p) => p.person === filter.person);
  }

  const day = parseFilterDate(filter.onDate);
  if (day) {
    return list.filter((p) => p.start <= day && p.end >= day);
  }

  const from = parseFilterDate(filter.from);
  const to = parseFilterDate(filter.to);
  return list.filter((p) => {
    if (from && p.end < from) return false;
    if (to && p.start > to) return false;
    return true;
  });
}

export function periodsActiveOnDate(
  periods: GuardPeriod[],
  date: Date,
): GuardPeriod[] {
  const day = startOfDay(date);
  return periods.filter((p) => p.start <= day && p.end >= day);
}
