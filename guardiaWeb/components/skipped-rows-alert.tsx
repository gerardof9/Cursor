"use client";

import { useState } from "react";
import { AlertTriangle, ChevronDown } from "lucide-react";
import { cn } from "@/lib/utils";

export function SkippedRowsAlert({ skippedRowCount }: { skippedRowCount: number }) {
  const [open, setOpen] = useState(false);
  if (skippedRowCount <= 0) return null;

  return (
    <div
      className="flex items-center gap-4 rounded-2xl px-6 py-4 ring-1 ring-amber-200/30 transition-all duration-200"
      style={{
        backgroundColor: "hsl(var(--warning-bg))",
        color: "hsl(var(--warning-fg))",
      }}
      role="status"
    >
      <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-amber-100/80 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300">
        <AlertTriangle className="h-4 w-4" />
      </div>
      <div className="min-w-0 flex-1 text-sm leading-snug">
        <span className="font-semibold">Filas omitidas</span>
        <span className="mx-1.5 text-amber-800/40">·</span>
        {skippedRowCount} fila{skippedRowCount === 1 ? "" : "s"} no se cargaron por datos
        incompletos o inválidos.
      </div>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="flex shrink-0 items-center gap-1 rounded-lg px-2 py-1 text-xs font-medium transition-colors hover:bg-amber-100/50"
      >
        Ver detalles
        <ChevronDown className={cn("h-3.5 w-3.5 transition-transform", open && "rotate-180")} />
      </button>
    </div>
  );
}
