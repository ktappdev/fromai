# Coding Gym CLI

Command-line interface for the Coding Gym task manager.

## Installation

```bash
cd cli && go build -o cg ./cmd/cg
```

## Authentication

### Quick Setup

```bash
# Get your API key from the web app: http://localhost:5173/settings
cg init --key "<your-api-key>"
```

This stores your key in `~/.config/coding-gym/config.toml` with secure permissions (0600).

### Verify Auth

```bash
cg whoami
```

### Manual/Advanced Usage

**API key** (recommended, doesn't expire):
```bash
cg --api-key "<your-api-key>" task list
```

**Token fallback** (expires after 120h):
```bash
export CODING_GYM_TOKEN="<your-token>"
cg task list
```

Or use the `--token` flag:
```bash
cg --token "<your-token>" task list
```

## Command Reference

### Task Operations

```bash
# Create a task
cg task create --title "Binary Search" --starter-code "func search..." --language "go"

# List all tasks
cg task list

# Get a task
cg task get <task-id>

# Update task code
cg task update <task-id> --code "func improved..."

# Submit task for grading
cg task submit <task-id>

# Grade a task
cg task grade <task-id> --grade "A" --feedback "Well done!"

# Delete a task
cg task delete <task-id>

# Poll until status changes
cg task poll <task-id> --interval 2s --timeout 5m
```

### Global Flags

- `--api-key` - PocketBase API key (recommended, doesn't expire)
- `--token` - PocketBase auth token (expires after 120h)
- `--base-url` - Base URL (default: http://127.0.0.1:8090)
- `--json` - Output as raw JSON

## Library Usage (for AI agents)

```go
package main

import (
    "fmt"
    "time"

    "github.com/kentaylor/coding-gym/cli/client"
)

func main() {
    c := client.NewClient("http://127.0.0.1:8090", "")
    c.SetAPIKey("your-api-key")

    // List tasks
    tasks, err := c.ListTasks()
    if err != nil {
        panic(err)
    }
    fmt.Println(tasks)

    // Create task
    task, err := c.CreateTask(&client.CreateTaskRequest{
        Title:       "My Task",
        Description: "Description",
        StarterCode: "package main",
        Language:    "go",
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(task)

    // Poll for changes
    result, err := client.PollTask(c, task.ID, 5*time.Second, 10*time.Minute)
    if err != nil {
        panic(err)
    }
    fmt.Println(result)
}
```