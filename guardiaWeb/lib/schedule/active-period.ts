import type { GuardPeriod } from "@/lib/excel/types";
import { startOfDay } from "@/lib/excel/parse-date";

export function findActivePeriodsForDate(
  periods: GuardPeriod[],
  date: Date = new Date(),
): GuardPeriod[] {
  const day = startOfDay(date);
  return periods.filter((p) => p.start <= day && p.end >= day);
}
