import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { GuardPeriod } from "@/lib/excel/types";
import { formatDateRange } from "@/lib/format-date";

type Props = {
  periods: GuardPeriod[];
};

export function TodayGuardBanner({ periods }: Props) {
  if (periods.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Guardia hoy</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-muted-foreground">
            No hay guardia asignada para hoy.
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="border-primary/40 bg-primary/5">
      <CardHeader>
        <CardTitle>Guardia hoy</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {periods.map((period) => (
          <div key={period.id} className="space-y-1">
            <p className="text-lg font-semibold">{period.person}</p>
            <p className="text-sm text-muted-foreground">
              {formatDateRange(period.start, period.end)}
            </p>
          </div>
        ))}
        {periods.length > 1 && (
          <p className="text-sm text-amber-700">
            Hay más de una guardia activa hoy (solape en planilla).
          </p>
        )}
      </CardContent>
    </Card>
  );
}
