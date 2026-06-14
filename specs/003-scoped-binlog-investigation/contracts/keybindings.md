# Keybindings Contract

**Version**: 0.2.0 (iteration 003)

Extends [002 keybindings](../../002-explore-binlog-events/contracts/keybindings.md).

## Navigation

| Key | Action |
|-----|--------|
| `↑` / `k` | Previous event in filtered list |
| `↓` / `j` | Next event in filtered list |
| `PgUp` | Page up in list |
| `PgDn` | Page down in list |
| `Home` / `g` | First event in filtered list |
| `End` / `G` | Last event in filtered list |

## Sources

| Key | Action |
|-----|--------|
| `o` | Open additional binlog file (triggers analysis → scope flow) |

## Investigation Scope

| Key | Action |
|-----|--------|
| `s` | Open investigation scope dialog (change time window) |

## Filtering (secondary)

| Key | Action |
|-----|--------|
| `f` | Open filter editor (operation, schema, table only—no time range) |
| `c` | Clear secondary filters (does not change investigation scope) |

## View

| Key | Action |
|-----|--------|
| `?` | Toggle help overlay |
| `q` | Quit application |
| `Ctrl+C` | Quit application |

## Scope Dialog (modal)

| Key | Action |
|-----|--------|
| `1`–`4` | Select preset option |
| `Tab` | Cycle custom From/To fields (when option 4 active) |
| `Enter` | Confirm scope and start scoped indexing |
| `Esc` | Cancel (abort scope change; in-session revert to prior scope if changing) |

## Large-File Warning (modal)

| Key | Action |
|-----|--------|
| `Y` / `y` | Continue with entire file |
| `F` / `f` | Return to scope dialog |

## Filter Editor (modal)

| Key | Action |
|-----|--------|
| `Tab` | Cycle: operation → schema → table |
| `Enter` | Apply filter and close |
| `Esc` | Cancel without applying |

## Open File (modal)

| Key | Action |
|-----|--------|
| `Enter` | Open path (analysis begins) |
| `Esc` | Cancel |

## Analysis In Progress

| Key | Action |
|-----|--------|
| `Esc` | Cancel analysis and abort open for pending source |

## Consistency Rules

- `s` is scope (time window); `f` is secondary filter (table/op)—never combined
- `j`/`k` preserved across list views
- Modals take precedence over global keys
