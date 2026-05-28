# PocketBase Go Framework Gotchas (v0.39.0)

This document captures critical learnings from migrating to PocketBase as a Go framework. **Version matters enormously** — the API changed significantly between v0.23.x and v0.39.0.

---

## 1. Auth Header Format (Breaking Change)

**Old:** `Authorization: Bearer <token>`  
**New:** `Authorization: <token>` (raw token, no prefix)

PocketBase v0.39.0 simplified auth headers. The "Bearer ", "Admin ", and "User " prefixes are no longer required or expected. The token type is auto-detected from the JWT payload.

**Frontend impact:** Remove `Bearer ` prefix from all requests.
```typescript
// WRONG
headers['Authorization'] = `Bearer ${token}`;

// CORRECT
headers['Authorization'] = token;
```

**Backend impact:** Don't strip prefixes in custom handlers.
```go
// WRONG
token := authHeader
if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
    token = authHeader[7:]
}

// CORRECT (just use the header directly)
token := e.Request.Header.Get("Authorization")
```

---

## 2. Getting the Authenticated User

### The `e.Auth` Field

`core.RequestEvent` has an `Auth *Record` field, but **it is only populated if auth middleware is bound to the route**.

```go
func getAuthUser(e *core.RequestEvent) *core.Record {
    return e.Auth  // nil if no auth middleware applied
}
```

### Applying Auth Middleware

Use `apis.RequireAuth()` bound via `.Bind()`:

```go
se.Router.GET("/api/tasks", listTasksHandler(app)).Bind(apis.RequireAuth())
se.Router.POST("/api/tasks", createTaskHandler(app)).Bind(apis.RequireAuth())
```

Without this middleware, `e.Auth` is always `nil` and your custom auth validation will fail.

### Manual Token Validation (Discouraged)

If you must validate manually, use `app.FindAuthRecordByToken(token, maxAge)` where maxAge is a duration string like `"24h"` or `""` for default. But prefer `e.Auth` when middleware is available.

---

## 3. App Methods Replaced `app.Dao()`

**Gone:** `app.Dao().FindRecordById()`, `app.Dao().SaveRecord()`, etc.  
**New:** Methods exist directly on `app`:

```go
// WRONG
record, err := app.Dao().FindRecordById("tasks", id)
err = app.Dao().SaveRecord(record)

// CORRECT
record, err := app.FindRecordById("tasks", id)
err = app.Save(record)
```

Key methods on `*pocketbase.PocketBase`:
- `app.FindRecordById(collection, id)`
- `app.FindRecordsByFilter(collection, filter, sort, limit, offset, params)`
- `app.FindCollectionByNameOrId(name)`
- `app.Save(record)` / `app.SaveNoValidate(record)`
- `app.FindAuthRecordByToken(token, maxAge)`
- `app.Delete(model)`

---

## 4. Record Creation and Field Access

### Creating Records

```go
// WRONG (old daos package)
record := daos.NewRecord(collection)
record.SetDataValue("title", "foo")

// CORRECT
collection, err := app.FindCollectionByNameOrId("tasks")
record := core.NewRecord(collection)
record.Set("title", "foo")
```

### Reading Fields

```go
// WRONG
record.GetStringDataValue("user")

// CORRECT
record.GetString("user")
record.GetInt("created_at")
```

---

## 5. Request Handling Changes

### Path Parameters

```go
// WRONG
taskId := e.PathParam("id")

// CORRECT
taskId := e.Request.PathValue("id")
```

### Reading JSON Body

```go
// WRONG (no apis.Bind or e.BindJSON)
var data map[string]any
if err := apis.Bind(e.Request, &data); err != nil { ... }

// CORRECT
var data map[string]any
if err := json.NewDecoder(e.Request.Body).Decode(&data); err != nil { ... }
```

### Response Helpers

```go
// Still works
return e.JSON(http.StatusOK, data)
return e.String(http.StatusOK, "Hello")
```

---

## 6. Route Registration

### Hook Registration

```go
// WRONG (old Add method)
app.OnServe().Add(func(e *core.ServeEvent) error { ... })

// CORRECT (BindFunc)
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
    se.Router.GET("/hello", func(e *core.RequestEvent) error {
        return e.String(200, "Hello world!")
    })
    return se.Next()
})
```

### Handler Signatures

```go
// WRONG (old RequestHandlerFunc type)
func listTasksHandler(app *pocketbase.PocketBase) core.RequestHandlerFunc {
    return func(e *core.RequestEvent) error { ... }
}

// CORRECT (plain function type)
func listTasksHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
    return func(e *core.RequestEvent) error { ... }
}
```

---

## 7. Settings API Changes

Field names were renamed from `User*` to `Record*`:

```go
// WRONG
app.Settings().UserAuthToken.Duration

// CORRECT
app.Settings().RecordAuthToken.Duration
```

Other renamed fields:
- `UserAuthToken` → `RecordAuthToken`
- `UserPasswordResetToken` → `RecordPasswordResetToken`
- `UserEmailChangeToken` → `RecordEmailChangeToken`
- `UserVerificationToken` → `RecordVerificationToken`

---

## 8. Collection Names

The users auth collection has a system name:

```go
// System collection ID for users
"_pb_users_auth_"  // or just "users" in most API calls
```

When creating relation fields pointing to users, use `"users"` as the collection name.

---

## 9. Running the Server

```bash
# WRONG (just compiles and exits)
go run main.go

# CORRECT (starts the web server)
go run main.go serve
```

Default binds to `127.0.0.1:8090`.

---

## 10. Superuser Creation

```bash
# Create admin user via CLI
go run main.go superuser upsert admin@example.com password123
```

Admin auth endpoint:
```
POST /api/collections/_superusers/auth-with-password
{"identity": "admin@example.com", "password": "password123"}
```

---

## 11. CORS Handling

PocketBase v0.39.0 doesn't automatically add CORS headers to custom routes. You must set them manually:

```go
func setCORSHeaders(w http.ResponseWriter) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}
```

Call this at the start of every custom handler. The built-in OPTIONS handling works for preflight requests.

---

## 12. Frontend Auth Response Structure

Auth responses now use `record` instead of `user`:

```json
{
  "record": { "id": "...", "email": "..." },
  "token": "jwt_token_here"
}
```

**Not** `user` anymore. Update frontend code accordingly.

---

## 13. No `dbx.Params` for FindRecordsByFilter

```go
// WRONG
app.FindRecordsByFilter("tasks", "user = {:userId}", "-created", 100, 0, dbx.Params{"userId": id})

// CORRECT
app.FindRecordsByFilter("tasks", "user = {:userId}", "-created", 100, 0, map[string]any{"userId": id})
```

---

## 14. Common Error Messages and Fixes

| Error | Cause | Fix |
|-------|-------|-----|
| `Not authenticated` (401) | `e.Auth` is nil | Add `.Bind(apis.RequireAuth())` to route |
| `bind: address already in use` | Old process still running | `pkill -f "go run main.go"` or `lsof -ti:8090 \| xargs kill -9` |
| `undefined: core.RequestHandlerFunc` | Using old type name | Use `func(*core.RequestEvent) error` |
| `app.Dao undefined` | Old API | Use `app.FindRecordById()` etc. directly |
| `e.PathParam undefined` | Old method | Use `e.Request.PathValue()` |

---

## 15. Architecture Pattern for Custom + Default Routes

Best practice when extending PocketBase with custom Go routes:

1. **Auth** → Use PocketBase's built-in collection auth endpoints (`/api/collections/users/auth-with-password`, etc.)
2. **Custom business logic** → Register custom routes with `apis.RequireAuth()` middleware
3. **Data access** → Use `app.FindRecordById()`, `app.Save()`, etc.
4. **Auth state** → Read from `e.Auth` in handlers (after middleware is applied)

This keeps auth standardized while allowing custom business logic in Go.

---

## 16. Checklist for Future Edits

When modifying this PocketBase backend:

- [ ] Are routes protected with `.Bind(apis.RequireAuth())`?
- [ ] Is `e.Auth` used instead of manual token parsing?
- [ ] Are CORS headers set in custom handlers?
- [ ] Is `json.NewDecoder()` used for request body parsing?
- [ ] Is `e.Request.PathValue()` used for path params?
- [ ] Are app methods used directly (no `app.Dao()`)?
- [ ] Is `core.NewRecord()` used for new records?
- [ ] Are fields accessed with `.GetString()`, `.GetInt()`, etc.?
- [ ] Does the frontend send raw token (no `Bearer ` prefix)?
- [ ] Is auth response expecting `.record` not `.user`?

---

## 17. Useful Resources

- Context7 `/pocketbase/pocketbase` for API docs
- `go doc github.com/pocketbase/pocketbase/core` for type signatures
- `go doc github.com/pocketbase/pocketbase/apis` for middleware helpers
- Admin UI: `http://127.0.0.1:8090/_/` (create first superuser via CLI)

---

*Document created after migrating from Convex to PocketBase v0.39.0 Go framework. Last updated: 2026-05-28.*

---

## Production Readiness — Critical Gotchas

These items matter most when deploying to production. Many are footguns that only surface under real load.

### 18. SQLITE_BUSY Under Parallel Requests

SQLite serializes writes with a mutex. Under burst parallel reads (e.g., `Promise.all()` firing multiple requests), PocketBase may return `404 {"errorDetails": "database is locked (5) (SQLITE_BUSY)"}`. The logs DB write path is a common culprit.

**Mitigations:**
- Reduce client-side request fan-out — avoid `Promise.all()` for >3-4 concurrent PocketBase calls
- Consider reducing log verbosity in production (Admin UI > Settings > Logs)
- PocketBase v0.39.0 uses `modernc.org/sqlite` (pure Go driver) — CGO driver may perform better under write-heavy loads but requires CGO_ENABLED=1

### 19. Backup Size Explosion & WAL Growth

The built-in ZIP backup (`POST /api/backups`) uses a **write-blocking transaction** that triggers a SQLite WAL checkpoint. If long-running transactions are active, the WAL file balloons dramatically — backups can be 10-100x larger than the actual database.

**Rules:**
- **Never** use `fs.copyFile()` or OS-level `cp` on pb_data while PocketBase is running — produces silently corrupted backups (WAL file desyncs from main db)
- Ensure at least **2x disk space** of pb_data directory before running backups
- For pb_data > 2GB: use `sqlite3 pb_data/data.db ".backup 'backup.db'"` + rsync instead of built-in ZIP (avoids write lock)
- Run `PRAGMA optimize;` periodically via v0.39.0 SQL console (Settings > Debug) to keep query planner stats fresh

### 20. Token Lifetime & Invalidation Confusion

**Tokens are NOT invalidated on server restart.** This is a common misconception. Tokens remain valid until they naturally expire or:
- User changes password
- User changes email
- Token secret is rotated (Admin UI > Collections > Users > Options)
- Superuser manually deletes user

**Rotating the token secret invalidates ALL tokens for ALL users** — not per-user. There is no built-in per-user token revocation.

**Token expiry default**: Configurable in `app.Settings().RecordAuthToken.Duration`. Default is 120h (5 days). Consider shortening in production.

### 21. Settings Encryption at Rest

Without `--encryptionEnv=ENCRYPTION_KEY` or `PB_ENCRYPTION_KEY` env var, SMTP passwords, S3 credentials, and OAuth2 client secrets are stored as **plaintext JSON** in the SQLite database. Anyone with filesystem access to pb_data can read them.

```bash
# Set a 32-char encryption key before first run
export PB_ENCRYPTION_KEY=$(openssl rand -hex 16)
```

**Cannot be retrofitted** — if you've already stored secrets without encryption, you must re-save them after setting the key.

### 22. CORS: Use apis.CORS() Not Manual Headers

Current code sets CORS headers manually in every handler. The framework provides `apis.CORS()` middleware that handles preflight OPTIONS correctly:

```go
// Better: register CORS once on route group
se.Router.Group("/api/tasks", func(g *core.RouterGroup) {
    g.GET("", listTasksHandler(app))
    g.GET("/{id}", getTaskHandler(app))
    g.POST("", createTaskHandler(app))
    g.PATCH("/{id}", updateTaskHandler(app))
    g.POST("/{id}/submit", submitTaskHandler(app))
    g.POST("/{id}/grade", gradeTaskHandler(app))
}).Bind(
    apis.RequireAuth(),
    apis.CORS(apis.CORSConfig{
        AllowOrigins: []string{"http://localhost:5173"},
        AllowMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders: []string{"Content-Type", "Authorization"},
    }),
)
```

**Security**: Never use `AllowCredentials: true` with `AllowOrigins: ["*"]`.

### 23. Rate Limiting (Built-in Since v0.38)

PocketBase v0.38+ has configurable rate limiting. Enable it in production:

```go
app.Settings().RateLimits.Enabled = true
// Default rules in Admin UI > Settings > Application:
// /api/ — 300 req/10s, auth endpoints — 3 req/1s
```

No built-in IP exemption for internal services — if your own backend needs to bypass limits, implement custom middleware.

### 24. System Resource Limits

**Always in production:**
```bash
# Increase open file descriptors (SQLite needs them)
ulimit -n 4096

# Prevent OOM in memory-constrained environments
export GOMEMLIMIT=512MiB

# Enable swap as safety net
```

**systemd service** should include:
```ini
LimitNOFILE=4096
Environment=GOMEMLIMIT=512MiB
```

### 25. SQLite Single-Writer Bottleneck

SQLite serializes ALL writes through a single mutex. Under heavy concurrent writes (bulk inserts, many users creating records simultaneously), throughput degrades significantly. PocketBase wraps this with its own mutex.

**Signs**: increasing request latency under concurrent POST/PATCH, 503 errors.
**Limit**: ~50-100 concurrent write requests/second on typical hardware.
**Mitigation**: batch writes where possible, use `POST /api/batch` endpoint for multi-record operations.

### 26. Reverse Proxy & TLS


PocketBase serves plain HTTP. For production, put it behind nginx/Caddy:

```nginx
# nginx critical directives
location / {
    proxy_pass http://127.0.0.1:8090;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_read_timeout 120s;  # long-polling / realtime support
}
```

**Configure trusted proxy** in Admin UI > Settings so `RealIP()` works correctly — otherwise you'll log proxy IPs instead of client IPs.

### 27. Graceful Shutdown

PocketBase handles `SIGINT`/`SIGTERM` gracefully but ongoing requests may be cut off. For zero-downtime deploys:

- Use nginx `proxy_smooth_drain` or Caddy graceful shutdown
- Wait for active connections to drain before killing process
- systemd: `TimeoutStopSec=30s`

### 28. Disk Space & Corruption


**Running out of disk space** causes "malformed database" errors. SQLite cannot recover from writes that fail mid-transaction due to ENOSPC.

- Monitor disk space aggressively in production
- Set alerts at 20% free space
- Keep at least 2x pb_data size as free buffer
- Network-mounted drives (NFS, SMB) are **dangerous** for SQLite — don't use them for pb_data

### 29. Migrations Workflow


PocketBase auto-migrates pb_data on startup when the binary version changes. But for intentional schema changes:

```bash
# Generate migration snapshot of current schema
./your-binary migrate collections

# This creates pb_migrations/ directory with auto-migration files
# Commit these to version control
```

Custom Go code changes (new collections, field changes in `main.go`) should be paired with migration snapshots. Without migrations, schema changes only live in the running database, not in version control.

---

## 30. Realtime Subscriptions via JavaScript SDK

PocketBase supports live updates through Server-Sent Events (SSE). The official JS SDK abstracts this with `subscribe()` / `unsubscribe()`:

```typescript
import PocketBase from 'pocketbase';
const pb = new PocketBase('http://127.0.0.1:8090');

// Subscribe to all changes in a collection
const unsub = await pb.collection('tasks').subscribe('*', (e) => {
    console.log(e.action); // 'create' | 'update' | 'delete'
    console.log(e.record); // the changed record
});

// Subscribe to a specific record
const unsub2 = await pb.collection('tasks').subscribe('RECORD_ID', (e) => {
    console.log(e.record);
});

// Cleanup
await unsub();
await pb.collection('tasks').unsubscribe(); // unsub all
```

**Important:**
- `subscribe()` returns a Promise that resolves to an unsubscribe function. Must `await` it.
- The callback receives `{action, record}` — check `action` to decide how to update local state.
- If auth is required (which it is for the `tasks` collection), the SDK must have a valid token in `pb.authStore` before subscribing.
- Subscriptions are per-tab. If you need cross-tab sync, use `BroadcastChannel` or re-fetch on `visibilitychange`.
- PocketBase v0.39.0 backend uses SSE over `GET /api/realtime`. Reverse proxies must support long-lived connections (`proxy_read_timeout 120s` in nginx).
