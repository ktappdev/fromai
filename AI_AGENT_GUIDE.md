# Coding Gym — AI Agent Integration Guide

This document is for AI coding agents (LLMs, autonomous bots, task routers) that need to send coding challenges to humans, collect submissions, and provide feedback.

---

## What This Platform Does

Coding Gym is a task-based coding practice platform. As an AI agent, your role is:

1. **Discover users** on the platform
2. **Create coding tasks** and assign them to users
3. **Poll for completion** — wait for the human to finish
4. **Retrieve the submitted code**
5. **Grade the submission** and provide feedback

The human works in a browser-based Monaco editor. You interact entirely via HTTP API.

---

## Base Configuration

| Setting | Value |
|---------|-------|
| Base URL | `http://localhost:8090` |
| Auth header | `X-API-Key: <your_api_key>` |
| Content-Type | `application/json` |

The `EXTERNAL_API_KEY` is set by the human running the backend. Ask them for the current key.

---

## Full Agent Workflow

### Step 1: Discover Users

Before sending tasks, find out who is on the platform.

```bash
curl http://localhost:8090/api/external/users \
  -H "X-API-Key: dev-key-123"
```

**Response:**
```json
[
  {
    "id": "drxvuwj1f95lqa1",
    "email": "user@example.com",
    "name": ""
  }
]
```

Save the `id` — you need it for every task creation.

---

### Step 2: Create a Task

Send a coding challenge to a user.

```bash
curl -X POST http://localhost:8090/api/external/tasks \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-key-123" \
  -d '{
    "user_id": "drxvuwj1f95lqa1",
    "title": "Reverse a Linked List",
    "description": "Implement a function that reverses a singly linked list in-place. Return the new head. Handle edge cases like empty list and single node.",
    "starter_code": "function reverseList(head) {\n  // your code here\n  return head;\n}",
    "language": "javascript"
  }'
```

**Required fields:**
- `user_id` — the target user's PocketBase ID (from Step 1)
- `title` — short task name
- `starter_code` — initial code shown in the editor
- `language` — one of: `typescript`, `javascript`, `python`, `go`, `rust`, `java`, `cpp`

**Optional:**
- `description` — full instructions (supports newlines)

**Response:** `201 Created`
```json
{
  "id": "dtc3c0mlsplkpcc",
  "title": "Reverse a Linked List",
  "description": "...",
  "status": "pending",
  "user": "drxvuwj1f95lqa1",
  ...
}
```

Save the returned `id` — this is the task ID you poll against.

---

### Step 3: Poll for Completion

The human receives the task instantly (the frontend uses PocketBase real-time subscriptions). They work in the editor and click **Finish** when done.

Poll every 30–60 seconds:

```bash
curl "http://localhost:8090/api/external/tasks?user_id=drxvuwj1f95lqa1&status=completed" \
  -H "X-API-Key: dev-key-123"
```

Or poll the specific task:

```bash
curl http://localhost:8090/api/external/tasks/dtc3c0mlsplkpcc \
  -H "X-API-Key: dev-key-123"
```

Look for `"status": "completed"`.

**Tip:** You can poll multiple users at once by omitting `user_id`, or check only pending tasks with `status=pending`.

---

### Step 4: Retrieve the Submission

Once status is `completed`, get the full task to read the submitted code.

```bash
curl http://localhost:8090/api/external/tasks/dtc3c0mlsplkpcc \
  -H "X-API-Key: dev-key-123"
```

**Response:**
```json
{
  "id": "dtc3c0mlsplkpcc",
  "title": "Reverse a Linked List",
  "code": "function reverseList(head) {\n  let prev = null;\n  let curr = head;\n  while (curr) {\n    const next = curr.next;\n    curr.next = prev;\n    prev = curr;\n    curr = next;\n  }\n  return prev;\n}",
  "language": "javascript",
  "status": "completed",
  "grade": "",
  "feedback": "",
  "user": "drxvuwj1f95lqa1"
}
```

The `code` field contains the human's final submission.

---

### Step 5: Grade the Task

Evaluate the code and send back a grade with feedback.

```bash
curl -X POST http://localhost:8090/api/external/tasks/dtc3c0mlsplkpcc/grade \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-key-123" \
  -d '{
    "grade": "A",
    "feedback": "Clean iterative solution. O(n) time, O(1) space. Well done."
  }'
```

**Fields:**
- `grade` — any string (e.g. `"A"`, `"B+"`, `"Pass"`, `"Needs Work"`)
- `feedback` — detailed textual feedback

The human sees the grade and feedback immediately in the frontend (via real-time subscription).

---

## Pseudocode: Complete Agent Loop

```python
api_key = "dev-key-123"
base_url = "http://localhost:8090"

def run_agent():
    # 1. Find users
    users = get(f"{base_url}/api/external/users", headers={"X-API-Key": api_key})
    target_user = users[0]["id"]

    # 2. Create a task
    task = post(f"{base_url}/api/external/tasks", json={
        "user_id": target_user,
        "title": "Two Sum",
        "description": "Find two numbers in the array that add up to the target.",
        "starter_code": "function twoSum(nums, target) {\n  // your code\n}",
        "language": "javascript"
    }, headers={"X-API-Key": api_key})
    task_id = task["id"]

    # 3. Poll until completed
    while True:
        sleep(30)
        task = get(f"{base_url}/api/external/tasks/{task_id}",
                   headers={"X-API-Key": api_key})
        if task["status"] == "completed":
            break

    # 4. Retrieve code
    submitted_code = task["code"]

    # 5. Evaluate and grade
    grade, feedback = evaluate_code(submitted_code)
    post(f"{base_url}/api/external/tasks/{task_id}/grade", json={
        "grade": grade,
        "feedback": feedback
    }, headers={"X-API-Key": api_key})
```

---

## Error Handling

| Scenario | HTTP Status | What to Do |
|----------|-------------|------------|
| Invalid API key | `401` | Ask the human for the correct key |
| Invalid `user_id` | `400` | Re-run user discovery |
| Task not found | `404` | The task ID may be wrong or deleted |
| Backend unreachable | connection error | The server is not running — ask the human to start it |

---

## Design Tips for Agents

- **Be specific in descriptions.** The human only sees what you write. Include constraints, expected behavior, and edge cases.
- **Choose the right language.** Match the human's preference if known. Default to `javascript` or `typescript`.
- **Starter code matters.** Provide a function signature and comments so the human knows where to write.
- **Polling frequency.** 30 seconds is polite. Don't hammer the server.
- **Graceful degradation.** If grading fails, log the error and retry once. The human is waiting.

---

## API Quick Reference

| Action | Method | Endpoint | Auth |
|--------|--------|----------|------|
| List users | GET | `/api/external/users` | `X-API-Key` |
| Create task | POST | `/api/external/tasks` | `X-API-Key` |
| List tasks | GET | `/api/external/tasks?user_id=&status=` | `X-API-Key` |
| Get task | GET | `/api/external/tasks/{id}` | `X-API-Key` |
| Grade task | POST | `/api/external/tasks/{id}/grade` | `X-API-Key` |

See `API_REFERENCE.md` for full request/response schemas and human-facing endpoints.
