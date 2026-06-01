package main

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/kentaylor/fromai/cli/client"
	"github.com/kentaylor/fromai/cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	flagToken   string
	flagAPIKey  string
	flagBaseURL string
	flagJSON    bool
)

func resolveToken() string {
	if flagToken != "" {
		return flagToken
	}
	if token := os.Getenv("FROMAI_TOKEN"); token != "" {
		return token
	}
	return ""
}

func resolveClient() *client.Client {
	token := resolveToken()
	apiKey := flagAPIKey

	// Resolve base URL: flag > env var > config file > default
	baseURL := flagBaseURL
	if baseURL == "https://fromai-backend.lyricut.com" {
		if envURL := os.Getenv("FROMAI_BASE_URL"); envURL != "" {
			baseURL = envURL
		}
	}
	if baseURL == "https://fromai-backend.lyricut.com" {
		cfg, err := config.Load()
		if err == nil && cfg.BaseURL != "" {
			baseURL = cfg.BaseURL
		}
	}

	// Try config file for API key
	if token == "" && apiKey == "" {
		cfg, err := config.Load()
		if err == nil && cfg.APIKey != "" {
			apiKey = cfg.APIKey
		}
	}

	c := client.NewClient(baseURL, token)
	if apiKey != "" {
		c.SetAPIKey(apiKey)
	}
	return c
}

func isUnauthorized(err error) bool {
	return err != nil && strings.Contains(err.Error(), "401")
}

func printError(cmd *cobra.Command, err error) {
	if isUnauthorized(err) {
		fmt.Fprintln(os.Stderr, "Authentication failed. Run 'fai init --key <your-api-key>' or set FROMAI_TOKEN.")
	} else {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func printTask(cmd *cobra.Command, task *client.Task) {
	if flagJSON {
		data, _ := json.MarshalIndent(task, "", "  ")
		fmt.Println(string(data))
		return
	}
	fmt.Printf("ID:          %s\n", task.ID)
	fmt.Printf("Title:       %s\n", task.Title)
	fmt.Printf("Status:      %s\n", task.Status)
	fmt.Printf("Language:    %s\n", task.Language)
	fmt.Printf("Grade:       %s\n", task.Grade)
	fmt.Printf("Description: %s\n", html.UnescapeString(task.Description))
	fmt.Printf("StarterCode: %s\n", html.UnescapeString(task.StarterCode))
}

func printTaskTable(tasks []client.Task) {
	if flagJSON {
		data, _ := json.MarshalIndent(tasks, "", "  ")
		fmt.Println(string(data))
		return
	}
	fmt.Println("ID | TITLE | STATUS | LANGUAGE | GRADE")
	fmt.Println(strings.Repeat("-", 80))
	for _, t := range tasks {
		fmt.Printf("%s | %s | %s | %s | %s\n", t.ID, t.Title, t.Status, t.Language, t.Grade)
	}
}

var rootCmd = &cobra.Command{
	Use:   "fai",
	Short: "fromai CLI",
	Long: `Create coding tasks for humans to solve. Typical agent workflow:

  1. fai task create --title "Sort Array" --starter-code "function sort(arr) {}" --language typescript
  2. Report task ID to user and continue with other work
  3. fai task grade <id> --grade "A" --feedback "clean work" (when user asks)

Only use poll if user explicitly asks to wait:
  fai task poll <id>

Also usable as a Go library: import "github.com/kentaylor/fromai/cli/client"`,
}

var taskCmd = &cobra.Command{
	Use:       "task",
	Short:     "Task operations",
	SuggestFor: []string{"tasks", "tsk", "taks"},
}

var createCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a new task",
	Long:    `Create a new coding task for a human to solve. Requires title, starter code, and language.`,
	Example: `  fai task create --title "Reverse String" --starter-code "function reverse(s) {}" --language typescript
  fai task create --title "Sort" --description "Implement quicksort" --starter-code "// TODO" --language go --json`,
	Run: func(cmd *cobra.Command, args []string) {
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		starterCode, _ := cmd.Flags().GetString("starter-code")
		language, _ := cmd.Flags().GetString("language")

		req := &client.CreateTaskRequest{
			Title:       title,
			Description: description,
			StarterCode: starterCode,
			Language:    language,
		}

		task, err := resolveClient().CreateTask(req)
		if err != nil {
			printError(cmd, err)
			os.Exit(1)
		}
		printTask(cmd, task)
	},
}

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all tasks",
	Long:    `List all tasks assigned to the authenticated user.`,
	Example: `  fai task list
  fai task list --json`,
	Run: func(cmd *cobra.Command, args []string) {
		tasks, err := resolveClient().ListTasks()
		if err != nil {
			printError(cmd, err)
			os.Exit(1)
		}
		printTaskTable(tasks)
	},
}

var getCmd = &cobra.Command{
	Use:     "get <id>",
	Short:   "Get a task by ID",
	Long:    `Retrieve a single task by its ID.`,
	Example: `  fai task get abc123
  fai task get abc123 --json`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		task, err := resolveClient().GetTask(args[0])
		if err != nil {
			printError(cmd, err)
			os.Exit(1)
		}
		printTask(cmd, task)
	},
}

var updateCmd = &cobra.Command{
	Use:     "update <id>",
	Short:   "Update a task's code",
	Long:    `Update the code field of an existing task.`,
	Example: `  fai task update abc123 --code "function solve() { return 42; }"`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		code, _ := cmd.Flags().GetString("code")
		if code == "" {
			fmt.Fprintln(os.Stderr, "Error: --code is required")
			os.Exit(1)
		}

		req := &client.UpdateTaskRequest{Code: code}
		task, err := resolveClient().UpdateTask(args[0], req)
		if err != nil {
			printError(cmd, err)
			os.Exit(1)
		}
		printTask(cmd, task)
	},
}

var submitCmd = &cobra.Command{
	Use:     "submit <id>",
	Short:   "Submit a task for grading",
	Long:    `Submit a task for grading — marks it as completed.`,
	Example: `  fai task submit abc123`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		task, err := resolveClient().SubmitTask(args[0])
		if err != nil {
			printError(cmd, err)
			os.Exit(1)
		}
		printTask(cmd, task)
	},
}

var gradeCmd = &cobra.Command{
	Use:     "grade <id>",
	Short:   "Grade a task",
	Long:    `Grade a completed task and optionally provide feedback.`,
	Example: `  fai task grade abc123 --grade "A" --feedback "clean solution, good edge cases"`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		grade, _ := cmd.Flags().GetString("grade")
		if grade == "" {
			fmt.Fprintln(os.Stderr, "Error: --grade is required")
			os.Exit(1)
		}
		feedback, _ := cmd.Flags().GetString("feedback")

		req := &client.GradeRequest{Grade: grade, Feedback: feedback}
		task, err := resolveClient().GradeTask(args[0], req)
		if err != nil {
			printError(cmd, err)
			os.Exit(1)
		}
		printTask(cmd, task)
	},
}

var deleteCmd = &cobra.Command{
	Use:     "delete <id>",
	Short:   "Archive or permanently delete a task",
	Long:    `Archives a task by default. Use --hard to permanently delete it. Only the owner can remove their own tasks.`,
	Example: `  fai task delete abc123          # archive
  fai task delete abc123 --hard   # permanent delete`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hard, _ := cmd.Flags().GetBool("hard")
		if hard {
			resp, err := resolveClient().DeleteTask(args[0])
			if err != nil {
				printError(cmd, err)
				os.Exit(1)
			}
			if flagJSON {
				data, _ := json.MarshalIndent(resp, "", "  ")
				fmt.Println(string(data))
				return
			}
			fmt.Println("Task deleted permanently")
			return
		}
		task, err := resolveClient().ArchiveTask(args[0])
		if err != nil {
			printError(cmd, err)
			os.Exit(1)
		}
		if flagJSON {
			data, _ := json.MarshalIndent(task, "", "  ")
			fmt.Println(string(data))
			return
		}
		fmt.Println("Task archived")
	},
}

var pollCmd = &cobra.Command{
	Use:     "poll <id>",
	Short:   "Poll a task until status changes",
	Long:    `Poll a task until its status changes (e.g., human marks it complete). Blocks until change detected or timeout reached.`,
	Example: `  fai task poll abc123
  fai task poll abc123 --interval 10s --timeout 5m
  fai task poll abc123 --json | jq .status`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		interval, _ := cmd.Flags().GetDuration("interval")
		timeout, _ := cmd.Flags().GetDuration("timeout")

		task, err := client.PollTask(resolveClient(), args[0], interval, timeout)
		if err != nil {
			printError(cmd, err)
			os.Exit(1)
		}
		printTask(cmd, task)
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize CLI with an API key",
	Long:  `Store your API key in ~/.config/fromai/config.toml for persistent auth. Get your API key from the fromai settings page.`,
	Example: `  fai init --key "your-api-key-here"
  fai init --key "your-api-key" --base-url "https://fromai.example.com:8090"`,
	Run: func(cmd *cobra.Command, args []string) {
		key, _ := cmd.Flags().GetString("key")
		if key == "" {
			fmt.Fprintln(os.Stderr, "Error: --key is required. Get your API key from the settings page.")
			os.Exit(1)
		}

		cfg := &config.Config{
			APIKey:  key,
			BaseURL: flagBaseURL,
		}

		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}

		path, _ := config.Path()
		fmt.Printf("Config saved to %s\n", path)
		fmt.Println("Run 'fai whoami' to verify.")
	},
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show authenticated user info",
	Long:  `Verify your API key or token by fetching your user profile via /api/me.`,
	Example: `  fai whoami
  fai whoami --json`,
	Run: func(cmd *cobra.Command, args []string) {
		c := resolveClient()
		user, err := c.GetMe()
		if err != nil {
			printError(cmd, err)
			os.Exit(1)
		}
		if flagJSON {
			data, _ := json.MarshalIndent(user, "", "  ")
			fmt.Println(string(data))
			return
		}
		fmt.Printf("Authenticated as: %s (%s)\n", user.Email, user.ID)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "PocketBase auth token (or set FROMAI_TOKEN env var)")
	rootCmd.PersistentFlags().StringVar(&flagAPIKey, "api-key", "", "API key for X-API-Key header")
	rootCmd.PersistentFlags().StringVar(&flagBaseURL, "base-url", "https://fromai-backend.lyricut.com", "Base URL (default: $FROMAI_BASE_URL or https://fromai-backend.lyricut.com)")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output as raw JSON")

	createCmd.Flags().String("title", "", "Task title (required)")
	createCmd.Flags().String("description", "", "Task description")
	createCmd.MarkFlagRequired("title")
	createCmd.Flags().String("starter-code", "", "Starter code (required)")
	createCmd.MarkFlagRequired("starter-code")
	createCmd.Flags().String("language", "", "Programming language (required)")
	createCmd.MarkFlagRequired("language")

	updateCmd.Flags().String("code", "", "Code to set (required)")
	updateCmd.MarkFlagRequired("code")

	gradeCmd.Flags().String("grade", "", "Grade (required)")
	gradeCmd.MarkFlagRequired("grade")
	gradeCmd.Flags().String("feedback", "", "Feedback")

	pollCmd.Flags().Duration("interval", 5*time.Second, "Poll interval")
	pollCmd.Flags().Duration("timeout", 10*time.Minute, "Poll timeout")

	deleteCmd.Flags().Bool("hard", false, "Permanently delete instead of archiving")

	initCmd.Flags().String("key", "", "API key from settings page (required)")
	initCmd.MarkFlagRequired("key")

	taskCmd.AddCommand(createCmd, listCmd, getCmd, updateCmd, submitCmd, gradeCmd, deleteCmd, pollCmd)
	rootCmd.AddCommand(taskCmd, initCmd, whoamiCmd)
}

func main() {
	// Load .env files (local overrides base, like Vite)
	godotenv.Load(".env.local")
	godotenv.Load()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
