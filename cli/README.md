# fromai CLI

Command-line interface for the fromai task manager.

## Installation

```bash
cd cli && go build -o fai ./cmd/fai
```

## Authentication

### Quick Setup

```bash
# Get your API key from the web app: http://localhost:5173/settings
fai init --key "<your-api-key>"
```

This stores your key in `~/.config/fromai/config.toml` with secure permissions (0600).

### Verify Auth

```bash
fai whoami
```

### Manual/Advanced Usage

**API key** (recommended, doesn't expire):
```bash
fai --api-key "<your-api-key>" task list
```

**Token fallback** (expires after 120h):
```bash
export FROMAI_TOKEN="<your-token>"
fai task list
```

Or use the `--token` flag:
```bash
fai --token "<your-token>" task list
```

## Command Reference

### Task Operations

```bash
# Create a task
fai task create --title "Binary Search" --starter-code "func search..." --language "go"

# List all tasks
fai task list

# Get a task
fai task get <task-id>

# Update task code
fai task update <task-id> --code "func improved..."

# Submit task for grading
fai task submit <task-id>

# Grade a task
fai task grade <task-id> --grade "A" --feedback "Well done!"

# Archive a task (default)
fai task delete <task-id>

# Permanently delete a task
fai task delete <task-id> --hard

# Poll until status changes
fai task poll <task-id> --interval 2s --timeout 5m
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

    "github.com/kentaylor/fromai/cli/client"
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
