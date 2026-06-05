"use client";

import { useCallback, useMemo, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import type { ScheduleDataSerialized } from "@/lib/excel/types";
import { startOfDay } from "@/lib/excel/parse-date";
import { formatIsoDate } from "@/lib/format-date";
import { deserializePeriods } from "@/lib/schedule/serialize";
import { findActivePeriodsForDate } from "@/lib/schedule/active-period";
import {
  applyPeriodFilter,
  parseFilterDate,
} from "@/lib/schedule/filter-periods";
import {
  computeScheduleStats,
  getUpcomingPeriods,
} from "@/lib/schedule/stats";
import type { CalendarView } from "@/lib/calendar/views";
import { DashboardHeader } from "@/components/dashboard-header";
import { DashboardHero } from "@/components/dashboard-hero";
import {
  ScheduleCalendar,
  type NavigateToDateRequest,
  type ScheduleCalendarHandle,
} from "@/components/schedule-calendar";
import { ScheduleToolbar } from "@/components/schedule-toolbar";
import {
  RangeSearchAccordion,
  type RangeFilter,
} from "@/components/range-search-accordion";
import { UpcomingGuardsPanel } from "@/components/upcoming-guards-panel";
import { ScheduleSummaryPanel } from "@/components/schedule-summary-panel";
import { ScheduleFooter } from "@/components/schedule-footer";
import { SkippedRowsAlert } from "@/components/skipped-rows-alert";
import { ScheduleWarningsAlert } from "@/components/schedule-warnings-alert";

type Props = {
  data: ScheduleDataSerialized;
  excelFileName: string;
};

export function ScheduleDashboard({ data, excelFileName }: Props) {
  const router = useRouter();
  const [view, setView] = useState<CalendarView>("month");
  const [selectedPerson, setSelectedPerson] = useState<string | null>(null);
  /** Filtro del calendario (rango / persona / búsqueda por día en el panel) */
  const [calendarFilter, setCalendarFilter] = useState<RangeFilter | null>(
    null,
  );
  const [calendarTitle, setCalendarTitle] = useState("");
  const [navigateToDate, setNavigateToDate] =
    useState<NavigateToDateRequest | null>(null);
  const [searchResetKey, setSearchResetKey] = useState(0);
  const calendarRef = useRef<ScheduleCalendarHandle>(null);

  const requestCalendarDate = useCallback((date: Date) => {
    const day = startOfDay(date);
    setNavigateToDate({ id: Date.now(), date: day });
    calendarRef.current?.gotoDate(day);
  }, []);

  const allPeriods = useMemo(
    () => deserializePeriods(data.periods),
    [data.periods],
  );

  const filteredPeriods = useMemo(() => {
    let list = allPeriods;
    if (selectedPerson) list = list.filter((p) => p.person === selectedPerson);
    return applyPeriodFilter(list, calendarFilter);
  }, [allPeriods, selectedPerson, calendarFilter]);

  const dayLookup = useMemo(() => {
    if (!calendarFilter?.onDate) return null;
    const date = parseFilterDate(calendarFilter.onDate);
    if (!date) return null;
    let list = allPeriods;
    if (selectedPerson) list = list.filter((p) => p.person === selectedPerson);
    const periods = applyPeriodFilter(list, {
      onDate: calendarFilter.onDate,
      from: null,
      to: null,
      person: calendarFilter.person,
    });
    return { date, periods };
  }, [calendarFilter, allPeriods, selectedPerson]);

  const clearScheduleFilters = useCallback(() => {
    setCalendarFilter(null);
    setSelectedPerson(null);
  }, []);

  const restoreDefaultView = useCallback(() => {
    clearScheduleFilters();
    setView("month");
    setSearchResetKey((key) => key + 1);
    requestCalendarDate(new Date());
  }, [clearScheduleFilters, requestCalendarDate]);

  const handleRefresh = useCallback(() => {
    restoreDefaultView();
    router.refresh();
  }, [restoreDefaultView, router]);

  const handleSearchApply = useCallback(
    (filter: RangeFilter) => {
      setCalendarFilter(filter);
      const day = parseFilterDate(filter.onDate);
      if (day) {
        setView("week");
        requestCalendarDate(day);
        return;
      }

      const rangeStart = parseFilterDate(filter.from);
      const rangeEnd = parseFilterDate(filter.to);
      if (rangeStart || rangeEnd) {
        setView("month");
        requestCalendarDate(rangeStart ?? rangeEnd!);
      }
    },
    [requestCalendarDate],
  );

  const handleViewChange = useCallback((next: CalendarView) => {
    setView(next);
    clearScheduleFilters();
  }, [clearScheduleFilters]);

  const todayPeriods = useMemo(
    () => findActivePeriodsForDate(allPeriods),
    [allPeriods],
  );

  const stats = useMemo(
    () => computeScheduleStats(allPeriods, data.people.length),
    [allPeriods, data.people.length],
  );

  const upcoming = useMemo(() => getUpcomingPeriods(allPeriods, 3), [allPeriods]);

  const nextPeriod = upcoming[0] ?? null;

  const handleGoToday = useCallback(() => {
    const today = startOfDay(new Date());
    setSelectedPerson(null);
    setCalendarFilter({
      onDate: formatIsoDate(today),
      from: null,
      to: null,
      person: null,
    });
    setView("week");
    requestCalendarDate(today);
  }, [requestCalendarDate]);

  const onViewAll = useCallback(() => {
    const first = upcoming[0];
    if (first) requestCalendarDate(first.start);
  }, [upcoming, requestCalendarDate]);

  if (allPeriods.length === 0) {
    return (
      <div className="app-shell">
        <DashboardHeader loadedAt={data.loadedAt} />
        <div className="app-card border-dashed p-10 text-center text-muted-foreground">
          No hay períodos de guardia en la planilla cargada.
        </div>
      </div>
    );
  }

  return (
    <div className="app-shell">
      <DashboardHeader loadedAt={data.loadedAt} onRefresh={handleRefresh} />
      <DashboardHero
        todayPeriods={todayPeriods}
        nextPeriod={nextPeriod}
        peopleCount={data.people.length}
        totalWeeks={stats.totalWeeks}
        onGoToday={handleGoToday}
      />
      <SkippedRowsAlert skippedRowCount={data.skippedRowCount} />
      <ScheduleWarningsAlert warnings={data.warnings} />
      <ScheduleToolbar
        view={view}
        onViewChange={handleViewChange}
        people={data.people}
        selectedPerson={selectedPerson}
        onPersonChange={setSelectedPerson}
        calendarTitle={calendarTitle}
        onPrev={() => calendarRef.current?.prev()}
        onNext={() => calendarRef.current?.next()}
        onToday={handleGoToday}
        onRefresh={handleRefresh}
      />
      <div className="grid gap-6 xl:grid-cols-[1fr_300px]">
        <div className="min-w-0">
          <ScheduleCalendar
            ref={calendarRef}
            periods={filteredPeriods}
            view={view}
            people={data.people}
            onTitleChange={setCalendarTitle}
            navigateToDate={navigateToDate}
          />
        </div>
        <aside className="space-y-6">
          <RangeSearchAccordion
            people={data.people}
            onApply={handleSearchApply}
            onClear={restoreDefaultView}
            dayLookup={dayLookup}
            resetKey={searchResetKey}
          />
          <UpcomingGuardsPanel periods={upcoming} onViewAll={onViewAll} />
          <ScheduleSummaryPanel stats={stats} />
        </aside>
      </div>
      <ScheduleFooter />
    </div>
  );
}
