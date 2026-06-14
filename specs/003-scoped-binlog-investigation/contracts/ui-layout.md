# UI Layout Contract

**Version**: 0.2.0 (iteration 003)

Extends [002 UI layout](../../002-explore-binlog-events/contracts/ui-layout.md).

## Screen Regions

```text
┌─────────────────────────────────────────────────────────────────┐
│ Top bar: scope summary | secondary filter summary | help hints  │
├──────────────────────────────┬──────────────────────────────────┤
│  Event List (~60%)           │  Event Detail (~40%)             │
├──────────────────────────────┴──────────────────────────────────┤
│ Status: phase | progress | shown/scoped counts | errors         │
└─────────────────────────────────────────────────────────────────┘
```

## Minimum Terminal Size

Unchanged: 80×24.

## Visual States

| State | List area | Detail area | Status bar |
|-------|-----------|-------------|------------|
| Analyzing | "Analyzing file…" or empty | Placeholder | `Analyzing: 42% \| mysql-bin.000123` |
| Awaiting scope | Empty / frozen | Placeholder | `Analysis complete — select scope (s)` |
| Scoped indexing | Events arrive within scope | Placeholder / loading | `Indexing scope: 2025-06-08..09 \| 67%` |
| Ready | Full scoped index | Detail on selection | `340 events` |
| Secondary filtered | Subset | Detail on selection | `12 / 340 events (filtered)` |
| Empty scoped index | Empty-index message | Placeholder | `0 events in scope` |
| Scope change | Cleared during re-index | Cleared | `Scope changed — re-indexing…` |
| Error | Prior data if any | Unchanged | Error text |

## Top Bar Examples

```text
Scope: 2025-06-08 00:00:00 .. 2025-06-09 23:59:59  |  o open  s scope  f filter  ? help  q quit

Scope: Last day (of file)  |  Filters: table=orders  |  ...
```

## Modal Overlays

| Modal | Trigger | Dismiss |
|-------|---------|---------|
| Open file | `o` | Enter / Esc |
| Investigation scope | After analysis, or `s` | Enter confirm / Esc |
| Large-file warning | Entire file on large binlog | Y / F |
| Secondary filter | `f` | Enter / Esc |
| Help | `?` | Any key |

Only one modal at a time. Scope dialog blocks indexing start.

## Selection Behavior

Unchanged from 002: async detail load, latest selection wins.

## Progress Display

| Phase | Progress metric |
|-------|-----------------|
| Analysis | Bytes read / file size (%) |
| Scoped index | Bytes read and/or current event timestamp |

Copy guidance: use **"Analyzing file…"** and **"Indexing scope…"** (not "pass 1 of 2" unless both phases visibly sequential for non-Entire scopes).
