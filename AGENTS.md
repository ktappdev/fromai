# fromai — Agent Instructions

## Project Description

**fromai** is a coding task manager where users create tasks with starter code and edit them in a Monaco editor. SvelteKit frontend (Svelte 5 runes mode) talks to a PocketBase v0.39.0 Go backend. All data persists to SQLite via PocketBase. Frontend is client-side only (no SSR).

## Commands

```bash
# Frontend (pnpm with engine-strict=true)
cd frontend
pnpm install              # Install deps
pnpm run dev              # Dev server (opens browser via vite.server.open)
pnpm run build            # Production build
pnpm run check            # svelte-check typecheck
pnpm run check:watch      # Watch mode typecheck
pnpm run preview          # Preview production build

# Backend
cd backend && go run main.go   # Start PocketBase server (port 8090)

# Both must run concurrently for full stack development
```

## Coding Conventions

### General
- Keep files and components small (< 400 lines)
- TypeScript strict mode enabled
- Svelte 5 runes mode (`$state`, `$derived`, etc.)
- No typed interfaces yet — `$state<any>` used throughout
- Plain HTML/CSS with Svelte — no UI kits, no form libraries

### Frontend Patterns
- **Client-side only**: No SSR, no `+page.server.ts`. Guard `localStorage` with `typeof window !== 'undefined'`
- **Data fetching**: All in `.svelte` files via SvelteKit `load()` or direct effects
- **PocketBase client**: Use `PocketBaseClient` class from `frontend/src/lib/pocketbase.ts` — not raw fetch
- **Auth token**: Stored in localStorage as raw string (no "Bearer " prefix)
- **Monaco editor**: Loaded from CDN, not bundled

### Backend Patterns
- All custom routes in `backend/main.go` (~180 lines)
- Routes use `apis.RequireAuth()` wrapper with `.Bind()` for auth
- Read authenticated user via `e.Auth`
- CORS set manually per handler
- Single PocketBase binary with embedded assets

## PocketBase Gotchas (Critical)

See `POCKETBASE_GOTCHAS.md` for complete reference. It covers both API migration gotchas (sections 1-17) and production deployment gotchas (sections 18-29). Key points:

1. **Auth header**: `Authorization: <token>` — raw token, NO "Bearer " prefix (v0.39.0 breaking change)
2. **Path params**: `e.Request.PathValue("id")` — NOT `e.PathParam()`
3. **Record access**: Methods on `app` directly (`app.FindRecordById()`, `app.Save()`) — NOT `app.Dao()`
4. **New records**: `core.NewRecord(collection)` — NOT `daos.NewRecord()`
5. **Field access**: `record.GetString("field")` — NOT `record.GetStringDataValue()`
6. **CORS**: Must set manually on every custom handler (not automatic)
7. **Auth middleware**: `.Bind(apis.RequireAuth())` on routes, then use `e.Auth` in handler
8. **JSON body**: `json.NewDecoder(e.Request.Body).Decode(&data)` — no `apis.Bind()`
9. **FindRecordsByFilter params**: `map[string]any{}` — NOT `dbx.Params{}`
10. **Route registration**: `se.Router.GET("/path", handler).Bind(apis.RequireAuth())` inside `app.OnServe().BindFunc()`
11. **Production encryption**: Set `PB_ENCRYPTION_KEY` env var before first run to encrypt SMTP passwords, OAuth2 secrets at rest (cannot be retrofitted)

## Collection Setup

See `SETUP_POCKETBASE.md` for step-by-step PocketBase collection creation.

## CLI (Agent Interface)

AI agents can manage tasks via the Go CLI at `cli/`. It's both a binary and an importable library.

**For AI agents**: See `skills/fromai/SKILL.md` for the complete skill guide on when and how to use `fai`, including async workflow, grading rubrics, and best practices.

### Setup (one-time)

```bash
# Get your API key from the fromai settings page (one-time human setup)
fai init --key "your-api-key"
```

This stores the key in `~/.config/fromai/config.toml`. Subsequent commands read it automatically.

### Binary usage

```bash
# Build
cd cli && go build -o fai ./cmd/fai

# Verify auth
fai whoami

# Commands
fai task create --title "Sort Array" --starter-code "// TODO" --language typescript
fai task list
fai task get <id>
fai task update <id> --code "..."
fai task submit <id>                   # human action
fai task grade <id> --grade "A" --feedback "nice work"
fai task delete <id>                    # archive by default
fai task delete <id> --hard           # permanent delete
fai task poll <id>                    # only if user explicitly asks to wait
fai task poll <id> --interval 10s --timeout 5m
```

All commands accept `--json` for raw JSON output. Auth supports both `--api-key` (X-API-Key header) and `--token` (Authorization header, JWT). If neither is set, the CLI reads from config file.

### Library usage (Go)

```go
import "github.com/kentaylor/fromai/cli/client"

c := client.NewClient("http://127.0.0.1:8090", "")
c.SetAPIKey("your-api-key")

// Create a task
task, err := c.CreateTask(&client.CreateTaskRequest{
    Title:       "Implement sort",
    StarterCode: "function sort(arr) { }",
    Language:    "typescript",
})

// Default: report ID and continue
// Poll only if user explicitly asks to wait
// task, err = client.PollTask(c, task.ID, 5*time.Second, 10*time.Minute)
```

### Auth

Two mechanisms:
- **API key** (recommended): `fai init --key <key>` → stored in config → sent as `X-API-Key` header. Keys don't expire. Get your key from the `/settings` page.
- **JWT token**: `--token` flag or `FROMAI_TOKEN` env var → sent as `Authorization` header. Tokens expire after 120h.

API keys are auto-generated on signup and can be regenerated from the settings page.

See `cli/README.md` for full reference.

## Project Structure

```
fromai/
├── frontend/              # SvelteKit app (adapter-node)
│   ├── src/               # Source code
│   ├── static/            # Static assets
│   ├── package.json       # Frontend deps
│   └── ecosystem.config.js (root)
├── backend/               # PocketBase v0.39.0
│   ├── main.go            # Go server + custom routes
│   ├── pb_data/           # SQLite database
│   └── ecosystem.config.js
├── cli/                   # Go CLI (fai)
├── Caddyfile              # Reverse proxy config
└── docs...                # AGENTS.md, README.md, etc.
```

## Deployment

### PM2 (separate configs)

```bash
# Frontend (from project root)
cd frontend && pnpm run build && cd ..
pm2 start ecosystem.config.js

# Backend
cd backend && go build -o pocketbase main.go
pm2 start ecosystem.config.js
```

Both configs set env vars with placeholders — fill in secrets before deploying:
- `PB_ENCRYPTION_KEY`, `PUBLIC_URL`, `EXTERNAL_API_KEY` in `backend/ecosystem.config.js`

### Caddy

Reverse proxy routes `/api/*` → PocketBase `:8090`, everything else → frontend `:5173`. Change `:80` to your domain for auto-TLS.

## User Preferences

- **DO NOT auto-run**: Never run `pnpm run dev`, `go run`, `pm2 start`, or similar. User starts things manually.
- **Keep changes minimal**: Surgical edits only — don't refactor unrelated code
- **Simplicity first**: If a simpler approach exists, suggest it before implementing
- **Verify typecheck**: Run `cd frontend && pnpm run check` after TypeScript changes
- **Check memory**: Run `engram search` before debugging or implementing patterns that may have been solved before
