# keygen-cli — Product Requirements Document

## 1. Overview
keygen-cli is a Go-based command-line tool for managing Keygen.sh license server resources. It is designed for:
- **Scripting**: Automating license management tasks in CI/CD, deployment scripts, and cron jobs
- **LLM Integration**: Machine-readable JSON output for AI agent consumption
- **Admin Operations**: Quick license/user/component management from the terminal

## 2. Target API
Self-hosted Keygen EE instance (JSON:API format). All endpoints under `/v1/accounts/{account_id}/`.

## 3. Commands

### Authentication
| Command | Description |
|---------|-------------|
| `keygen login token --token <t>` | Validate and save an admin API token |
| `keygen login password --email --password` | Generate token via Basic Auth |
| `keygen whoami` | Show current auth context |

### Licenses
| Command | Description |
|---------|-------------|
| `keygen licenses list` | List licenses with filters (--user, --product, --policy, --status, --limit, --page) |
| `keygen licenses show <id>` | Show license details |
| `keygen licenses status <id>` | Validate license + machine/component counts + days remaining |
| `keygen licenses renew <id>` | Renew license, show old/new expiry |
| `keygen licenses components <id>` | List all components across all machines for a license |

### Components
| Command | Description |
|---------|-------------|
| `keygen components check <fingerprint>` | Check if a device fingerprint is registered |
| `keygen components delete <fingerprint>` | Delete component by fingerprint (--force) |

### Users
| Command | Description |
|---------|-------------|
| `keygen users show <id-or-email>` | Show user details + license count |
| `keygen users status <id-or-email>` | Aggregate: active/expiring/expired/suspended licenses, machines, components |

### Config
| Command | Description |
|---------|-------------|
| `keygen config show` | Show current config (masked token) |
| `keygen config clear` | Remove saved config file |

## 4. Output Formats
- **JSON** (default): `{ "ok": true, "data": { ... } }` envelope
- **Table**: ASCII table via tablewriter
- **CSV**: Standard CSV output

## 5. Configuration Precedence
1. CLI flags (--account-id, --base-url, --token)
2. Environment variables (KEYGEN_ACCOUNT_ID, KEYGEN_BASE_URL, KEYGEN_TOKEN)
3. `.env` in current working directory
4. `~/.keygen-cli/.env` (global)
5. `~/.keygen-cli/config.json` (saved config)

## 6. Architecture
```
main.go              → entry point, version injection via ldflags
cmd/                 → cobra commands (one file per command group)
internal/api/        → HTTP client, JSON:API parsing, domain types
internal/config/     → .env loading, config file management
internal/output/     → JSON/table/CSV formatters
internal/auth/       → token resolution, auto-refresh
```

## 7. Build & Distribution
- Go 1.22+, CGO_ENABLED=0 for static binaries
- GoReleaser for cross-platform releases (linux/darwin/windows × amd64/arm64)
- GitHub Actions: CI on push/PR, Release on `v*` tags
- Local deploy: `deploy-local.bat`

## 8. Error Handling
- All errors return `{ "ok": false, "error": "message" }` JSON
- API errors parsed from Keygen's error envelope (title, detail, code)
- Non-zero exit codes for failures

## 9. Security
- TLS verification configurable (skip for self-signed certs)
- Tokens stored in `~/.keygen-cli/config.json` with 0600 permissions
- .env files excluded from git

## 10. Dependencies
- github.com/spf13/cobra — CLI framework
- github.com/joho/godotenv — .env file loading
- github.com/olekukonko/tablewriter — ASCII tables
- github.com/fatih/color — Terminal colors

## 11. Acceptance Criteria
- [ ] All commands return valid JSON by default
- [ ] `--format table` and `--format csv` work for list/status commands
- [ ] Global .env fallback works when no local .env present
- [ ] Cross-platform binaries build successfully
- [ ] `deploy-local.bat` builds and installs to PATH
