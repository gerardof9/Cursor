# Keybindings Contract

**Version**: 0.1.0 (iteration 002)

Global bindings active unless a modal overlay has focus (modal keys take precedence).

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
| `o` | Open additional binlog file (path input overlay) |

## Filtering

| Key | Action |
|-----|--------|
| `f` | Open filter editor (operation, schema/table, time range) |
| `c` | Clear all active filters |

## View

| Key | Action |
|-----|--------|
| `?` | Toggle help overlay |
| `q` | Quit application |
| `Ctrl+C` | Quit application |

## Filter Editor (modal)

| Key | Action |
|-----|--------|
| `Tab` | Cycle fields: operation → schema → table → start time → end time |
| `Enter` | Apply filter and close |
| `Esc` | Cancel without applying |

## Open File (modal)

| Key | Action |
|-----|--------|
| `Enter` | Open path and close |
| `Esc` | Cancel |

## Consistency Rules (future features)

- `j`/`k` vim-style navigation preserved across new list views
- `?`, `q`, `Esc` behavior unchanged when adding panes
- New features bind unused keys before rebinding existing ones
