# Scope Dialog Contract

**Version**: 0.1.0 (iteration 003)

## Purpose

Mandatory investigation scope selection after file analysis (unless CLI `--from`/`--to` provided). Prevents blind full-file indexing on large binlogs.

## Trigger

- Analysis completes for all pending sources and no CLI launch scope set
- User presses `s` to change scope in-session (uses cached analysis)

## Layout

```text
┌─ Investigation Scope ─────────────────────────────────────────┐
│ Selected: mysql-bin.000123 (+ N others)                         │
│ Size: 8.2 GB                                                    │
│ Time span: 2025-05-01 00:12:03 .. 2025-06-10 18:44:11          │
│ Change events: ~2,145,000                                       │
│                                                                 │
│ No investigation scope defined.                                 │
│                                                                 │
│ [1] Entire file                                                 │
│ [2] Last hour (of file)                                         │
│ [3] Last day (of file)                                          │
│ [4] Custom range                                                │
│                                                                 │
│ Custom From: [2025-05-01 00:12:03    ]                          │
│ Custom To:   [2025-06-10 18:44:11    ]                          │
│                                                                 │
│ 1-4 select preset   Enter confirm   Esc cancel                  │
└─────────────────────────────────────────────────────────────────┘
```

## Options

| Key | Option | Scope applied |
|-----|--------|---------------|
| `1` | Entire file | `[MergedMin, MergedMax]`; may trigger large-file warning |
| `2` | Last hour (of file) | `[MergedMax - 1h, MergedMax]` inclusive |
| `3` | Last day (of file) | `[MergedMax - 24h, MergedMax]` inclusive |
| `4` | Custom range | User-edited From/To fields |

## Large-File Warning (secondary modal)

Shown when option **Entire file** selected and size ≥ 1 GiB OR approximate events ≥ 500,000.

```text
┌─ Warning ─────────────────────────────────────────────────────┐
│ Large file detected (8.2 GB).                                   │
│ ~2,145,000 change events                                        │
│ Time span: 2025-05-01 .. 2025-06-10                             │
│                                                                 │
│ Full indexing may take several minutes and use significant      │
│ memory.                                                         │
│                                                                 │
│ [Y] Continue with entire file   [F] Define date range           │
└─────────────────────────────────────────────────────────────────┘
```

| Key | Action |
|-----|--------|
| `Y` / `y` | Confirm Entire file; begin indexing (single-pass preferred) |
| `F` / `f` | Return to scope dialog (focus custom range) |

## Validation

- Custom From/To must parse as timestamps
- `From <= To`
- Range must fall within merged analyzed min/max (error message if not)
- Cannot dismiss without scope except Esc (cancel) or quit

## In-Session Scope Change

Same dialog; title may show `Change Investigation Scope`. Cached analysis reused. On confirm: replace index per [data-model.md](../data-model.md).

## Blocking Behavior

While scope dialog active:
- Browse list empty or frozen (no indexing until confirm)
- Analysis cancel still available if analysis in progress
