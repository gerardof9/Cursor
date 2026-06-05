"use client";

import { Moon, RefreshCw, Shield, Sun } from "lucide-react";
import { useTheme } from "@/components/theme-provider";
import { formatLoadedAt } from "@/lib/format-date";

export function DashboardHeader({
  loadedAt,
  onRefresh,
}: {
  loadedAt: string;
  onRefresh?: () => void;
}) {
  const { theme, toggleTheme } = useTheme();

  return (
    <header className="flex flex-col gap-6 pb-2 sm:flex-row sm:items-center sm:justify-between">
      <div className="flex items-center gap-4">
        <div
          className="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl text-white shadow-lg transition-transform duration-200 hover:scale-[1.02]"
          style={{
            backgroundColor: "var(--brand)",
            boxShadow: "0 4px 12px rgb(30 64 175 / 0.25)",
          }}
        >
          <Shield className="h-6 w-6" aria-hidden />
        </div>
        <div>
          <h1 className="text-[1.875rem] font-bold tracking-[-0.03em] text-slate-900 md:text-[2rem] dark:text-slate-50">
            GuardiaWeb
          </h1>
          <p className="mt-1.5 text-sm leading-relaxed text-slate-600 dark:text-slate-400">
            Consulta de guardias del equipo (solo lectura)
          </p>
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-2.5 sm:justify-end">
        <span className="mr-1 w-full text-xs text-slate-500 sm:w-auto sm:text-right">
          Última actualización: {formatLoadedAt(loadedAt)}
        </span>
        <button
          type="button"
          aria-label="Actualizar"
          onClick={onRefresh}
          className="btn-ghost-icon"
        >
          <RefreshCw className="h-4 w-4" />
        </button>
        <button
          type="button"
          aria-label="Cambiar tema"
          onClick={toggleTheme}
          className="btn-ghost-icon"
        >
          {theme === "dark" ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
        </button>
      </div>
    </header>
  );
}
