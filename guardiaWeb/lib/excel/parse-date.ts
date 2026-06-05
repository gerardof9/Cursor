import * as XLSX from "xlsx";

const DD_MM_YY = /^(\d{2})-(\d{2})-(\d{2})$/;

function expandYear(twoDigit: number): number {
  if (twoDigit <= 69) return 2000 + twoDigit;
  return 1900 + twoDigit;
}

/** Parse cell value to Date (DD-MM-YY string or Excel serial). */
export function parseExcelDate(value: unknown): Date | null {
  if (value === null || value === undefined || value === "") return null;

  if (value instanceof Date && !Number.isNaN(value.getTime())) {
    return value;
  }

  if (typeof value === "number" && Number.isFinite(value)) {
    const parsed = XLSX.SSF.parse_date_code(value);
    if (!parsed) return null;
    return new Date(parsed.y, parsed.m - 1, parsed.d);
  }

  const text = String(value).trim();
  const match = text.match(DD_MM_YY);
  if (!match) return null;

  const day = Number(match[1]);
  const month = Number(match[2]);
  const year = expandYear(Number(match[3]));
  const date = new Date(year, month - 1, day);
  if (
    date.getFullYear() !== year ||
    date.getMonth() !== month - 1 ||
    date.getDate() !== day
  ) {
    return null;
  }
  return date;
}

export function startOfDay(date: Date): Date {
  const d = new Date(date);
  d.setHours(0, 0, 0, 0);
  return d;
}

export function endOfDay(date: Date): Date {
  const d = new Date(date);
  d.setHours(23, 59, 59, 999);
  return d;
}
