"use client";

import {
  forwardRef,
  useCallback,
  useEffect,
  useImperativeHandle,
  useMemo,
  useRef,
} from "react";
import FullCalendar from "@fullcalendar/react";
import type { CalendarApi } from "@fullcalendar/core";
import type { DatesSetArg } from "@fullcalendar/core";
import esLocale from "@fullcalendar/core/locales/es";
import dayGridPlugin from "@fullcalendar/daygrid";
import multiMonthPlugin from "@fullcalendar/multimonth";
import interactionPlugin from "@fullcalendar/interaction";
import type { CalendarView } from "@/lib/calendar/views";
import { viewToFullCalendar } from "@/lib/calendar/views";
import type { GuardPeriod } from "@/lib/excel/types";
import { mapPeriodsToFullCalendarEvents } from "@/lib/calendar/map-to-events";
import { holidayDayCellClassNames } from "@/lib/calendar/holidays";
import { PersonLegend } from "@/components/person-legend";
import { personEventGradient } from "@/lib/person-colors";

export type ScheduleCalendarHandle = {
  prev: () => void;
  next: () => void;
  today: () => void;
  gotoDate: (date: Date) => void;
};

export type NavigateToDateRequest = {
  id: number;
  date: Date;
};

export const ScheduleCalendar = forwardRef<
  ScheduleCalendarHandle,
  {
    periods: GuardPeriod[];
    view: CalendarView;
    people?: string[];
    onTitleChange?: (title: string) => void;
    navigateToDate?: NavigateToDateRequest | null;
  }
>(function ScheduleCalendar(
  { periods, view, people, onTitleChange, navigateToDate },
  ref,
) {
  const calendarRef = useRef<FullCalendar>(null);
  const apiRef = useRef<CalendarApi | null>(null);
  const events = useMemo(() => mapPeriodsToFullCalendarEvents(periods), [periods]);

  const getApi = useCallback(
    () => apiRef.current ?? calendarRef.current?.getApi() ?? null,
    [],
  );

  useImperativeHandle(
    ref,
    () => ({
      prev: () => getApi()?.prev(),
      next: () => getApi()?.next(),
      today: () => getApi()?.today(),
      gotoDate: (d: Date) => getApi()?.gotoDate(d),
    }),
    [getApi],
  );

  useEffect(() => {
    getApi()?.changeView(viewToFullCalendar(view));
  }, [view, getApi]);

  useEffect(() => {
    if (!navigateToDate) return;
    const api = getApi();
    if (!api) return;
    api.gotoDate(navigateToDate.date);
  }, [navigateToDate, getApi]);

  return (
    <section className="app-card overflow-hidden">
      {people && people.length > 0 && (
        <div className="border-b border-slate-100/60 px-6 py-5 dark:border-border/60">
          <PersonLegend people={people} />
        </div>
      )}
      <div className="guardia-fc fc-theme-standard px-4 pb-6 pt-4 sm:px-6 sm:pb-8">
        <FullCalendar
          ref={calendarRef}
          plugins={[dayGridPlugin, multiMonthPlugin, interactionPlugin]}
          initialView={viewToFullCalendar(view)}
          locales={[esLocale]}
          locale="es"
          firstDay={3}
          height="auto"
          headerToolbar={false}
          datesSet={(info: DatesSetArg) => {
            apiRef.current = info.view.calendar;
            onTitleChange?.(info.view.title);
          }}
          events={events}
          eventDisplay="block"
          fixedWeekCount={false}
          showNonCurrentDates
          dayCellClassNames={holidayDayCellClassNames}
          eventDidMount={(info) => {
            const person = info.event.extendedProps.person as string | undefined;
            if (person && info.el) {
              info.el.style.background = personEventGradient(person);
              info.el.style.borderColor = "transparent";
              info.el.style.color = "white";
            }
          }}
        />
      </div>
    </section>
  );
});
