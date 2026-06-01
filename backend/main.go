package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"
	"github.com/pocketbase/pocketbase/tools/types"
)

func ensureTaskRules(app *pocketbase.PocketBase) error {
	collection, err := app.FindCollectionByNameOrId("tasks")
	if err != nil {
		return err
	}

	ownerRule := "user = @request.auth.id"
	changed := false

	if collection.ListRule == nil || *collection.ListRule != ownerRule {
		collection.ListRule = types.Pointer(ownerRule)
		changed = true
	}
	if collection.ViewRule == nil || *collection.ViewRule != ownerRule {
		collection.ViewRule = types.Pointer(ownerRule)
		changed = true
	}

	if changed {
		return app.SaveNoValidate(collection)
	}
	return nil
}

func ensureArchivedField(app *pocketbase.PocketBase) error {
	collection, err := app.FindCollectionByNameOrId("tasks")
	if err != nil {
		return err
	}
	for _, field := range collection.Fields {
		if field.GetName() == "archived" {
			return nil // already exists
		}
	}
	collection.Fields.Add(&core.BoolField{
		Name:     "archived",
		Required: false,
		Hidden:   false,
	})
	return app.SaveNoValidate(collection)
}

func ensureAPIKeyField(app *pocketbase.PocketBase) error {
	collection, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}
	for _, field := range collection.Fields {
		if field.GetName() == "api_key" {
			return nil // already exists
		}
	}
	collection.Fields.Add(&core.TextField{
		Name:     "api_key",
		Required: false,
		Max:      64,
		Hidden:   true, // don't expose in public API responses
	})
	return app.SaveNoValidate(collection)
}

func generateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func findUserByAPIKey(app *pocketbase.PocketBase, key string) (*core.Record, error) {
	records, err := app.FindRecordsByFilter(
		"users",
		"api_key = {:key}",
		"",
		1,
		0,
		map[string]any{"key": key},
	)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("invalid API key")
	}
	return records[0], nil
}

func optionalAPIKeyAuth(app *pocketbase.PocketBase) *hook.Handler[*core.RequestEvent] {
	return &hook.Handler[*core.RequestEvent]{
		Func: func(e *core.RequestEvent) error {
			// If already authenticated via JWT, skip
			if e.Auth != nil {
				return e.Next()
			}
			apiKey := e.Request.Header.Get("X-API-Key")
			if apiKey == "" {
				return e.Next() // no API key provided, let RequireAuth handle it
			}
			user, err := findUserByAPIKey(app, apiKey)
			if err != nil {
				return e.Next() // invalid key, let RequireAuth reject
			}
			e.Auth = user
			return e.Next()
		},
	}
}

func apiKeyHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)
		user := getAuthUser(e)
		if user == nil {
			return apis.NewApiError(http.StatusUnauthorized, "Not authenticated", nil)
		}
		switch e.Request.Method {
		case http.MethodGet:
			// If no API key yet, generate one
			key := user.GetString("api_key")
			if key == "" {
				var err error
				key, err = generateAPIKey()
				if err != nil {
					return apis.NewApiError(http.StatusInternalServerError, "Failed to generate API key", err)
				}
				user.Set("api_key", key)
				if err := app.Save(user); err != nil {
					return apis.NewApiError(http.StatusInternalServerError, "Failed to save API key", err)
				}
			}
			return e.JSON(http.StatusOK, map[string]string{"api_key": key})
		case http.MethodPost:
			// Regenerate
			key, err := generateAPIKey()
			if err != nil {
				return apis.NewApiError(http.StatusInternalServerError, "Failed to generate API key", err)
			}
			user.Set("api_key", key)
			if err := app.Save(user); err != nil {
				return apis.NewApiError(http.StatusInternalServerError, "Failed to save API key", err)
			}
			return e.JSON(http.StatusOK, map[string]string{"api_key": key})
		default:
			return apis.NewApiError(http.StatusMethodNotAllowed, "Method not allowed", nil)
		}
	}
}

func main() {
	// Load .env files (local overrides base, like Vite)
	godotenv.Load(".env.local")
	godotenv.Load()

	app := pocketbase.New()

	app.OnRecordCreate("users").BindFunc(func(e *core.RecordEvent) error {
		key, err := generateAPIKey()
		if err != nil {
			return err
		}
		e.Record.Set("api_key", key)
		return e.Next()
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Ensure tasks collection rules allow owner-scoped realtime subscriptions
		if err := ensureTaskRules(app); err != nil {
			log.Printf("Warning: failed to ensure tasks rules: %v", err)
		}

		// Ensure archived field exists on tasks collection
		if err := ensureArchivedField(app); err != nil {
			log.Printf("Warning: failed to ensure archived field: %v", err)
		}

		// Ensure api_key field exists on users collection
		if err := ensureAPIKeyField(app); err != nil {
			log.Printf("Warning: failed to ensure api_key field: %v", err)
		}

		// Ensure gamification collections and fields
		if err := ensureCompletedAtField(app); err != nil {
			log.Printf("Warning: failed to ensure completed_at field: %v", err)
		}
		if err := ensureUserStatsCollection(app); err != nil {
			log.Printf("Warning: failed to ensure user_stats collection: %v", err)
		}
		if err := ensureDailyChallengesCollection(app); err != nil {
			log.Printf("Warning: failed to ensure daily_challenges collection: %v", err)
		}
		if err := ensureChallengeCompletionsCollection(app); err != nil {
			log.Printf("Warning: failed to ensure challenge_completions collection: %v", err)
		}
		if err := seedDailyChallenges(app); err != nil {
			log.Printf("Warning: failed to seed daily challenges: %v", err)
		}

		// Ensure Telegram integration fields and collections
		if err := ensureTelegramChatIDField(app); err != nil {
			log.Printf("Warning: failed to ensure telegram_chat_id field: %v", err)
		}
		if err := ensureTelegramVerificationsCollection(app); err != nil {
			log.Printf("Warning: failed to ensure telegram_verifications collection: %v", err)
		}

		// Start Telegram polling if token is configured
		telegramBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")
		if telegramBotToken != "" {
			startTelegramPolling(app, telegramBotToken)
		} else {
			log.Printf("TELEGRAM_BOT_TOKEN not set — Telegram notifications disabled")
		}

		// Custom Go routes for tasks (protected with auth)
		se.Router.GET("/api/tasks", listTasksHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())
		se.Router.GET("/api/tasks/{id}", getTaskHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())
		se.Router.POST("/api/tasks", createTaskHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())
		se.Router.PATCH("/api/tasks/{id}", updateTaskHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())
		se.Router.POST("/api/tasks/{id}/submit", submitTaskHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())
		se.Router.POST("/api/tasks/{id}/grade", gradeTaskHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())
		se.Router.POST("/api/tasks/{id}/archive", archiveTaskHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())
		se.Router.GET("/api/me", getMeHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())
		se.Router.DELETE("/api/tasks/{id}", deleteTaskHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())
		se.Router.POST("/api/tasks/{id}/delete", deleteTaskHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())

		// API key management (JWT auth only — user must be logged in to retrieve key)
		se.Router.GET("/api/me/api-key", apiKeyHandler(app)).Bind(apis.RequireAuth())
		se.Router.POST("/api/me/api-key", apiKeyHandler(app)).Bind(apis.RequireAuth())

		// Gamification routes
		se.Router.GET("/api/me/stats", meStatsHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())
		se.Router.GET("/api/challenges/today", todayChallengeHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())
		se.Router.GET("/api/challenges", listChallengesHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())
		se.Router.POST("/api/challenges/{id}/start", startChallengeHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())

		// Telegram verification routes
		se.Router.GET("/api/me/telegram/status", telegramStatusHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())
		se.Router.POST("/api/me/telegram/verify", telegramVerifyHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())
		se.Router.POST("/api/telegram/unsubscribe", telegramUnsubscribeHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())

		// External LLM-facing routes (API-key auth)
		se.Router.POST("/api/external/tasks", externalCreateTaskHandler(app))
		se.Router.GET("/api/external/tasks", externalListTasksHandler(app))
		se.Router.GET("/api/external/tasks/{id}", externalGetTaskHandler(app))
		se.Router.POST("/api/external/tasks/{id}/grade", externalGradeTaskHandler(app))
		se.Router.GET("/api/external/users", externalListUsersHandler(app))
		se.Router.GET("/api/external/help", externalHelpHandler())

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

func getAuthUser(e *core.RequestEvent) *core.Record {
	return e.Auth
}

func listTasksHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		user := getAuthUser(e)
		if user == nil {
			return apis.NewApiError(http.StatusUnauthorized, "Not authenticated", nil)
		}

		records, err := app.FindRecordsByFilter(
			"tasks",
			"user = {:userId}",
			"-created_at",
			100,
			0,
			map[string]any{"userId": user.Id},
		)
		if err != nil {
			return err
		}

		result := make([]map[string]any, 0, len(records))
		for _, r := range records {
			if r.GetBool("archived") {
				continue
			}
			result = append(result, r.PublicExport())
		}

		return e.JSON(http.StatusOK, result)
	}
}

func isArchived(record *core.Record) bool {
	return record.GetBool("archived")
}

func getTaskHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		user := getAuthUser(e)
		if user == nil {
			return apis.NewApiError(http.StatusUnauthorized, "Not authenticated", nil)
		}

		taskId := e.Request.PathValue("id")
		record, err := app.FindRecordById("tasks", taskId)
		if err != nil {
			return apis.NewNotFoundError("", err)
		}

		if record.GetString("user") != user.Id {
			return apis.NewForbiddenError("", nil)
		}

		if isArchived(record) {
			return apis.NewNotFoundError("", nil)
		}

		return e.JSON(http.StatusOK, record.PublicExport())
	}
}

func createTaskHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		user := getAuthUser(e)
		if user == nil {
			return apis.NewApiError(http.StatusUnauthorized, "Not authenticated", nil)
		}

		var data map[string]any
		if err := json.NewDecoder(e.Request.Body).Decode(&data); err != nil {
			return apis.NewBadRequestError("Invalid request body", err)
		}

		collection, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		now := time.Now().UnixMilli()
		record := core.NewRecord(collection)
		record.Set("title", data["title"])
		record.Set("description", data["description"])
		record.Set("starter_code", data["starter_code"])
		record.Set("code", data["starter_code"])
		record.Set("language", data["language"])
		record.Set("status", "pending")
		record.Set("archived", false)
		record.Set("user", user.Id)
		record.Set("created_at", now)
		record.Set("updated_at", now)

		if err := app.Save(record); err != nil {
			return err
		}

		return e.JSON(http.StatusCreated, record.PublicExport())
	}
}

func updateTaskHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		user := getAuthUser(e)
		if user == nil {
			return apis.NewApiError(http.StatusUnauthorized, "Not authenticated", nil)
		}

		taskId := e.Request.PathValue("id")
		record, err := app.FindRecordById("tasks", taskId)
		if err != nil {
			return apis.NewNotFoundError("", err)
		}

		if record.GetString("user") != user.Id {
			return apis.NewForbiddenError("", nil)
		}

		if isArchived(record) {
			return apis.NewNotFoundError("", nil)
		}

		var data map[string]any
		if err := json.NewDecoder(e.Request.Body).Decode(&data); err != nil {
			return apis.NewBadRequestError("Invalid request body", err)
		}

		if code, ok := data["code"]; ok {
			record.Set("code", code)
		}
		record.Set("updated_at", time.Now().UnixMilli())

		if err := app.Save(record); err != nil {
			return err
		}

		return e.JSON(http.StatusOK, record.PublicExport())
	}
}

func submitTaskHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		user := getAuthUser(e)
		if user == nil {
			return apis.NewApiError(http.StatusUnauthorized, "Not authenticated", nil)
		}

		taskId := e.Request.PathValue("id")
		record, err := app.FindRecordById("tasks", taskId)
		if err != nil {
			return apis.NewNotFoundError("", err)
		}

		if record.GetString("user") != user.Id {
			return apis.NewForbiddenError("", nil)
		}

		if isArchived(record) {
			return apis.NewNotFoundError("", nil)
		}

		// Idempotency check: skip gamification if already completed
		wasAlreadyCompleted := record.GetString("status") == "completed"

		record.Set("status", "completed")
		record.Set("updated_at", time.Now().UnixMilli())

		// Set completed_at only on first completion
		if !wasAlreadyCompleted {
			record.Set("completed_at", time.Now().UnixMilli())
		}

		if err := app.Save(record); err != nil {
			return err
		}

		// Build base response with task data
		response := record.PublicExport()

		// Only run gamification on first completion
		if !wasAlreadyCompleted {
			// Get or create user stats
			stats, err := getOrCreateUserStats(app, user.Id)
			if err != nil {
				log.Printf("Warning: failed to get/create user stats: %v", err)
			} else {
				// Increment total tasks
				if err := incrementTotalTasks(stats); err != nil {
					log.Printf("Warning: failed to increment total tasks: %v", err)
				}
				// Add language
				if err := addLanguage(stats, record.GetString("language")); err != nil {
					log.Printf("Warning: failed to add language: %v", err)
				}
				// Update streak
				if err := updateStreak(stats); err != nil {
					log.Printf("Warning: failed to update streak: %v", err)
				}
				// Save stats once after all modifications
				if err := app.Save(stats); err != nil {
					log.Printf("Warning: failed to save user stats: %v", err)
				}
				// Evaluate badges
				newlyEarned, err := evaluateBadges(app, user.Id)
				if err != nil {
					log.Printf("Warning: failed to evaluate badges: %v", err)
				} else if len(newlyEarned) > 0 {
					response["newly_earned_badges"] = newlyEarned
					// Notify user of newly earned badges
					for _, badgeID := range newlyEarned {
						go notifyUser(app, user.Id, fmt.Sprintf("🏆 Badge earned: %s!", badgeID))
					}
				}
				// Add streak info to response (post-modification values)
				response["current_streak"] = stats.GetInt("current_streak")
				response["best_streak"] = stats.GetInt("best_streak")
			}
		}

		return e.JSON(http.StatusOK, response)
	}
}

func gradeTaskHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		user := getAuthUser(e)
		if user == nil {
			return apis.NewApiError(http.StatusUnauthorized, "Not authenticated", nil)
		}

		taskId := e.Request.PathValue("id")
		record, err := app.FindRecordById("tasks", taskId)
		if err != nil {
			return apis.NewNotFoundError("", err)
		}

		if record.GetString("user") != user.Id {
			return apis.NewForbiddenError("", nil)
		}

		if isArchived(record) {
			return apis.NewNotFoundError("", nil)
		}

		var data map[string]any
		if err := json.NewDecoder(e.Request.Body).Decode(&data); err != nil {
			return apis.NewBadRequestError("Invalid request body", err)
		}

		if grade, ok := data["grade"]; ok {
			record.Set("grade", grade)
		}
		if feedback, ok := data["feedback"]; ok {
			record.Set("feedback", feedback)
		}
		record.Set("updated_at", time.Now().UnixMilli())

		if err := app.Save(record); err != nil {
			return err
		}

		// Notify user via Telegram
		go func() {
			title := record.GetString("title")
			grade := record.GetString("grade")
			feedback := record.GetString("feedback")
			message := fmt.Sprintf("✅ Task graded: %s\nGrade: %s\n%s", title, grade, feedback)
			notifyUser(app, user.Id, message)
		}()

		// Build base response with task data
		response := record.PublicExport()

		// Check if this task came from a daily challenge
		completions, err := app.FindRecordsByFilter(
			"challenge_completions",
			"task = {:taskId}",
			"",
			1,
			0,
			map[string]any{"taskId": record.Id},
		)
		if err == nil && len(completions) > 0 {
			// Update the challenge completion with grade and feedback
			completion := completions[0]
			if grade, ok := data["grade"]; ok {
				completion.Set("grade", grade)
			}
			if feedback, ok := data["feedback"]; ok {
				completion.Set("feedback", feedback)
			}
			// Mark challenge as completed
			completion.Set("completed_at", time.Now().UTC().Format("2006-01-02"))
			if err := app.Save(completion); err != nil {
				log.Printf("Warning: failed to update challenge completion: %v", err)
			}
		}

		// Re-evaluate badges in case grade-dependent badges are earned
		newlyEarned, err := evaluateBadges(app, user.Id)
		if err != nil {
			log.Printf("Warning: failed to evaluate badges: %v", err)
		} else if len(newlyEarned) > 0 {
			response["newly_earned_badges"] = newlyEarned
			// Notify user of newly earned badges
			for _, badgeID := range newlyEarned {
				go notifyUser(app, user.Id, fmt.Sprintf("🏆 Badge earned: %s!", badgeID))
			}
		}

		return e.JSON(http.StatusOK, response)
	}
}

func deleteTaskHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		user := getAuthUser(e)
		if user == nil {
			return apis.NewApiError(http.StatusUnauthorized, "Not authenticated", nil)
		}

		taskId := e.Request.PathValue("id")
		record, err := app.FindRecordById("tasks", taskId)
		if err != nil {
			return apis.NewNotFoundError("", err)
		}

		if record.GetString("user") != user.Id {
			return apis.NewNotFoundError("", nil)
		}

		if err := app.Delete(record); err != nil {
			return apis.NewApiError(http.StatusInternalServerError, "Failed to delete task", err)
		}

		return e.JSON(http.StatusOK, map[string]bool{"success": true})
	}
}

func archiveTaskHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		user := getAuthUser(e)
		if user == nil {
			return apis.NewApiError(http.StatusUnauthorized, "Not authenticated", nil)
		}

		taskId := e.Request.PathValue("id")
		record, err := app.FindRecordById("tasks", taskId)
		if err != nil {
			return apis.NewNotFoundError("", err)
		}

		if record.GetString("user") != user.Id {
			return apis.NewNotFoundError("", nil)
		}

		record.Set("archived", true)
		record.Set("updated_at", time.Now().UnixMilli())

		if err := app.Save(record); err != nil {
			return err
		}

		return e.JSON(http.StatusOK, record.PublicExport())
	}
}

func getMeHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		user := getAuthUser(e)
		if user == nil {
			return e.JSON(http.StatusUnauthorized, nil)
		}

		return e.JSON(http.StatusOK, user.PublicExport())
	}
}

func checkExternalAPIKey(e *core.RequestEvent) bool {
	// Only accept static EXTERNAL_API_KEY (system-level, set by operator)
	// Per-user API keys are NOT valid for external routes (prevents privilege escalation)
	staticKey := os.Getenv("EXTERNAL_API_KEY")
	apiKey := e.Request.Header.Get("X-API-Key")
	return staticKey != "" && apiKey == staticKey
}

func externalCreateTaskHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		if !checkExternalAPIKey(e) {
			return apis.NewApiError(http.StatusUnauthorized, "Invalid or missing API key", nil)
		}

		var data map[string]any
		if err := json.NewDecoder(e.Request.Body).Decode(&data); err != nil {
			return apis.NewBadRequestError("Invalid request body", err)
		}

		userId, ok := data["user_id"].(string)
		if !ok || userId == "" {
			return apis.NewBadRequestError("user_id is required", nil)
		}

		_, err := app.FindRecordById("users", userId)
		if err != nil {
			return apis.NewBadRequestError("Invalid user_id", err)
		}

		collection, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		now := time.Now().UnixMilli()
		record := core.NewRecord(collection)
		record.Set("title", data["title"])
		record.Set("description", data["description"])
		record.Set("starter_code", data["starter_code"])
		record.Set("code", data["starter_code"])
		record.Set("language", data["language"])
		record.Set("status", "pending")
		record.Set("archived", false)
		record.Set("user", userId)
		record.Set("created_at", now)
		record.Set("updated_at", now)

		if err := app.Save(record); err != nil {
			return err
		}

		// Notify user via Telegram
		go func() {
			title := record.GetString("title")
			language := record.GetString("language")
			message := fmt.Sprintf("📋 New task: %s (%s)\nOpen your dashboard to start working.", title, language)
			notifyUser(app, userId, message)
		}()

		return e.JSON(http.StatusCreated, record.PublicExport())
	}
}

func externalGetTaskHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		if !checkExternalAPIKey(e) {
			return apis.NewApiError(http.StatusUnauthorized, "Invalid or missing API key", nil)
		}

		taskId := e.Request.PathValue("id")
		record, err := app.FindRecordById("tasks", taskId)
		if err != nil {
			return apis.NewNotFoundError("", err)
		}

		return e.JSON(http.StatusOK, record.PublicExport())
	}
}

func externalGradeTaskHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		if !checkExternalAPIKey(e) {
			return apis.NewApiError(http.StatusUnauthorized, "Invalid or missing API key", nil)
		}

		taskId := e.Request.PathValue("id")
		record, err := app.FindRecordById("tasks", taskId)
		if err != nil {
			return apis.NewNotFoundError("", err)
		}

		var data map[string]any
		if err := json.NewDecoder(e.Request.Body).Decode(&data); err != nil {
			return apis.NewBadRequestError("Invalid request body", err)
		}

		if grade, ok := data["grade"]; ok {
			record.Set("grade", grade)
		}
		if feedback, ok := data["feedback"]; ok {
			record.Set("feedback", feedback)
		}
		record.Set("updated_at", time.Now().UnixMilli())

		if err := app.Save(record); err != nil {
			return err
		}

		// Build base response with task data
		response := record.PublicExport()

		// Get the task's user ID
		userId := record.GetString("user")
		if userId != "" {
			// Notify user via Telegram
			go func() {
				title := record.GetString("title")
				grade := record.GetString("grade")
				feedback := record.GetString("feedback")
				message := fmt.Sprintf("✅ Task graded: %s\nGrade: %s\n%s", title, grade, feedback)
				notifyUser(app, userId, message)
			}()

			// Check if this task came from a daily challenge
			completions, err := app.FindRecordsByFilter(
				"challenge_completions",
				"task = {:taskId}",
				"",
				1,
				0,
				map[string]any{"taskId": record.Id},
			)
			if err == nil && len(completions) > 0 {
				// Update the challenge completion with grade and feedback
				completion := completions[0]
				if grade, ok := data["grade"]; ok {
					completion.Set("grade", grade)
				}
				if feedback, ok := data["feedback"]; ok {
					completion.Set("feedback", feedback)
				}
				// Mark challenge as completed
				completion.Set("completed_at", time.Now().UTC().Format("2006-01-02"))
				if err := app.Save(completion); err != nil {
					log.Printf("Warning: failed to update challenge completion: %v", err)
				}
			}

			// Re-evaluate badges in case grade-dependent badges are earned
			newlyEarned, err := evaluateBadges(app, userId)
			if err != nil {
				log.Printf("Warning: failed to evaluate badges: %v", err)
			} else if len(newlyEarned) > 0 {
				response["newly_earned_badges"] = newlyEarned
				// Notify user of newly earned badges
				for _, badgeID := range newlyEarned {
					go notifyUser(app, userId, fmt.Sprintf("🏆 Badge earned: %s!", badgeID))
				}
			}
		}

		return e.JSON(http.StatusOK, response)
	}
}

func externalListTasksHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		if !checkExternalAPIKey(e) {
			return apis.NewApiError(http.StatusUnauthorized, "Invalid or missing API key", nil)
		}

		filter := "1=1"
		params := map[string]any{}

		if userId := e.Request.URL.Query().Get("user_id"); userId != "" {
			filter = "user = {:userId}"
			params["userId"] = userId
		}
		if status := e.Request.URL.Query().Get("status"); status != "" {
			filter = filter + " && status = {:status}"
			params["status"] = status
		}

		records, err := app.FindRecordsByFilter(
			"tasks",
			filter,
			"-created_at",
			100,
			0,
			params,
		)
		if err != nil {
			return err
		}

		includeArchived := e.Request.URL.Query().Get("include_archived") == "true"
		result := make([]map[string]any, 0, len(records))
		for _, r := range records {
			if !includeArchived && r.GetBool("archived") {
				continue
			}
			result = append(result, r.PublicExport())
		}

		return e.JSON(http.StatusOK, result)
	}
}

func externalListUsersHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		if !checkExternalAPIKey(e) {
			return apis.NewApiError(http.StatusUnauthorized, "Invalid or missing API key", nil)
		}

		records, err := app.FindRecordsByFilter(
			"users",
			"1=1",
			"created",
			100,
			0,
			map[string]any{},
		)
		if err != nil {
			return err
		}

		result := make([]map[string]any, len(records))
		for i, r := range records {
			result[i] = r.PublicExport()
		}

		return e.JSON(http.StatusOK, result)
	}
}

func externalHelpHandler() func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		baseURL := os.Getenv("PUBLIC_URL")
		if baseURL == "" {
			baseURL = "http://localhost:8090"
		}

		help := map[string]any{
			"description": "fromai — AI Agent Integration API",
			"base_url":    baseURL,
			"auth": map[string]string{
				"header":     "X-API-Key",
				"env_source": "EXTERNAL_API_KEY (set by human operator)",
			},
			"endpoints": []map[string]any{
				{
					"method":      "GET",
					"path":        "/api/external/users",
					"description": "List all registered users to discover targets",
					"auth":        "X-API-Key",
				},
				{
					"method":      "POST",
					"path":        "/api/external/tasks",
					"description": "Create and assign a coding task to a user",
					"auth":        "X-API-Key",
					"body": map[string]string{
						"user_id":      "required — target user's PocketBase ID",
						"title":        "required — short task name",
						"starter_code": "required — initial code shown to user",
						"language":     "required — typescript|javascript|python|go|rust|java|cpp",
						"description":  "optional — full instructions",
					},
					"example": map[string]string{
						"user_id":      "drxvuwj1f95lqa1",
						"title":        "Reverse a Linked List",
						"starter_code": "function reverseList(head) {\n  // your code here\n  return head;\n}",
						"language":     "javascript",
						"description":  "Implement a function that reverses a singly linked list in-place.",
					},
				},
				{
					"method":      "GET",
					"path":        "/api/external/tasks",
					"description": "List tasks with optional filtering",
					"auth":        "X-API-Key",
					"query_params": map[string]string{
						"user_id":          "optional — filter by assigned user",
						"status":           "optional — pending or completed",
						"include_archived": "optional — include archived tasks (default false)",
					},
					"example": "GET /api/external/tasks?user_id=drxvuwj1f95lqa1&status=completed",
				},
				{
					"method":      "GET",
					"path":        "/api/external/tasks/{id}",
					"description": "Get a single task including submitted code",
					"auth":        "X-API-Key",
				},
				{
					"method":      "POST",
					"path":        "/api/external/tasks/{id}/grade",
					"description": "Grade a completed task and provide feedback",
					"auth":        "X-API-Key",
					"body": map[string]string{
						"grade":    "required — e.g. A, B+, Pass",
						"feedback": "required — detailed textual feedback",
					},
					"example": map[string]string{
						"grade":    "A",
						"feedback": "Clean iterative solution. O(n) time, O(1) space.",
					},
				},
			},
			"workflow": []string{
				"1. GET /api/external/users — discover available users",
				"2. POST /api/external/tasks — create and assign a task (save returned id)",
				"3. GET /api/external/tasks/{id} or GET /api/external/tasks?status=completed — poll until status == completed",
				"4. GET /api/external/tasks/{id} — retrieve submitted code from the 'code' field",
				"5. POST /api/external/tasks/{id}/grade — send grade and feedback",
			},
			"tips": []string{
				"Poll every 30 seconds. The human works in a browser editor and clicks 'Finish' when done.",
				"Use status=pending to see tasks that are still open.",
				"The 'code' field contains the human's final submission.",
				"Grades appear instantly in the human's UI via real-time subscriptions.",
			},
		}

		return e.JSON(http.StatusOK, help)
	}
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
}
