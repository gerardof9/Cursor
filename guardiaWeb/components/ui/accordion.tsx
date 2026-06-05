"use client";

import { ChevronDown } from "lucide-react";
import { createContext, useContext, useState } from "react";
import { cn } from "@/lib/utils";

const AccordionCtx = createContext<{
  openId: string | null;
  toggle: (id: string) => void;
} | null>(null);

export function Accordion({
  children,
  defaultOpenId = null,
}: {
  children: React.ReactNode;
  defaultOpenId?: string | null;
}) {
  const [openId, setOpenId] = useState<string | null>(defaultOpenId);
  return (
    <AccordionCtx.Provider
      value={{
        openId,
        toggle: (id) => setOpenId((p) => (p === id ? null : id)),
      }}
    >
      {children}
    </AccordionCtx.Provider>
  );
}

export function AccordionItem({
  id,
  title,
  children,
}: {
  id: string;
  title: React.ReactNode;
  children: React.ReactNode;
}) {
  const ctx = useContext(AccordionCtx);
  const open = ctx?.openId === id;
  return (
    <section className="app-card-sidebar overflow-hidden !p-0">
      <button
        type="button"
        className="flex w-full items-center justify-between gap-2 px-6 py-5 text-left transition-all duration-200 hover:bg-slate-50/50 dark:hover:bg-muted/20"
        onClick={() => ctx?.toggle(id)}
        aria-expanded={open}
      >
        {title}
        <ChevronDown
          className={cn(
            "h-4 w-4 shrink-0 text-slate-400 transition-transform duration-200",
            open && "rotate-180",
          )}
        />
      </button>
      {open && (
        <div className="border-t border-slate-100/80 px-6 pb-6 pt-1 dark:border-border">
          {children}
        </div>
      )}
    </section>
  );
}
