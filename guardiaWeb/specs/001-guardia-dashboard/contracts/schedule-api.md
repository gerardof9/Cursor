# Contract: Schedule API

**Version**: 1.0.0 | **Date**: 2026-05-15

Internal read-only API. Optional if the page loads data via direct server import;
keep contract stable for «Actualizar» via `fetch` if client refresh is added later.

## `GET /api/schedule`

Returns parsed guard periods from the configured Excel file.

### Request

- **Method**: `GET`
- **Auth**: None
- **Query**: none (v1)

### Response `200 OK`

```json
{
  "periods": [
    {
      "id": "row-2",
      "person": "Gerardo",
      "start": "2025-12-24T00:00:00.000Z",
      "end": "2025-12-30T23:59:59.999Z",
      "sourceRow": 2
    }
  ],
  "people": ["Diego", "Enrique", "Gerardo", "Pablo"],
  "skippedRowCount": 0,
  "warnings": [],
  "loadedAt": "2026-05-15T12:00:00.000Z"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `periods` | array | yes | Valid guard periods |
| `periods[].id` | string | yes | Stable id |
| `periods[].person` | string | yes | Display name |
| `periods[].start` | string (ISO 8601) | yes | Period start |
| `periods[].end` | string (ISO 8601) | yes | Period end |
| `periods[].sourceRow` | number | no | Excel row for debugging |
| `people` | string[] | yes | Distinct names for dropdown |
| `skippedRowCount` | number | yes | Rows failed validation |
| `warnings` | string[] | yes | Non-fatal issues (may be empty) |
| `loadedAt` | string (ISO 8601) | yes | Server read timestamp |

### Response `404` / `500` Error

```json
{
  "error": "FILE_NOT_FOUND",
  "message": "No se encontró la planilla de guardias."
}
```

| Code | `error` | When |
|------|---------|------|
| 404 | `FILE_NOT_FOUND` | `GUARDIA_EXCEL_PATH` missing |
| 500 | `PARSE_FAILED` | Headers invalid or workbook unreadable |
| 500 | `CONFIG_MISSING` | Env var not set |

User-facing `message` in Spanish; no stack traces in body.

### Client usage

- Initial load: Server Component may call `loadSchedule()` directly (same shape).
- Refresh button: `fetch('/api/schedule')` **or** `router.refresh()` (preferred in v1).

### Filtering

Person filter is **client-side only** in v1; API always returns full dataset.
