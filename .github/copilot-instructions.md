# keygen-cli — Copilot Instructions

## Project Overview
Go CLI tool for Keygen license server management. Uses cobra for commands, JSON:API parsing for the Keygen API.

## Architecture
- `main.go` — entry point, version via ldflags
- `cmd/` — cobra commands, one file per command group
- `internal/api/` — HTTP client + JSON:API types + domain methods
- `internal/config/` — .env + config file loading
- `internal/output/` — JSON/table/CSV formatters
- `internal/auth/` — token resolution + auto-refresh

## Code Conventions
- All public API methods return `(result, error)`
- JSON:API responses parsed via `JSONAPIDocument` → `JSONAPIResource` → domain types
- Output always through `output.Success()`, `output.Error()`, or `output.SuccessList()`
- Config loaded via `loadConfig()` in each command's Run function
- Client obtained via `auth.ResolveClient(cfg)` which handles token refresh

## API Reference
- Base: `{KEYGEN_BASE_URL}/v1/accounts/{KEYGEN_ACCOUNT_ID}`
- Auth: `Authorization: Bearer {token}` or Basic Auth for token creation
- Format: JSON:API (`application/vnd.api+json`)
- Key endpoints: `/licenses`, `/machines`, `/components`, `/users`, `/tokens`, `/me`

## Adding New Commands
1. Create `cmd/newcommand.go`
2. Define cobra command with `Use`, `Short`, `Args`, `Run`
3. In `Run`: `loadConfig()` → `auth.ResolveClient()` → API call → `output.Success()`
4. Register in `init()`: `parentCmd.AddCommand(newCmd)` + `rootCmd.AddCommand(parentCmd)`

## Adding New API Methods
1. Add method to appropriate file in `internal/api/`
2. Use `c.doRequest()` for authenticated calls
3. Parse response: `json.Unmarshal` → `JSONAPIDocument` → `JSONAPIResource` → domain type
4. Return parsed domain type

## Testing
- `go test ./...` for unit tests
- Live testing: ensure `.env` has valid credentials
- Build: `go build -buildvcs=false -ldflags "-s -w -X main.version=0.1.0" -o keygen.exe .`
