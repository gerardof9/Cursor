export type CalendarView = "day" | "week" | "month" | "year";

export function viewToFullCalendar(view: CalendarView): string {
  switch (view) {
    case "day":
      return "dayGridDay";
    case "week":
      return "dayGridWeek";
    case "month":
      return "dayGridMonth";
    case "year":
      return "multiMonthYear";
    default:
      return "dayGridMonth";
  }
}
