import type { EventInput } from "@fullcalendar/core";
import type { GuardPeriod } from "@/lib/excel/types";
import { personColor } from "@/lib/person-colors";
import { startOfDay } from "@/lib/excel/parse-date";

function addDays(date: Date, days: number): Date {
  const d = new Date(date);
  d.setDate(d.getDate() + days);
  return d;
}

function eventClass(person: string): string {
  const key = person
    .trim()
    .toLowerCase()
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .replace(/\s+/g, "-");
  return `guard-event guard-event--${key}`;
}

export function mapPeriodsToFullCalendarEvents(
  periods: GuardPeriod[],
  today: Date = new Date(),
): EventInput[] {
  const todayStart = startOfDay(today);

  return periods.map((period) => {
    const isActive = period.start <= todayStart && period.end >= todayStart;
    const exclusiveEnd = addDays(startOfDay(period.end), 1);
    const classes = [eventClass(period.person)];
    if (isActive) classes.push("guard-active");

    return {
      id: period.id,
      title: period.person,
      start: period.start,
      end: exclusiveEnd,
      allDay: true,
      backgroundColor: personColor(period.person),
      borderColor: "transparent",
      classNames: classes,
      extendedProps: {
        person: period.person,
        sourceRow: period.sourceRow,
      },
    };
  });
}
