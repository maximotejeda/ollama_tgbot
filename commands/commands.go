package commands

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/maximotejeda/ollama_tgbot/dbx"
	"github.com/maximotejeda/ollama_tgbot/helpers"
)

// CommandHandler
// Options for user to create queries on the bot
func CommandHandler(ctx context.Context, db *dbx.DB, log *slog.Logger, command, msgTxt string, chatID int64) tgbot.MessageConfig {
	var (
		msg tgbot.MessageConfig
		usr = dbx.NewUser(ctx, db, log)
	)

	usr.Query(chatID)
	msg.ChatID = chatID

	switch strings.ToLower(command) {
	case "model", "modelo":
		mod := dbx.NewModel(ctx, db, log)
		modelsRes, err := mod.Query()
		if err != nil {
			panic(err)
		}
		models := map[string]string{}
		for _, v := range modelsRes {
			models[v.ModelName] = "models=true&name=" + v.ModelName + ":" + v.ModelTag
		}
		keyboard := helpers.CreateKeyboard(models)

		msg.Text = "Different models available:\n\n\torca-mini: fast model for english query\n\n\tllama2: query in english or spanish\n\n\tphi: small query model from MS\n\n\tllama2-uncensored: query in english or spanish without gardrails\n\n\tmedllama2: medical querys to diagnose simptoms\n\n\tcodellama: programming query model\n\n\tmistral: query model on english or spanish"
		msg.ReplyMarkup = keyboard

	case "status", "info":
		btnSTR := map[string]string{"ok": "status=true"}
		keyboard := helpers.CreateKeyboard(btnSTR)
		msg.ReplyMarkup = keyboard
		msg.Text = fmt.Sprintf("User config information\n\tmodel: %s\n\tmode: %s", usr.Model, usr.Mode)
	case "reset":
		reset := map[string]string{"Reset": "reset=true", "Reset All": "resetAll=true"}
		keyboard := helpers.CreateKeyboard(reset)
		msg.ReplyMarkup = keyboard
		msg.Text = "Reset user chat interaction of a concrete model or all interactions with all models."
	case "listusers":
		msg.Text = "Listing distinct users"
	case "mode", "modo":
		modes := map[string]string{"chat": "mode=true&name=chat", "generate": "mode=true&name=generate"}
		keyboard := helpers.CreateKeyboard(modes)
		msg.ReplyMarkup = keyboard
		msg.Text = "Query mode will be changed"
	case "help", "start", "ayuda", "h":
		msg.Text = "Welcome to the bot\nHere are some command to interact with the different models \n\t/help: print this message\n\t/status: print actual query info\n\t/reset: reset chat context\n\t/model: specify the model to work with\n\t/mode: change the query without context"
	default:
		msg.Text = "unknown command try with \n/help: to get bot info."
	}
	return msg
}
