import { NextResponse } from "next/server";
import { loadSchedule, ScheduleLoadError } from "@/lib/schedule/load-schedule";
import { serializeScheduleData } from "@/lib/schedule/serialize";

export async function GET() {
  try {
    const data = await loadSchedule();
    return NextResponse.json(serializeScheduleData(data));
  } catch (error) {
    if (error instanceof ScheduleLoadError) {
      const status = error.code === "FILE_NOT_FOUND" ? 404 : 500;
      return NextResponse.json(
        { error: error.code, message: error.message },
        { status },
      );
    }
    return NextResponse.json(
      {
        error: "PARSE_FAILED",
        message: "No se pudo leer la planilla de guardias.",
      },
      { status: 500 },
    );
  }
}
