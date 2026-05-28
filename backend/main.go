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

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"
)

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
	return app.Save(collection)
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
				return nil
			}
			apiKey := e.Request.Header.Get("X-API-Key")
			if apiKey == "" {
				return nil // no API key provided, let RequireAuth handle it
			}
			user, err := findUserByAPIKey(app, apiKey)
			if err != nil {
				return nil // invalid key, let RequireAuth reject
			}
			e.Auth = user
			return nil
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
		// Ensure api_key field exists on users collection
		if err := ensureAPIKeyField(app); err != nil {
			log.Printf("Warning: failed to ensure api_key field: %v", err)
		}

		// Custom Go routes for tasks (protected with auth)
		se.Router.GET("/api/tasks", listTasksHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())
		se.Router.GET("/api/tasks/{id}", getTaskHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())
		se.Router.POST("/api/tasks", createTaskHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())
		se.Router.PATCH("/api/tasks/{id}", updateTaskHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())
		se.Router.POST("/api/tasks/{id}/submit", submitTaskHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())
		se.Router.POST("/api/tasks/{id}/grade", gradeTaskHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())
		se.Router.GET("/api/me", getMeHandler(app)).Bind(optionalAPIKeyAuth(app), apis.RequireAuth())

		// API key management (JWT auth only — user must be logged in to retrieve key)
		se.Router.GET("/api/me/api-key", apiKeyHandler(app)).Bind(apis.RequireAuth())
		se.Router.POST("/api/me/api-key", apiKeyHandler(app)).Bind(apis.RequireAuth())

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

		result := make([]map[string]any, len(records))
		for i, r := range records {
			result[i] = r.PublicExport()
		}

		return e.JSON(http.StatusOK, result)
	}
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

		record.Set("status", "completed")
		record.Set("updated_at", time.Now().UnixMilli())

		if err := app.Save(record); err != nil {
			return err
		}

		return e.JSON(http.StatusOK, record.PublicExport())
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
		record.Set("user", userId)
		record.Set("created_at", now)
		record.Set("updated_at", now)

		if err := app.Save(record); err != nil {
			return err
		}

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

		return e.JSON(http.StatusOK, record.PublicExport())
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

		result := make([]map[string]any, len(records))
		for i, r := range records {
			result[i] = r.PublicExport()
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

		help := map[string]any{
			"description": "Coding Gym — AI Agent Integration API",
			"base_url":    "http://localhost:8090",
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
						"user_id": "optional — filter by assigned user",
						"status":  "optional — pending or completed",
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
