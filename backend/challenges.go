package main

import (
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func seedDailyChallenges(app *pocketbase.PocketBase) error {
	today := time.Now().UTC()

	for i := 0; i < 30; i++ {
		date := today.AddDate(0, 0, i).Format("2006-01-02")
		seedIndex := i % len(challengeSeeds)
		seed := challengeSeeds[seedIndex]

		// Check if challenge already exists for this date
		existing, err := app.FindRecordsByFilter(
			"daily_challenges",
			"date = {:date}",
			"",
			1,
			0,
			map[string]any{"date": date},
		)
		if err == nil && len(existing) > 0 {
			continue // Skip if already exists
		}

		collection, err := app.FindCollectionByNameOrId("daily_challenges")
		if err != nil {
			return err
		}

		record := core.NewRecord(collection)
		record.Set("title", seed.Title)
		record.Set("description", seed.Description)
		record.Set("starter_code", seed.StarterCode)
		record.Set("language", seed.Language)
		record.Set("difficulty", seed.Difficulty)
		record.Set("category", seed.Category)
		record.Set("date", date)
		record.Set("is_active", i == 0) // Only today is active
		record.Set("tags", seed.Tags)

		if err := app.Save(record); err != nil {
			return err
		}
	}

	return nil
}

func todayChallengeHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		user := getAuthUser(e)
		if user == nil {
			return apis.NewApiError(http.StatusUnauthorized, "Not authenticated", nil)
		}

		today := time.Now().UTC().Format("2006-01-02")

		records, err := app.FindRecordsByFilter(
			"daily_challenges",
			"date = {:date}",
			"",
			1,
			0,
			map[string]any{"date": today},
		)
		if err != nil {
			return err
		}
		if len(records) == 0 {
			return apis.NewNotFoundError("No challenge found for today", nil)
		}

		challenge := records[0]

		// Check if user has already completed this challenge
		completions, err := app.FindRecordsByFilter(
			"challenge_completions",
			"user = {:userId} && challenge = {:challengeId}",
			"",
			1,
			0,
			map[string]any{"userId": user.Id, "challengeId": challenge.Id},
		)
		if err != nil {
			return err
		}

		result := challenge.PublicExport()
		completed := len(completions) > 0 && completions[0].GetString("completed_at") != ""
		result["completed"] = completed
		if len(completions) > 0 {
			result["completion"] = completions[0].PublicExport()
		}

		return e.JSON(http.StatusOK, result)
	}
}

func listChallengesHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		user := getAuthUser(e)
		if user == nil {
			return apis.NewApiError(http.StatusUnauthorized, "Not authenticated", nil)
		}

		records, err := app.FindRecordsByFilter(
			"daily_challenges",
			"date <= {:today}",
			"-date",
			30,
			0,
			map[string]any{"today": time.Now().UTC().Format("2006-01-02")},
		)
		if err != nil {
			return err
		}

		// Get all user completions for batch lookup
		userCompletions, err := app.FindRecordsByFilter(
			"challenge_completions",
			"user = {:userId}",
			"",
			100,
			0,
			map[string]any{"userId": user.Id},
		)
		if err != nil {
			return err
		}

		// Build a map of challenge_id -> completion for quick lookup
		completionMap := make(map[string]*core.Record)
		for _, comp := range userCompletions {
			challengeId := comp.GetString("challenge")
			completionMap[challengeId] = comp
		}

		result := make([]map[string]any, 0, len(records))
		for _, challenge := range records {
			export := challenge.PublicExport()
			if completion, ok := completionMap[challenge.Id]; ok {
				completed := completion.GetString("completed_at") != ""
				export["completed"] = completed
				export["completion"] = completion.PublicExport()
			} else {
				export["completed"] = false
			}
			result = append(result, export)
		}

		return e.JSON(http.StatusOK, result)
	}
}

func startChallengeHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		user := getAuthUser(e)
		if user == nil {
			return apis.NewApiError(http.StatusUnauthorized, "Not authenticated", nil)
		}

		challengeId := e.Request.PathValue("id")
		challenge, err := app.FindRecordById("daily_challenges", challengeId)
		if err != nil {
			return apis.NewNotFoundError("Challenge not found", err)
		}

		// Check if user already has a completion record for this challenge
		existingCompletions, err := app.FindRecordsByFilter(
			"challenge_completions",
			"user = {:userId} && challenge = {:challengeId}",
			"",
			1,
			0,
			map[string]any{"userId": user.Id, "challengeId": challengeId},
		)
		if err == nil && len(existingCompletions) > 0 {
			// Return the existing task
			taskId := existingCompletions[0].GetString("task")
			if taskId != "" {
				task, err := app.FindRecordById("tasks", taskId)
				if err == nil {
					return e.JSON(http.StatusOK, map[string]any{
						"task":       task.PublicExport(),
						"challenge":  challenge.PublicExport(),
						"completion": existingCompletions[0].PublicExport(),
					})
				}
			}
			return apis.NewApiError(http.StatusConflict, "Challenge already started", nil)
		}

		// Create a new task from the challenge
		tasksCollection, err := app.FindCollectionByNameOrId("tasks")
		if err != nil {
			return err
		}

		now := time.Now().UnixMilli()
		task := core.NewRecord(tasksCollection)
		task.Set("title", challenge.GetString("title"))
		task.Set("description", challenge.GetString("description"))
		task.Set("starter_code", challenge.GetString("starter_code"))
		task.Set("code", challenge.GetString("starter_code"))
		task.Set("language", challenge.GetString("language"))
		task.Set("status", "pending")
		task.Set("archived", false)
		task.Set("user", user.Id)
		task.Set("created_at", now)
		task.Set("updated_at", now)

		if err := app.Save(task); err != nil {
			return err
		}

		// Create a challenge completion record
		completionsCollection, err := app.FindCollectionByNameOrId("challenge_completions")
		if err != nil {
			return err
		}

		completion := core.NewRecord(completionsCollection)
		completion.Set("user", user.Id)
		completion.Set("challenge", challengeId)
		completion.Set("task", task.Id)
		completion.Set("grade", "")
		completion.Set("feedback", "")
		completion.Set("time_seconds", 0)

		if err := app.Save(completion); err != nil {
			return err
		}

		return e.JSON(http.StatusCreated, map[string]any{
			"task":       task.PublicExport(),
			"challenge":  challenge.PublicExport(),
			"completion": completion.PublicExport(),
		})
	}
}