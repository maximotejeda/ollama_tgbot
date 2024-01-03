package helpers

import (
	"context"
	"log/slog"
	"os"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/maximotejeda/ollama_tgbot/dbx"
	"github.com/maximotejeda/ollama_tgbot/ollama"
)

// DoWeKnowUser
// Limit users who can or not use the app
func DoWeKnowUser(db *dbx.DB, user *tgbot.User) bool {
	ctx := context.Background()
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	use := dbx.NewUser(ctx, db, log)
	boo := use.Query(user.ID)
	//log.Warn("user query", "result", boo)
	return boo
}

// QueryOllama
// Query ollama with the user params
func QueryOllama(ctx context.Context, db *dbx.DB, log *slog.Logger, chatID int64, query string, us *tgbot.User, user *dbx.User) tgbot.MessageConfig {
	oClient := ollama.NewOllamaClient(ctx, db, log, user)
	res := oClient.Do(query)
	msg := tgbot.NewMessage(chatID, res)
	return msg
}

// CreateKeyboard
// create keybowrds of two rows of any map[string]string input
func CreateKeyboard(data map[string]string) tgbot.InlineKeyboardMarkup {
	// hardcoded models
	keyboard := tgbot.NewInlineKeyboardMarkup()
	//	subbuttons := []tgbot.InlineKeyboardButton{}
	rows := tgbot.NewInlineKeyboardRow()
	counter := 0
	for key, val := range data {

		if counter != 0 && counter%2 == 0 {
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, rows)
			rows = tgbot.NewInlineKeyboardRow()
		}
		rows = append(rows, tgbot.NewInlineKeyboardButtonData(key, val))
		if counter >= len(data)-1 {
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, rows)
		}
		counter++
	}
	return keyboard
}
