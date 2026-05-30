# fromai

AI-powered coding task manager where users create tasks with starter code and edit them in a Monaco editor. SvelteKit frontend (Svelte 5 runes mode) talks to a PocketBase v0.39.0 Go backend. All data persists to SQLite via PocketBase.

**For AI agents**: See `skills/fromai/SKILL.md` for the complete skill guide on using the `fai` CLI, including async workflow, grading rubrics, and best practices.

## Architecture

This monorepo consists of three packages:

- **Frontend** (`src/`): SvelteKit application with Svelte 5 runes mode. Client-side only (no SSR). Uses PocketBase SDK for real-time subscriptions and auth.
- **Backend** (`backend/`): PocketBase v0.39.0 with custom Go routes for task management, grading, and external API integration.
- **CLI** (`cli/`): Go CLI tool (`fai`) for AI agents to create, list, update, submit, grade, and delete tasks. Both a binary and an importable library.

## Quick Start

### Prerequisites

- **Frontend**: Node.js 18+, pnpm
- **Backend**: Go 1.21+
- **CLI**: Go 1.21+ (for building)

### Setup

```bash
# Install frontend dependencies
pnpm install

# (Optional) Copy environment variables
cp .env.example .env.local
```

### Development

Both frontend and backend must run concurrently for full stack development:

```bash
# Terminal 1: Frontend (port 5173)
pnpm run dev

# Terminal 2: Backend (port 8090)
cd backend && go run main.go
```

The dev server will automatically open your browser to `http://localhost:5173`.

### Production Build

```bash
# Build frontend
pnpm run build

# Preview production build
pnpm run preview
```

For production, run the PocketBase server with the production configuration (set `PB_ENCRYPTION_KEY` before first run to encrypt secrets at rest).

## CLI Installation and Usage

### Build

```bash
cd cli && go build -o fai ./cmd/fai
```

### Setup (one-time)

```bash
# Get your API key from the fromai settings page (one-time human setup)
fai init --key "your-api-key"
```

This stores the key in `~/.config/fromai/config.toml`. Subsequent commands read it automatically.

### Commands

```bash
# Verify auth
fai whoami

# Create a task
fai task create --title "Sort Array" --starter-code "// TODO" --language typescript

# List tasks
fai task list

# Get task details
fai task get <id>

# Update task code
fai task update <id> --code "// new code"

# Submit task (human action)
fai task submit <id>

# Grade task
fai task grade <id> --grade "A" --feedback "nice work"

# Delete task (archive by default)
fai task delete <id>

# Permanent delete
fai task delete <id> --hard

# Poll for completion (only if explicitly requested)
fai task poll <id>
fai task poll <id> --interval 10s --timeout 5m
```

All commands accept `--json` for raw JSON output.

### Auth

Two mechanisms:

- **API key** (recommended): `fai init --key <key>` → stored in config → sent as `X-API-Key` header. Keys don't expire. Get your key from the `/settings` page.
- **JWT token**: `--token` flag or `FROMAI_TOKEN` env var → sent as `Authorization` header. Tokens expire after 120h.

### Library Usage (Go)

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
```

## Environment Variables

### Frontend (Vite)

- `VITE_POCKETBASE_URL`: PocketBase server URL (default: `http://127.0.0.1:8090`). Must be prefixed with `VITE_` to be exposed to client-side code.

### Backend (Go)

- `PUBLIC_URL`: Base URL for the external help handler (default: `http://localhost:8090`).

### CLI (Go)

- `FROMAI_BASE_URL`: Base URL for the API client (default: `http://127.0.0.1:8090`).
- `FROMAI_TOKEN`: JWT token for authentication (alternative to API key).

Priority for base URL resolution: `--base-url` flag > `FROMAI_BASE_URL` env var > config file > hardcoded default.

## Documentation

- **Agent Skill Guide**: `skills/fromai/SKILL.md` — Complete guide for AI agents using the `fai` CLI
- **API Reference**: `API_REFERENCE.md` — Full API documentation for backend endpoints
- **PocketBase Gotchas**: `POCKETBASE_GOTCHAS.md` — Critical PocketBase v0.39.0 migration notes and production deployment gotchas
- **Collection Setup**: `SETUP_POCKETBASE.md` — Step-by-step PocketBase collection creation

## Coding Conventions

### Frontend

- Svelte 5 runes mode (`$state`, `$derived`, etc.)
- Client-side only (no SSR, no `+page.server.ts`)
- Guard `localStorage` with `typeof window !== 'undefined'`
- Use `PocketBaseClient` class from `src/lib/pocketbase.ts` — not raw fetch
- Auth token stored in localStorage as raw string (no "Bearer " prefix)
- Monaco editor loaded from CDN, not bundled

### Backend

- All custom routes in `backend/main.go` (~180 lines)
- Routes use `apis.RequireAuth()` wrapper with `.Bind()` for auth
- Read authenticated user via `e.Auth`
- CORS set manually per handler
- Single PocketBase binary with embedded assets

### CLI

- Go 1.21+ with cobra for CLI
- Config stored in `~/.config/fromai/config.toml`
- Library can be imported as `github.com/kentaylor/fromai/cli/client`

## License

MIT License — Copyright (c) 2025, Kent Taylor

This project is open source and welcomes contributions. The MIT license requires attribution — please preserve the copyright notice in all copies.