import type { GuardPeriod, ScheduleData, ScheduleDataSerialized } from "@/lib/excel/types";

export function serializeScheduleData(data: ScheduleData): ScheduleDataSerialized {
  return {
    periods: data.periods.map((p) => ({
      id: p.id,
      person: p.person,
      start: p.start.toISOString(),
      end: p.end.toISOString(),
      sourceRow: p.sourceRow,
    })),
    people: data.people,
    skippedRowCount: data.skippedRowCount,
    warnings: data.warnings,
    loadedAt: data.loadedAt,
  };
}

export function deserializePeriods(
  periods: ScheduleDataSerialized["periods"],
): GuardPeriod[] {
  return periods.map((p) => ({
    id: p.id,
    person: p.person,
    start: new Date(p.start),
    end: new Date(p.end),
    sourceRow: p.sourceRow,
  }));
}
