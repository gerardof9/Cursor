# UI Layout Contract

**Version**: 0.1.0 (iteration 002)

## Screen Regions

```text
┌─────────────────────────────────────────────────────────────────┐
│ Help/Filter bar (1 line, collapsible)                           │
├──────────────────────────────┬──────────────────────────────────┤
│                              │                                  │
│  Event List (primary)        │  Event Detail (secondary)        │
│  ~60% terminal width         │  ~40% terminal width             │
│                              │                                  │
│  Columns:                    │  Sections:                       │
│  TIME  OP  SCHEMA  TABLE SRC │  - Header (time, op, object)     │
│                              │  - Format / transaction hints    │
│                              │  - SQL or row values             │
│                              │  - Completeness notes            │
│                              │                                  │
├──────────────────────────────┴──────────────────────────────────┤
│ Status bar: total/filtered counts | indexing % | errors        │
└─────────────────────────────────────────────────────────────────┘
```

## Minimum Terminal Size

- Width: 80 columns
- Height: 24 rows

Below minimum: show centered message "Terminal too small; resize to 80x24".

## Visual States

| State | List area | Detail area | Status bar |
|-------|-----------|-------------|------------|
| Indexing | Shows events as they arrive; cursor active | "Select an event" placeholder | `Indexing: 42% (source/mysql-bin.000123)` |
| Ready | Full index browsable | Detail on selection | `1234 events` |
| Filtered | Subset only | Detail on selection | `12 / 1234 events (filtered)` |
| Empty filter | "No events match filter" | Cleared or last detail grayed | `0 / 1234 events (filtered)` |
| No sources | "Open a binlog: press o" | Empty | `No sources open` |
| Error | Prior data if any | Unchanged | Red error text |

## Selection Behavior

- Single selection; highlighted row in list
- Detail loads asynchronously; show `Loading...` in detail pane until ready
- Changing selection cancels prior detail load (latest wins)

## Modal Overlays

| Modal | Trigger | Dismiss |
|-------|---------|---------|
| Open file path input | `o` | Enter confirm, Esc cancel |
| Filter editor | `f` | Enter apply, Esc cancel |
| Help | `?` | Any key dismiss |

Only one modal active at a time.
