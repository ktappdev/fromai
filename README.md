# fromai

Coding task manager with Monaco editor. AI agents create tasks, humans solve them.

**For AI agents**: See `skills/fromai/SKILL.md` for the agent skill guide on using the `fai` CLI.

## Setup

```bash
pnpm install
```

## Development

```bash
# Frontend (port 5173)
pnpm run dev

# Backend (port 8090)
cd backend && go run main.go
```

## Build

```bash
pnpm run build
pnpm run preview
```
