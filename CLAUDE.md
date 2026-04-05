# cursor-usage

CLI tool for querying Cursor IDE usage data via the unofficial dashboard API.

## Build and Test

```bash
go build -o cursor-usage .
go test ./...
```

Cross-compile:
```bash
make release
```

## Project Structure

- `cmd/` - Cobra command definitions (thin orchestration layer)
- `internal/api/` - HTTP client, API types, endpoint methods
- `internal/config/` - Config file I/O with XDG path resolution
- `internal/output/` - Table and JSON renderers
- `internal/dateparse/` - Human-friendly date string to millisecond timestamp conversion
- `testdata/` - JSON fixtures for API response tests

## Architecture

- `api.Client` handles cookie injection and CSRF headers; all endpoint methods live on it
- Config is stored as JSON at the platform-appropriate XDG config path
- Output formatters take typed structs and an `io.Writer`; the `cmd/` layer picks table vs JSON based on `--json` flag
- Tests use `net/http/httptest` servers with fixture files; no mocking frameworks

## API Reference

The API is documented in `~/claude-working/cursor-unofficial-api.md`. Key points:
- Auth: `WorkosCursorSessionToken` cookie (httpOnly, extracted from browser DevTools)
- CSRF: POST endpoints require `Origin: https://cursor.com` header
- Three endpoints: `GET /api/usage-summary`, `GET /api/usage`, `POST /api/dashboard/get-filtered-usage-events`
