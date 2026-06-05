import {
  Calendar,
  CalendarDays,
  Shield,
  Users,
} from "lucide-react";
import type { ReactNode } from "react";
import { Badge } from "@/components/ui/badge";
import type { GuardPeriod } from "@/lib/excel/types";
import { formatDateRange } from "@/lib/format-date";

type Props = {
  todayPeriods: GuardPeriod[];
  nextPeriod: GuardPeriod | null;
  peopleCount: number;
  totalWeeks: number;
  onGoToday?: () => void;
};

function formatDay(date: Date): string {
  return date.toLocaleDateString("es-AR", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  });
}

function HeroDivider() {
  return <div className="hero-vdivider" aria-hidden />;
}

function HeroStatSegment({
  icon,
  label,
  value,
  subtitle,
  valueClassName = "hero-stat-value",
}: {
  icon: ReactNode;
  label: string;
  value: string;
  subtitle: string;
  valueClassName?: string;
}) {
  return (
    <>
      <HeroDivider />
      <div className="hero-segment">
        <div className="hero-icon-circle" aria-hidden>
          {icon}
        </div>
        <div className="hero-stat-copy">
          <p className="hero-stat-label">{label}</p>
          <p className={valueClassName}>{value}</p>
          <p className="hero-stat-sub">{subtitle}</p>
        </div>
      </div>
    </>
  );
}

export function DashboardHero({
  todayPeriods,
  nextPeriod,
  peopleCount,
  totalWeeks,
  onGoToday,
}: Props) {
  const primary = todayPeriods[0];
  const today = new Date();

  const statusValue = primary ? "En curso" : "Sin guardia";
  const statusSubtitle = primary
    ? `Desde ${formatDay(primary.start)}`
    : "—";

  const nextName = nextPeriod?.person ?? "—";
  const nextSubtitle = nextPeriod
    ? formatDay(nextPeriod.start)
    : "—";

  return (
    <section className="hero-shell">
      <div className="hero-inner">
        <div className="hero-primary">
          <p className="hero-eyebrow">Guardia hoy</p>
          {primary ? (
            <>
              <p className="hero-name">{primary.person}</p>
              <div className="hero-primary-meta">
                <span className="hero-date-range">
                  <CalendarDays className="hero-date-range-icon" />
                  {formatDateRange(primary.start, primary.end)}
                </span>
                <Badge variant="success">En curso</Badge>
              </div>
            </>
          ) : (
            <p className="hero-empty">No hay guardia asignada para hoy.</p>
          )}
        </div>

        <HeroStatSegment
          icon={<Shield />}
          label="Estado actual"
          value={statusValue}
          subtitle={statusSubtitle}
          valueClassName={
            primary ? "hero-stat-value hero-stat-value--success" : "hero-stat-value"
          }
        />

        <HeroStatSegment
          icon={<Calendar />}
          label="Próximo cambio"
          value={nextName}
          subtitle={nextSubtitle}
          valueClassName="hero-stat-value hero-stat-value--accent"
        />

        <HeroStatSegment
          icon={<Users />}
          label="Equipo de guardia"
          value={`${peopleCount} técnicos`}
          subtitle={`${totalWeeks} períodos cargados`}
        />

        <HeroDivider />
        <div className="hero-segment hero-segment-today">
          {onGoToday ? (
            <button
              type="button"
              className="hero-today-trigger"
              onClick={onGoToday}
              aria-label="Ver guardia de hoy"
            >
              <div className="hero-today-icon" aria-hidden>
                <Calendar />
              </div>
              <div className="hero-stat-copy">
                <p className="hero-today-label">Hoy</p>
                <p className="hero-today-date">{formatDay(today)}</p>
              </div>
            </button>
          ) : (
            <div className="hero-today-trigger">
              <div className="hero-today-icon" aria-hidden>
                <Calendar />
              </div>
              <div className="hero-stat-copy">
                <p className="hero-today-label">Hoy</p>
                <p className="hero-today-date">{formatDay(today)}</p>
              </div>
            </div>
          )}
          {onGoToday ? (
            <button
              type="button"
              className="hero-btn-today"
              onClick={onGoToday}
            >
              <CalendarDays />
              Ir a hoy
            </button>
          ) : null}
        </div>
      </div>
    </section>
  );
}
