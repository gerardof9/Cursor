import * as XLSX from "xlsx";
import type { GuardPeriod, ScheduleData } from "@/lib/excel/types";
import { endOfDay, parseExcelDate, startOfDay } from "@/lib/excel/parse-date";

const DATE_COLS = [1, 2, 3, 4, 5, 6, 7]; // B–H
const PERSON_COL = 8; // I

/** Miércoles — planilla real usa "Mi"; spec de ejemplo usa "M". */
const WED_HEADERS = new Set(["M", "MI", "MIE"]);
/** Martes */
const TUE_HEADERS = new Set(["M", "MA", "MAR"]);
const MIDWEEK_HEADERS = ["J", "V", "S", "D", "L"] as const;
/** Columna de nombre — planilla real usa "Técnico"; spec usa "Persona". */
const PERSON_HEADERS = new Set([
  "PERSONA",
  "TECNICO",
  "GUARDIA",
  "NOMBRE",
]);

function normalizeHeader(value: unknown): string {
  return String(value ?? "")
    .trim()
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .toUpperCase();
}

function cellValue(
  sheet: XLSX.WorkSheet,
  row: number,
  col: number,
): unknown {
  const addr = XLSX.utils.encode_cell({ r: row, c: col });
  return sheet[addr]?.v;
}

function validateHeaders(sheet: XLSX.WorkSheet): boolean {
  const headers = DATE_COLS.concat(PERSON_COL).map((col) =>
    normalizeHeader(cellValue(sheet, 0, col)),
  );
  if (headers.length !== 8) return false;

  if (!WED_HEADERS.has(headers[0]!)) return false;
  for (let i = 0; i < MIDWEEK_HEADERS.length; i++) {
    if (headers[i + 1] !== MIDWEEK_HEADERS[i]) return false;
  }
  if (!TUE_HEADERS.has(headers[6]!)) return false;
  if (!PERSON_HEADERS.has(headers[7]!)) return false;

  return true;
}

function normalizePersonName(raw: string): string {
  return raw.trim().replace(/\s+/g, " ");
}

function detectOverlaps(periods: GuardPeriod[]): string[] {
  const warnings: string[] = [];
  for (let i = 0; i < periods.length; i++) {
    for (let j = i + 1; j < periods.length; j++) {
      const a = periods[i]!;
      const b = periods[j]!;
      if (a.start <= b.end && b.start <= a.end) {
        warnings.push(
          `Solape detectado entre filas ${a.sourceRow} y ${b.sourceRow}.`,
        );
      }
    }
  }
  return warnings;
}

export function parseScheduleFromWorkbook(
  workbook: XLSX.WorkBook,
): ScheduleData {
  const sheetName = workbook.SheetNames[0];
  if (!sheetName) {
    throw new Error("PARSE_FAILED");
  }

  const sheet = workbook.Sheets[sheetName];
  if (!sheet || !validateHeaders(sheet)) {
    throw new Error("PARSE_FAILED");
  }

  const ref = sheet["!ref"];
  if (!ref) {
    return {
      periods: [],
      people: [],
      skippedRowCount: 0,
      warnings: [],
      loadedAt: new Date().toISOString(),
    };
  }

  const range = XLSX.utils.decode_range(ref);
  const periods: GuardPeriod[] = [];
  let skippedRowCount = 0;

  for (let row = 1; row <= range.e.r; row++) {
    const personRaw = cellValue(sheet, row, PERSON_COL);
    const person = normalizePersonName(String(personRaw ?? ""));
    const dates: Date[] = [];

    for (const col of DATE_COLS) {
      const parsed = parseExcelDate(cellValue(sheet, row, col));
      if (parsed) dates.push(startOfDay(parsed));
    }

    if (!person || dates.length !== 7) {
      skippedRowCount++;
      continue;
    }

    const start = dates[0]!;
    const end = endOfDay(dates[6]!);
    if (end < start) {
      skippedRowCount++;
      continue;
    }

    periods.push({
      id: `row-${row + 1}`,
      person,
      start,
      end,
      sourceRow: row + 1,
    });
  }

  const people = [
    ...new Set(periods.map((p) => p.person)),
  ].sort((a, b) => a.localeCompare(b, "es"));

  const warnings = detectOverlaps(periods);

  return {
    periods,
    people,
    skippedRowCount,
    warnings,
    loadedAt: new Date().toISOString(),
  };
}

export function parseScheduleFromBuffer(buffer: Buffer): ScheduleData {
  const workbook = XLSX.read(buffer, { type: "buffer", cellDates: true });
  return parseScheduleFromWorkbook(workbook);
}
