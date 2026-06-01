package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

func ensureTelegramChatIDField(app *pocketbase.PocketBase) error {
	collection, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}
	for _, field := range collection.Fields {
		if field.GetName() == "telegram_chat_id" {
			return nil // already exists
		}
	}
	collection.Fields.Add(&core.TextField{
		Name:     "telegram_chat_id",
		Required: false,
		Hidden:   true, // don't expose in public API responses
	})
	return app.SaveNoValidate(collection)
}

func ensureTelegramVerificationsCollection(app *pocketbase.PocketBase) error {
	if _, err := app.FindCollectionByNameOrId("telegram_verifications"); err == nil {
		return nil
	}

	collection := core.NewCollection("telegram_verifications", "telegram_verifications")
	collection.ListRule = types.Pointer("id = ''") // Internal-only: never matches, no public list access
	collection.ViewRule = types.Pointer("id = ''") // Internal-only: never matches, no public view access

	collection.Fields.Add(&core.TextField{
		Name:     "code",
		Required: true,
		Max:      6,
	})
	collection.Fields.Add(&core.TextField{
		Name:     "chat_id",
		Required: true,
	})
	collection.Fields.Add(&core.NumberField{
		Name:     "expires_at",
		Required: true,
	})
	collection.Fields.Add(&core.NumberField{
		Name:     "used_at",
		Required: false,
	})

	return app.SaveNoValidate(collection)
}

func generateVerificationCode() string {
	b := make([]byte, 3)
	_, err := rand.Read(b)
	if err != nil {
		return "000000" // fallback
	}
	// Convert to 6 digits (000000-999999)
	n := (int(b[0])<<16 | int(b[1])<<8 | int(b[2])) % 1000000
	return fmt.Sprintf("%06d", n)
}

type telegramUpdate struct {
	UpdateID int64 `json:"update_id"`
	Message  *struct {
		Chat *struct {
			ID   int64  `json:"id"`
			Type string `json:"type"`
		} `json:"chat"`
		Text string `json:"text"`
	} `json:"message"`
}

type telegramResponse struct {
	OK     bool             `json:"ok"`
	Result []telegramUpdate `json:"result"`
}

var globalTelegramBot *telegramBot

type telegramBot struct {
	token  string
	client *http.Client
}

func newTelegramBot(token string) *telegramBot {
	return &telegramBot{
		token:  token,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (b *telegramBot) getUpdates(offset int64, timeout int) ([]telegramUpdate, int64, error) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=%d", b.token, offset, timeout)

	resp, err := b.client.Get(apiURL)
	if err != nil {
		return nil, offset, err
	}
	defer resp.Body.Close()

	var result telegramResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, offset, err
	}

	if !result.OK {
		return nil, offset, fmt.Errorf("telegram API error")
	}

	nextOffset := offset
	for _, update := range result.Result {
		if update.UpdateID >= nextOffset {
			nextOffset = update.UpdateID + 1
		}
	}

	return result.Result, nextOffset, nil
}

func (b *telegramBot) sendMessage(chatID int64, text string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.token)

	body := strings.NewReader(fmt.Sprintf(`{"chat_id": %d, "text": %q}`, chatID, text))
	resp, err := b.client.Post(apiURL, "application/json", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if ok, _ := result["ok"].(bool); !ok {
		return fmt.Errorf("telegram API error: %v", result)
	}

	return nil
}

func startTelegramPolling(app *pocketbase.PocketBase, token string) {
	bot := newTelegramBot(token)
	globalTelegramBot = bot
	go pollTelegramUpdates(bot, app)
	log.Printf("Telegram polling started")
}

func pollTelegramUpdates(bot *telegramBot, app *pocketbase.PocketBase) {
	offset := int64(0)
	cycleCount := 0

	for {
		updates, nextOffset, err := bot.getUpdates(offset, 30)
		if err != nil {
			log.Printf("Telegram getUpdates error: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		offset = nextOffset

		// Opportunistic cleanup of expired records every 10 cycles
		cycleCount++
		if cycleCount%10 == 0 {
			now := time.Now().UnixMilli()
			expired, err := app.FindRecordsByFilter(
				"telegram_verifications",
				fmt.Sprintf("expires_at < %d", now),
				"",
				50,
				0,
			)
			if err == nil {
				for _, v := range expired {
					if err := app.Delete(v); err != nil {
						log.Printf("Warning: failed to delete expired verification: %v", err)
					}
				}
			}
		}

		for _, update := range updates {
			if update.Message == nil || update.Message.Chat == nil {
				continue
			}

			chat := update.Message.Chat
			if chat.Type != "private" {
				continue
			}

			text := strings.TrimSpace(update.Message.Text)
			if text != "/start" && text != "start" {
				continue
			}

			chatID := strconv.FormatInt(chat.ID, 10)

			// Invalidate all active verifications for this chat (prevent multiple valid codes)
			now := time.Now().Unix()
			oldVerifications, err := app.FindRecordsByFilter(
				"telegram_verifications",
				"chat_id = {:chat_id} && used_at = 0 && expires_at > {:now}",
				"",
				100,
				0,
				map[string]any{"chat_id": chatID, "now": now},
			)
			if err == nil {
				for _, v := range oldVerifications {
					v.Set("used_at", now)
					if err := app.Save(v); err != nil {
						log.Printf("Warning: failed to invalidate old verification: %v", err)
					}
				}
			}

			// Clean up expired/used verifications for this chat_id
			oldVerifications, err = app.FindRecordsByFilter(
				"telegram_verifications",
				"chat_id = {:chat_id}",
				"",
				100,
				0,
				map[string]any{"chat_id": chatID},
			)
			if err == nil {
				for _, v := range oldVerifications {
					expiresAt := int64(v.GetInt("expires_at"))
					usedAt := v.GetInt("used_at")
					if expiresAt < now || usedAt > 0 {
						if err := app.Delete(v); err != nil {
							log.Printf("Warning: failed to delete old verification: %v", err)
						}
					}
				}
			}

			// Generate new verification code
			code := generateVerificationCode()

			// Create verification record
			collection, err := app.FindCollectionByNameOrId("telegram_verifications")
			if err != nil {
				log.Printf("Error finding telegram_verifications collection: %v", err)
				continue
			}

			record := core.NewRecord(collection)
			record.Set("code", code)
			record.Set("chat_id", chatID)
			record.Set("expires_at", time.Now().Add(5*time.Minute).Unix())
			record.Set("used_at", 0)

			if err := app.Save(record); err != nil {
				log.Printf("Error saving verification record: %v", err)
				continue
			}

			// Send verification code to user
			message := fmt.Sprintf("Your verification code is: %s\n\nEnter this in your account settings.", code)
			if err := bot.sendMessage(chat.ID, message); err != nil {
				log.Printf("Error sending Telegram message: %v", err)
				continue
			}

			log.Printf("Sent verification code to chat_id=%s", chatID)
		}

		time.Sleep(1 * time.Second)
	}
}

func telegramStatusHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		user := e.Auth
		if user == nil {
			return apis.NewApiError(http.StatusUnauthorized, "Not authenticated", nil)
		}

		chatID := user.GetString("telegram_chat_id")
		connected := chatID != ""

		return e.JSON(http.StatusOK, map[string]any{
			"connected": connected,
			"chat_id":  chatID,
		})
	}
}

func telegramVerifyHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		user := e.Auth
		if user == nil {
			return apis.NewApiError(http.StatusUnauthorized, "Not authenticated", nil)
		}

		var data map[string]any
		if err := json.NewDecoder(e.Request.Body).Decode(&data); err != nil {
			return apis.NewApiError(http.StatusBadRequest, "Invalid request body", err)
		}

		code, ok := data["code"].(string)
		if !ok {
			return e.JSON(http.StatusBadRequest, map[string]string{"message": "Code is required"})
		}

		if len(code) != 6 {
			return e.JSON(http.StatusBadRequest, map[string]string{"message": "Code must be 6 digits"})
		}
		for _, c := range code {
			if c < '0' || c > '9' {
				return e.JSON(http.StatusBadRequest, map[string]string{"message": "Code must be 6 digits"})
			}
		}

		now := time.Now().Unix()
		verifications, err := app.FindRecordsByFilter(
			"telegram_verifications",
			"code = {:code} && expires_at > {:now} && used_at = 0",
			"",
			1,
			0,
			map[string]any{"code": code, "now": now},
		)
		if err != nil || len(verifications) == 0 {
			return e.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid or expired code"})
		}

		verification := verifications[0]
		chatID := verification.GetString("chat_id")

		// Mark verification as used immediately to prevent race condition
		verification.Set("used_at", time.Now().UnixMilli())
		if err := app.Save(verification); err != nil {
			return apis.NewApiError(http.StatusInternalServerError, "Failed to consume verification code", err)
		}

		// Check if chat_id is already linked to another user
		otherUsers, err := app.FindRecordsByFilter(
			"users",
			"telegram_chat_id = {:chat_id} && id != {:userId}",
			"",
			1,
			0,
			map[string]any{"chat_id": chatID, "userId": user.Id},
		)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{"message": "Server error during verification"})
		}
		if len(otherUsers) > 0 {
			return e.JSON(http.StatusBadRequest, map[string]string{"message": "This Telegram account is already linked to another user"})
		}

		// If already linked to same user, idempotent success
		if user.GetString("telegram_chat_id") == chatID {
			return e.JSON(http.StatusOK, map[string]any{
				"connected": true,
				"chat_id":  chatID,
			})
		}

		// Link chat_id to user
		user.Set("telegram_chat_id", chatID)
		if err := app.Save(user); err != nil {
			return apis.NewApiError(http.StatusInternalServerError, "Failed to link Telegram account", err)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"connected": true,
			"chat_id":  chatID,
		})
	}
}

func telegramUnsubscribeHandler(app *pocketbase.PocketBase) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		setCORSHeaders(e.Response)

		user := e.Auth
		if user == nil {
			return apis.NewApiError(http.StatusUnauthorized, "Not authenticated", nil)
		}

		user.Set("telegram_chat_id", "")
		if err := app.Save(user); err != nil {
			return apis.NewApiError(http.StatusInternalServerError, "Failed to unlink Telegram account", err)
		}

		return e.JSON(http.StatusOK, map[string]bool{"success": true})
	}
}

func notifyUser(app *pocketbase.PocketBase, userID string, message string) {
	go func() {
		user, err := app.FindRecordById("users", userID)
		if err != nil {
			log.Printf("notifyUser: user not found: %v", err)
			return
		}

		chatIDStr := user.GetString("telegram_chat_id")
		if chatIDStr == "" {
			return // user not subscribed
		}

		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			log.Printf("notifyUser: invalid chat_id format: %v", err)
			return
		}

		if globalTelegramBot == nil {
			return // Telegram not configured
		}

		if err := globalTelegramBot.sendMessage(chatID, message); err != nil {
			log.Printf("notifyUser: failed to send message: %v", err)
		}
	}()
}