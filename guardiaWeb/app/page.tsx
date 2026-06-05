import { ScheduleDashboard } from "@/components/schedule-dashboard";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { loadSchedule, ScheduleLoadError } from "@/lib/schedule/load-schedule";
import { serializeScheduleData } from "@/lib/schedule/serialize";

function excelFileName(): string {
  const path = process.env.GUARDIA_EXCEL_PATH ?? "./GuardiaWeb2026.xlsx";
  const parts = path.replace(/\\/g, "/").split("/");
  return parts[parts.length - 1] ?? "planilla.xlsx";
}

export default async function HomePage() {
  try {
    const data = await loadSchedule();
    return (
      <ScheduleDashboard
        data={serializeScheduleData(data)}
        excelFileName={excelFileName()}
      />
    );
  } catch (error) {
    const message =
      error instanceof ScheduleLoadError
        ? error.message
        : "No se pudo cargar la planilla de guardias.";

    return (
      <main className="app-shell">
        <Card className="app-card">
          <CardHeader>
            <CardTitle>Error al cargar datos</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm text-muted-foreground">
            <p>{message}</p>
            <p>
              Verifique que exista el archivo configurado en{" "}
              <code className="rounded bg-muted px-1">GUARDIA_EXCEL_PATH</code>{" "}
              y que el formato de la planilla sea el acordado.
            </p>
          </CardContent>
        </Card>
      </main>
    );
  }
}
