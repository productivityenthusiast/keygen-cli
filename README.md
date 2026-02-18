# keygen-cli

A self-describing CLI tool for managing [Keygen](https://keygen.sh) licenses, machines, components, and users. Designed for scripting and LLM integration.

## Installation

### From release
Download the binary for your platform from the [Releases](https://github.com/productivityenthusiast/keygen-cli/releases) page.

### From source
```bash
go build -ldflags "-s -w -X main.version=0.1.0" -o keygen.exe .
```

### Local deploy (Windows)
```cmd
deploy-local.bat
```

## Configuration

Set environment variables or create a `.env` file:

```env
KEYGEN_ACCOUNT_ID=your-account-id
KEYGEN_BASE_URL=https://your-keygen-server.com
KEYGEN_TOKEN=your-admin-token
```

Config lookup order:
1. CLI flags (`--account-id`, `--base-url`, `--token`)
2. Environment variables
3. `.env` in current directory
4. `~/.keygen-cli/.env` (global)
5. `~/.keygen-cli/config.json` (saved via `keygen login`)

## Commands

```
keygen whoami                           # Check auth status
keygen login token --token <token>      # Login with token
keygen login password --email --pass    # Login with credentials
keygen licenses list [--status ...]     # List licenses
keygen licenses show <id>               # Show license details
keygen licenses status <id>             # Validate + status summary
keygen licenses renew <id>              # Renew a license
keygen licenses components <id>         # List components for license
keygen components check <fingerprint>   # Check if device registered
keygen components delete <fp> [--force] # Delete component
keygen users show <id-or-email>         # Show user details
keygen users status <id-or-email>       # User status summary
keygen config show                      # Show config (masked token)
keygen config clear                     # Clear saved config
```

## Output Formats

```
--format json   (default)
--format table
--format csv
```

All JSON output follows: `{ "ok": true/false, "data": ... }` envelope.

## License

MIT
