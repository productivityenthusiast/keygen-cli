# keygen-cli — Claude Context

## What is this?
Go CLI for managing a self-hosted Keygen.sh license server. Commands: login, licenses (list/show/status/renew/components), components (check/delete), users (show/status), config, whoami.

## Key Files
- `main.go` — entry, version via `-X main.version=`
- `cmd/*.go` — cobra commands
- `internal/api/*.go` — HTTP client, JSON:API parsing, per-resource methods
- `internal/config/config.go` — .env loading (cwd → ~/.keygen-cli/.env), config file
- `internal/output/formatter.go` — JSON/table/CSV output
- `internal/auth/auth.go` — client creation, token refresh

## JSON:API Pattern
Keygen uses JSON:API. Every response: `{ "data": { "id", "type", "attributes", "relationships" }, "included": [] }`.
Parse: `JSONAPIDocument` → unwrap `Data` → `JSONAPIResource` → `parseLicense()` / `parseMachine()` etc.

## Common Tasks
- **Add command**: new file in `cmd/`, cobra command struct, register in `init()`
- **Add API method**: new func on `*Client` in `internal/api/`, use `doRequest()`, parse JSON:API
- **Build**: `go build -buildvcs=false -ldflags "-s -w -X main.version=0.1.0" -o keygen.exe .`
- **Deploy locally**: `deploy-local.bat` (builds + copies to c:\data\dev + copies .env to ~/.keygen-cli/)
- **Release**: tag `v*` → GitHub Actions → GoReleaser → cross-platform binaries

## Environment
- `KEYGEN_ACCOUNT_ID`, `KEYGEN_BASE_URL`, `KEYGEN_TOKEN` (required)
- `KEYGEN_EMAIL`, `KEYGEN_PASSWORD` (optional, for token refresh)
- Config: `~/.keygen-cli/config.json` and `~/.keygen-cli/.env`
