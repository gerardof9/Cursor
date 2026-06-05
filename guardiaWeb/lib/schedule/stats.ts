import type { GuardPeriod } from "@/lib/excel/types";
import { startOfDay } from "@/lib/excel/parse-date";

export type ScheduleStats = {
  totalWeeks: number;
  totalPeople: number;
  completed: number;
  inProgress: number;
  upcoming: number;
};

export function computeScheduleStats(
  periods: GuardPeriod[],
  totalPeople: number,
): ScheduleStats {
  const today = startOfDay(new Date());

  let completed = 0;
  let inProgress = 0;
  let upcoming = 0;

  for (const p of periods) {
    if (p.end < today) completed++;
    else if (p.start <= today && p.end >= today) inProgress++;
    else if (p.start > today) upcoming++;
  }

  return {
    totalWeeks: periods.length,
    totalPeople,
    completed,
    inProgress,
    upcoming,
  };
}

export function getUpcomingPeriods(
  periods: GuardPeriod[],
  limit = 3,
): GuardPeriod[] {
  const today = startOfDay(new Date());
  return periods
    .filter((p) => p.start > today)
    .sort((a, b) => a.start.getTime() - b.start.getTime())
    .slice(0, limit);
}
