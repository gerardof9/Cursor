# CLI Contract: binlog-explorer

**Version**: 0.1.0 (iteration 002)

## Invocation

```text
binlog-explorer [flags] [binlog-file ...]
```

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `binlog-file` | No (0+) | One or more paths to MySQL binary log files opened at launch |

## Flags (v1)

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--help` | `-h` | — | Show usage and exit |
| `--version` | `-v` | — | Print version and exit |

No config file or subcommands in this iteration.

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Normal quit (`q` or Ctrl+C after clean shutdown) |
| `1` | Startup failure (no TTY, invalid flag, all launch files failed) |
| `2` | Partial launch failure (some files failed, at least one succeeded—still enters TUI with warning) |

## Startup Behavior

1. If TTY unavailable → exit `1` with message.
2. Open each positional file; failures collected per file.
3. If zero files opened successfully and at least one path given → exit `1`.
4. If zero paths given → enter TUI empty; user opens files in-session (`o`).
5. Begin background indexing for all successfully opened sources.

## Environment

No required environment variables for v1.
