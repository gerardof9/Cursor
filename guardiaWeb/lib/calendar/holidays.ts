/** Feriados fijos (mes-día, cualquier año). */
const HOLIDAY_MONTH_DAY = new Set([
  "01-01",
  "01-06",
  "04-19",
  "05-01",
  "07-18",
  "08-25",
  "11-02",
  "12-25",
]);

/** FullCalendar usa fechas UTC en celdas; getDate() local desplaza el día. */
function monthDayKey(date: Date): string {
  const month = String(date.getUTCMonth() + 1).padStart(2, "0");
  const day = String(date.getUTCDate()).padStart(2, "0");
  return `${month}-${day}`;
}

export function isHoliday(date: Date): boolean {
  return HOLIDAY_MONTH_DAY.has(monthDayKey(date));
}

export function holidayDayCellClassNames(arg: { date: Date }): string[] {
  return isHoliday(arg.date) ? ["fc-day-holiday"] : [];
}
