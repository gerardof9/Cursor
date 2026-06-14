# CLI Contract: binlog-explorer

**Version**: 0.2.0 (iteration 003)

Extends [002 CLI contract](../../002-explore-binlog-events/contracts/cli.md).

## Invocation

```text
binlog-explorer [flags] [binlog-file ...]
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `binlog-file` | No (0+) | Paths to MySQL binary log files opened at launch |

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--help` | `-h` | — | Show usage and exit |
| `--version` | `-v` | — | Print version and exit |
| `--from` | — | — | Investigation scope start (inclusive). Requires `--to`. |
| `--to` | — | — | Investigation scope end (inclusive). Requires `--from`. |

### Timestamp formats

| Format | Example | Semantics |
|--------|---------|-----------|
| Date and time | `2006-01-02 15:04:05` | Exact boundary |
| Date only | `2006-01-02` | `--from`: start of day 00:00:00; `--to`: end of day 23:59:59 |

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Normal quit |
| `1` | Startup failure (no TTY, invalid flags, only one of `--from`/`--to`, all launch files failed, invalid timestamp parse at CLI) |
| `2` | Partial launch failure (some files failed, at least one succeeded—enters TUI with warning) |

**Note**: CLI scope **out-of-bounds** (FR-011a) is detected **after analysis inside the TUI**, not at flag parse time. The application shows a clear status/error message, does not begin scoped indexing, and remains in the TUI (does not exit `1` unless the user quits with no viable session).

## Startup Behavior

1. If TTY unavailable → exit `1`.
2. Parse flags; if exactly one of `--from`/`--to` → exit `1` with message requiring both or neither.
3. Open each positional file; collect failures.
4. If zero files opened successfully and at least one path given → exit `1`.
5. If zero paths given → enter TUI empty; user opens via `o` (analysis + scope flow).
6. For each opened source: run **analysis pass** (background).
7. If `--from` and `--to` provided: apply as investigation scope, **skip scope dialog**, run analysis, then validate scope against merged analysis min/max (FR-011a).
8. If scope validation fails (out of bounds): show TUI error; do not begin scoped indexing; user may quit or open a different file.
9. If scope valid: begin **scoped indexing**.
10. Else (no CLI scope): show **scope dialog** when analysis completes; begin scoped indexing only after scope confirmed.

## Examples

```bash
# Known window — skip scope dialog
binlog-explorer --from "2025-06-08" --to "2025-06-09" /var/lib/mysql/mysql-bin.000123

# Open file; scope dialog after analysis
binlog-explorer /var/lib/mysql/mysql-bin.000123

# Empty session
binlog-explorer
```

## Environment

No required environment variables.
