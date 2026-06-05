import "server-only";

import { readFile } from "fs/promises";
import { resolve } from "path";
import type { ScheduleData } from "@/lib/excel/types";
import { parseScheduleFromBuffer } from "@/lib/excel/parse-schedule";

export class ScheduleLoadError extends Error {
  constructor(
    public readonly code: "FILE_NOT_FOUND" | "CONFIG_MISSING" | "PARSE_FAILED",
    message: string,
  ) {
    super(message);
    this.name = "ScheduleLoadError";
  }
}

export async function loadSchedule(): Promise<ScheduleData> {
  const configured = process.env.GUARDIA_EXCEL_PATH;
  if (!configured?.trim()) {
    throw new ScheduleLoadError(
      "CONFIG_MISSING",
      "Falta la variable GUARDIA_EXCEL_PATH.",
    );
  }

  const filePath = resolve(process.cwd(), configured);
  let buffer: Buffer;
  try {
    buffer = await readFile(filePath);
  } catch {
    throw new ScheduleLoadError(
      "FILE_NOT_FOUND",
      "No se encontró la planilla de guardias.",
    );
  }

  try {
    return parseScheduleFromBuffer(buffer);
  } catch {
    throw new ScheduleLoadError(
      "PARSE_FAILED",
      "No se pudo leer la planilla. Verifique el formato.",
    );
  }
}
