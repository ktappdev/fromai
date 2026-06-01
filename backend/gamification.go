package main

import (
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

type badge struct {
	ID          string
	Name        string
	Description string
	Icon        string
}

var badgeCatalog = []badge{
	{ID: "first_task", Name: "First Steps", Description: "Complete your first task", Icon: "🎯"},
	{ID: "streak_3", Name: "On Fire", Description: "Maintain a 3-day streak", Icon: "🔥"},
	{ID: "streak_7", Name: "Week Warrior", Description: "Maintain a 7-day streak", Icon: "⚔️"},
	{ID: "streak_30", Name: "Monthly Master", Description: "Maintain a 30-day streak", Icon: "🏆"},
	{ID: "perfect_five", Name: "Perfect Five", Description: "Get 5 A grades in a row", Icon: "⭐"},
	{ID: "speed_runner", Name: "Speed Runner", Description: "Complete 5 tasks in under 10 minutes each", Icon: "⚡"},
	{ID: "polyglot", Name: "Polyglot", Description: "Complete tasks in 5 different languages", Icon: "🌍"},
	{ID: "algorithm_master", Name: "Algorithm Master", Description: "Complete 10 algorithm category tasks", Icon: "🧮"},
	{ID: "daily_champion", Name: "Daily Champion", Description: "Complete 7 daily challenges", Icon: "👑"},
	{ID: "century", Name: "Century", Description: "Complete 100 total tasks", Icon: "💯"},
}

func ensureCompletedAtField(app *pocketbase.PocketBase) error {
	collection, err := app.FindCollectionByNameOrId("tasks")
	if err != nil {
		return err
	}
	for _, field := range collection.Fields {
		if field.GetName() == "completed_at" {
			return nil
		}
	}
	collection.Fields.Add(&core.NumberField{
		Name:     "completed_at",
		Required: false,
		Hidden:   false,
	})
	return app.SaveNoValidate(collection)
}

func ensureUserStatsCollection(app *pocketbase.PocketBase) error {
	if _, err := app.FindCollectionByNameOrId("user_stats"); err == nil {
		return nil
	}

	collection := core.NewCollection("user_stats", "user_stats")
	collection.ListRule = types.Pointer("user = @request.auth.id")
	collection.ViewRule = types.Pointer("user = @request.auth.id")
	collection.CreateRule = types.Pointer("user = @request.auth.id")
	collection.UpdateRule = types.Pointer("user = @request.auth.id")

	collection.Fields.Add(&core.RelationField{
		Name:       "user",
		MaxSelect:  1,
		Presentable: false,
		Required:    true,
		CollectionId: "", // Will be set to users collection ID
	})
	collection.Fields.Add(&core.NumberField{
		Name:     "current_streak",
		Required: false,
	})
	collection.Fields.Add(&core.NumberField{
		Name:     "best_streak",
		Required: false,
	})
	collection.Fields.Add(&core.TextField{
		Name:     "last_activity_date",
		Required: false,
	})
	collection.Fields.Add(&core.NumberField{
		Name:     "total_tasks_completed",
		Required: false,
	})
	collection.Fields.Add(&core.JSONField{
		Name:     "badges",
		Required: false,
	})
	collection.Fields.Add(&core.JSONField{
		Name:     "languages_used",
		Required: false,
	})

	if err := app.SaveNoValidate(collection); err != nil {
		return err
	}

	// Set the relation to users collection after creation
	usersCollection, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}
	collection, _ = app.FindCollectionByNameOrId("user_stats")
	for _, field := range collection.Fields {
		if relField, ok := field.(*core.RelationField); ok && relField.GetName() == "user" {
			relField.CollectionId = usersCollection.Id
			break
		}
	}
	return app.SaveNoValidate(collection)
}

func ensureDailyChallengesCollection(app *pocketbase.PocketBase) error {
	if _, err := app.FindCollectionByNameOrId("daily_challenges"); err == nil {
		return nil
	}

	collection := core.NewCollection("daily_challenges", "daily_challenges")
	collection.ListRule = types.Pointer("") // Public readable
	collection.ViewRule = types.Pointer("") // Public readable

	collection.Fields.Add(&core.TextField{
		Name:     "title",
		Required: true,
	})
	collection.Fields.Add(&core.TextField{
		Name:       "description",
		Required:   true,
		Presentable: false,
	})
	collection.Fields.Add(&core.TextField{
		Name:       "starter_code",
		Required:   true,
		Presentable: false,
	})
	collection.Fields.Add(&core.SelectField{
		Name:     "language",
		Required: true,
		Values:   []string{"javascript", "typescript", "python", "go", "rust", "java", "cpp"},
	})
	collection.Fields.Add(&core.SelectField{
		Name:     "difficulty",
		Required: true,
		Values:   []string{"easy", "medium", "hard"},
	})
	collection.Fields.Add(&core.SelectField{
		Name:     "category",
		Required: true,
		Values:   []string{"algorithm", "bugfix", "utility", "data-structure", "system-design"},
	})
	collection.Fields.Add(&core.DateField{
		Name:     "date",
		Required: true,
	})
	collection.Fields.Add(&core.BoolField{
		Name:     "is_active",
		Required: false,
	})
	collection.Fields.Add(&core.JSONField{
		Name:     "tags",
		Required: false,
	})

	return app.SaveNoValidate(collection)
}

func ensureChallengeCompletionsCollection(app *pocketbase.PocketBase) error {
	if _, err := app.FindCollectionByNameOrId("challenge_completions"); err == nil {
		return nil
	}

	collection := core.NewCollection("challenge_completions", "challenge_completions")
	collection.ListRule = types.Pointer("user = @request.auth.id")
	collection.ViewRule = types.Pointer("user = @request.auth.id")
	collection.CreateRule = types.Pointer("user = @request.auth.id")

	collection.Fields.Add(&core.RelationField{
		Name:       "user",
		MaxSelect:  1,
		Presentable: false,
		Required:    true,
		CollectionId: "", // Will be set to users collection ID
	})
	collection.Fields.Add(&core.RelationField{
		Name:       "challenge",
		MaxSelect:  1,
		Presentable: false,
		Required:    true,
		CollectionId: "", // Will be set to daily_challenges collection ID
	})
	collection.Fields.Add(&core.RelationField{
		Name:       "task",
		MaxSelect:  1,
		Presentable: false,
		Required:    false,
		CollectionId: "", // Will be set to tasks collection ID
	})
	collection.Fields.Add(&core.DateField{
		Name:     "completed_at",
		Required: false,
	})
	collection.Fields.Add(&core.TextField{
		Name:     "grade",
		Required: false,
	})
	collection.Fields.Add(&core.TextField{
		Name:       "feedback",
		Required:   false,
		Presentable: false,
	})
	collection.Fields.Add(&core.NumberField{
		Name:     "time_seconds",
		Required: false,
	})

	if err := app.SaveNoValidate(collection); err != nil {
		return err
	}

	// Set relations after creation
	usersCollection, _ := app.FindCollectionByNameOrId("users")
	challengesCollection, _ := app.FindCollectionByNameOrId("daily_challenges")
	tasksCollection, _ := app.FindCollectionByNameOrId("tasks")

	collection, _ = app.FindCollectionByNameOrId("challenge_completions")
	for _, field := range collection.Fields {
		if relField, ok := field.(*core.RelationField); ok {
			switch relField.GetName() {
			case "user":
				relField.CollectionId = usersCollection.Id
			case "challenge":
				relField.CollectionId = challengesCollection.Id
			case "task":
				relField.CollectionId = tasksCollection.Id
			}
		}
	}
	return app.SaveNoValidate(collection)
}

func getOrCreateUserStats(app *pocketbase.PocketBase, userId string) (*core.Record, error) {
	records, err := app.FindRecordsByFilter(
		"user_stats",
		"user = {:userId}",
		"",
		1,
		0,
		map[string]any{"userId": userId},
	)
	if err != nil {
		return nil, err
	}
	if len(records) > 0 {
		return records[0], nil
	}

	collection, err := app.FindCollectionByNameOrId("user_stats")
	if err != nil {
		return nil, err
	}

	record := core.NewRecord(collection)
	record.Set("user", userId)
	record.Set("current_streak", 0)
	record.Set("best_streak", 0)
	record.Set("last_activity_date", "")
	record.Set("total_tasks_completed", 0)
	record.Set("badges", []string{})
	record.Set("languages_used", []string{})

	if err := app.Save(record); err != nil {
		return nil, err
	}
	return record, nil
}

func todayUTC() string {
	return time.Now().UTC().Format("2006-01-02")
}

func yesterdayUTC() string {
	return time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
}

func updateStreak(stats *core.Record) error {
	today := todayUTC()
	yesterday := yesterdayUTC()
	lastActivity := stats.GetString("last_activity_date")

	var newStreak int
	if lastActivity == "" {
		newStreak = 1
	} else if lastActivity == today {
		// Already completed today, streak unchanged
		return nil
	} else if lastActivity == yesterday {
		// Consecutive day, increment
		newStreak = stats.GetInt("current_streak") + 1
	} else {
		// Streak broken, reset to 1
		newStreak = 1
	}

	stats.Set("current_streak", newStreak)
	stats.Set("last_activity_date", today)

	bestStreak := stats.GetInt("best_streak")
	if newStreak > bestStreak {
		stats.Set("best_streak", newStreak)
	}

	return nil
}

func incrementTotalTasks(stats *core.Record) error {
	total := stats.GetInt("total_tasks_completed")
	stats.Set("total_tasks_completed", total+1)
	return nil
}

func addLanguage(stats *core.Record, language string) error {
	languages := getJSONStringSlice(stats, "languages_used")
	for _, lang := range languages {
		if lang == language {
			return nil // Already exists
		}
	}

	languages = append(languages, language)
	stats.Set("languages_used", languages)
	return nil
}

func getJSONStringSlice(record *core.Record, field string) []string {
	val := record.Get(field)
	if val == nil {
		return []string{}
	}

	switch v := val.(type) {
	case []string:
		return v
	case []any:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	default:
		return []string{}
	}
}

func evaluateBadges(app *pocketbase.PocketBase, userId string) ([]string, error) {
	stats, err := getOrCreateUserStats(app, userId)
	if err != nil {
		return nil, err
	}

	currentBadges := getJSONStringSlice(stats, "badges")
	newBadges := []string{}

	// Check each badge condition
	for _, badge := range badgeCatalog {
		alreadyHas := false
		for _, b := range currentBadges {
			if b == badge.ID {
				alreadyHas = true
				break
			}
		}
		if alreadyHas {
			continue
		}

		earned := false
		switch badge.ID {
		case "first_task":
			earned = stats.GetInt("total_tasks_completed") >= 1
		case "streak_3":
			earned = stats.GetInt("current_streak") >= 3
		case "streak_7":
			earned = stats.GetInt("current_streak") >= 7
		case "streak_30":
			earned = stats.GetInt("current_streak") >= 30
		case "perfect_five":
			earned = checkPerfectFive(app, userId)
		case "speed_runner":
			earned = checkSpeedRunner(app, userId)
		case "polyglot":
			earned = len(getJSONStringSlice(stats, "languages_used")) >= 5
		case "algorithm_master":
			count, _ := countCompletedByCategory(app, userId, "algorithm")
			earned = count >= 10
		case "daily_champion":
			count, _ := countDailyCompletions(app, userId)
			earned = count >= 7
		case "century":
			earned = stats.GetInt("total_tasks_completed") >= 100
		}

		if earned {
			newBadges = append(newBadges, badge.ID)
		}
	}

	if len(newBadges) > 0 {
		updatedBadges := append(currentBadges, newBadges...)
		stats.Set("badges", updatedBadges)
		if err := app.Save(stats); err != nil {
			return nil, err
		}
	}

	return newBadges, nil
}

func checkPerfectFive(app *pocketbase.PocketBase, userId string) bool {
	records, err := app.FindRecordsByFilter(
		"tasks",
		"user = {:userId} && status = {:status}",
		"-created_at",
		5,
		0,
		map[string]any{"userId": userId, "status": "completed"},
	)
	if err != nil || len(records) < 5 {
		return false
	}

	for _, record := range records {
		if record.GetString("grade") != "A" {
			return false
		}
	}
	return true
}

func checkSpeedRunner(app *pocketbase.PocketBase, userId string) bool {
	records, err := app.FindRecordsByFilter(
		"tasks",
		"user = {:userId} && status = {:status}",
		"-created_at",
		100,
		0,
		map[string]any{"userId": userId, "status": "completed"},
	)
	if err != nil {
		return false
	}

	count := 0
	for _, record := range records {
		completedAtVal := record.Get("completed_at")
		createdAtVal := record.Get("created_at")
		completedAt := int64(0)
		createdAt := int64(0)
		if v, ok := completedAtVal.(int64); ok {
			completedAt = v
		} else if v, ok := completedAtVal.(float64); ok {
			completedAt = int64(v)
		}
		if v, ok := createdAtVal.(int64); ok {
			createdAt = v
		} else if v, ok := createdAtVal.(float64); ok {
			createdAt = int64(v)
		}
		if completedAt > 0 && createdAt > 0 {
			timeSeconds := (completedAt - createdAt) / 1000
			if timeSeconds > 0 && timeSeconds <= 600 { // 10 minutes = 600 seconds
				count++
			}
		}
		if count >= 5 {
			return true
		}
	}
	return false
}

func countCompletedByCategory(app *pocketbase.PocketBase, userId, category string) (int, error) {
	completions, err := app.FindRecordsByFilter(
		"challenge_completions",
		"user = {:userId}",
		"",
		100,
		0,
		map[string]any{"userId": userId},
	)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, completion := range completions {
		challengeId := completion.GetString("challenge")
		if challengeId == "" {
			continue
		}
		challenge, err := app.FindRecordById("daily_challenges", challengeId)
		if err != nil {
			continue
		}
		if challenge.GetString("category") == category {
			count++
		}
	}
	return count, nil
}

func countDailyCompletions(app *pocketbase.PocketBase, userId string) (int, error) {
	records, err := app.FindRecordsByFilter(
		"challenge_completions",
		"user = {:userId}",
		"",
		100,
		0,
		map[string]any{"userId": userId},
	)
	if err != nil {
		return 0, err
	}
	return len(records), nil
}

func meStatsHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		user := getAuthUser(e)
		if user == nil {
			return apis.NewApiError(http.StatusUnauthorized, "Not authenticated", nil)
		}

		stats, err := getOrCreateUserStats(app, user.Id)
		if err != nil {
			return err
		}

		// Evaluate badges on each fetch
		newBadges, _ := evaluateBadges(app, user.Id)

		result := map[string]any{
			"id":                     stats.Id,
			"current_streak":         stats.GetInt("current_streak"),
			"best_streak":            stats.GetInt("best_streak"),
			"last_activity_date":     stats.GetString("last_activity_date"),
			"total_tasks_completed":  stats.GetInt("total_tasks_completed"),
			"badges":                 getJSONStringSlice(stats, "badges"),
			"languages_used":         getJSONStringSlice(stats, "languages_used"),
			"newly_earned_badges":    newBadges,
			"badge_catalog":          badgeCatalog,
		}

		return e.JSON(http.StatusOK, result)
	}
}