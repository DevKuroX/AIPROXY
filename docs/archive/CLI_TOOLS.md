# docs/CLI_TOOLS.md — One-click CLI Configuration

> **Status:** scaffold.

## What this does
The dashboard can write the config file for popular AI CLIs so they point at
this gateway. Port of 9router's "CLI Tools" page.

## Supported tools

| Tool | Config path (POSIX) | Format |
|---|---|---|
| Claude Code | `~/.claude/settings.json` | JSON |
| Codex CLI | `~/.codex/config.toml` | TOML |
| Cursor | `~/.cursor/cli-config.json` | JSON |
| Cline (VSCode) | `<VSCode user data>/User/globalStorage/saoudrizwan.claude-dev/settings/cline_mcp_settings.json` | JSON |
| Continue (VSCode/JetBrains) | `~/.continue/config.json` | JSON |
| OpenClaw | `~/.openclaw/config.yaml` | YAML |

> Port the exact paths and key names from `_ref/9router/`. Do NOT invent.

## Endpoints

`GET /api/cli-tools` → enumerate installed tools (detected via path existence).

`POST /api/cli-tools/:name/install` body:
```json
{ "apiKey": "9rk_...", "baseUrl": "http://localhost:20128/v1" }
```
Writes the appropriate file. Backs up the existing file to `*.bak.<unix>`.

`POST /api/cli-tools/:name/uninstall` reverts the config to its backup if present.

## Safety
- Never overwrite without backup.
- Validate the target file is parseable in its declared format before writing.
- Refuse to write outside `$HOME` (block path traversal).

## Code layout
- `internal/cli/<tool>.go` — one file per tool with `Install`, `Uninstall`, `Detect`.
- `internal/api/admin/cli_tools.go` — handlers.
