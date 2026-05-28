# Coding Gym API Reference

Base URL: `http://localhost:8090`

---

## Authentication

There are two authentication modes:

### 1. Human Users (JWT)

PocketBase handles auth. Obtain a token via the built-in auth endpoint, then send it raw in the `Authorization` header.

**Note:** PocketBase v0.39.0 uses raw tokens. No `Bearer ` prefix.

```
Authorization: <jwt_token>
```

### 2. External Agents (API Key)

LLM agents and external services authenticate with a shared API key.

```
X-API-Key: <external_api_key>
```

The key is read from the `EXTERNAL_API_KEY` environment variable on the backend.

---

## Built-in Auth Endpoints (PocketBase)

### POST /api/collections/users/auth-with-password

Sign in an existing user.

**Body:**
```json
{
  "identity": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "token": "eyJhbG...",
  "record": {
    "id": "drxvuwj1f95lqa1",
    "email": "user@example.com",
    "name": ""
  }
}
```

### POST /api/collections/users/records

Register a new user.

**Body:**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "passwordConfirm": "password123",
  "name": "Alice"
}
```

### POST /api/collections/users/auth-refresh

Refresh the current session. Returns the user record.

---

## Human-Facing Custom Endpoints

All require `Authorization: <token>` header.

### GET /api/me

Returns the currently authenticated user.

**Response:**
```json
{
  "id": "drxvuwj1f95lqa1",
  "email": "user@example.com",
  "name": "",
  "created": "2026-05-28T18:32:05.385Z",
  "updated": "2026-05-28T18:32:05.385Z"
}
```

---

### GET /api/tasks

List all tasks for the authenticated user.

**Response:**
```json
[
  {
    "id": "dtc3c0mlsplkpcc",
    "title": "Reverse a linked list",
    "description": "...",
    "starter_code": "function reverseList(head) { ... }",
    "code": "function reverseList(head) { ... }",
    "language": "javascript",
    "status": "pending",
    "grade": "",
    "feedback": "",
    "user": "drxvuwj1f95lqa1",
    "created_at": 1779996316835,
    "updated_at": 1779996316835
  }
]
```

---

### GET /api/tasks/{id}

Get a single task. Must belong to the authenticated user.

**Response:** Same shape as list item.

**Errors:**
- `401` — Not authenticated
- `403` — Task does not belong to you
- `404` — Task not found

---

### POST /api/tasks

Create a new task for the authenticated user.

**Body:**
```json
{
  "title": "Two Sum",
  "description": "Find two numbers that add up to target.",
  "starter_code": "function twoSum(nums, target) {\n  // your code\n}",
  "language": "javascript"
}
```

**Allowed languages:** `typescript`, `javascript`, `python`, `go`, `rust`, `java`, `cpp`

**Response:** `201 Created` — the created task record.

---

### PATCH /api/tasks/{id}

Update the code for a task. Must belong to the authenticated user.

**Body:**
```json
{
  "code": "function twoSum(nums, target) {\n  return [0, 1];\n}"
}
```

**Response:** `200 OK` — the updated task record.

---

### POST /api/tasks/{id}/submit

Submit a task as completed. Must belong to the authenticated user.

Sets `status` to `"completed"`.

**Response:** `200 OK` — the updated task record.

---

### POST /api/tasks/{id}/grade

Grade a task. Must belong to the authenticated user.

**Body:**
```json
{
  "grade": "A-",
  "feedback": "Clean solution, O(n) time."
}
```

**Response:** `200 OK` — the updated task record.

---

## External LLM-Facing Endpoints

All require `X-API-Key: <key>` header. No JWT needed.

### GET /api/external/users

List all registered users. Useful for discovering who to assign tasks to.

**Response:**
```json
[
  {
    "id": "drxvuwj1f95lqa1",
    "email": "user@example.com",
    "name": "",
    "created": "2026-05-28T18:32:05.385Z"
  }
]
```

---

### POST /api/external/tasks

Create a task and assign it to a user.

**Body:**
```json
{
  "user_id": "drxvuwj1f95lqa1",
  "title": "Reverse a linked list",
  "description": "Implement a function that reverses a singly linked list in-place.",
  "starter_code": "function reverseList(head) {\n  // your code here\n  return head;\n}",
  "language": "javascript"
}
```

**Required fields:** `user_id`, `title`, `starter_code`, `language`

**Response:** `201 Created` — the created task record.

**Errors:**
- `400` — Missing `user_id` or invalid user
- `401` — Invalid or missing API key

---

### GET /api/external/tasks

List tasks with optional filtering.

**Query parameters:**
- `user_id` — Filter by assigned user
- `status` — Filter by status (`pending` | `completed`)

**Example:**
```
GET /api/external/tasks?user_id=drxvuwj1f95lqa1&status=completed
```

**Response:**
```json
[
  {
    "id": "dtc3c0mlsplkpcc",
    "title": "Reverse a linked list",
    "status": "completed",
    "code": "function reverseList(head) { ... }",
    "grade": "B+",
    "feedback": "Good effort...",
    "user": "drxvuwj1f95lqa1",
    ...
  }
]
```

---

### GET /api/external/tasks/{id}

Get a single task by ID, including its code.

**Response:** The full task record.

---

### POST /api/external/tasks/{id}/grade

Grade a completed task.

**Body:**
```json
{
  "grade": "A",
  "feedback": "Excellent solution with optimal time complexity."
}
```

**Response:** `200 OK` — the updated task record.

---

## Task Record Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | yes | Auto-generated UUID |
| `title` | string | yes | Task title |
| `description` | string | no | Task instructions |
| `starter_code` | string | yes | Initial code shown to user |
| `code` | string | yes | Current code (starts as `starter_code`) |
| `language` | string | yes | `typescript` \| `javascript` \| `python` \| `go` \| `rust` \| `java` \| `cpp` |
| `status` | string | yes | `pending` or `completed` |
| `grade` | string | no | Grading result (e.g. "A", "B+") |
| `feedback` | string | no | Grading feedback text |
| `user` | relation | yes | Assigned user ID |
| `created_at` | number | yes | Unix timestamp (ms) |
| `updated_at` | number | yes | Unix timestamp (ms) |

---

## Error Responses

All errors follow this shape:

```json
{
  "message": "Invalid or missing API key",
  "status": 401,
  "data": {}
}
```

| Status | Meaning |
|--------|---------|
| `400` | Bad Request — invalid body or missing required field |
| `401` | Unauthorized — missing/invalid auth token or API key |
| `403` | Forbidden — resource belongs to another user |
| `404` | Not Found — resource does not exist |
| `500` | Internal Server Error |

---

## CORS

All custom endpoints return these headers:

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization, X-API-Key
```

Preflight `OPTIONS` requests are handled automatically.
