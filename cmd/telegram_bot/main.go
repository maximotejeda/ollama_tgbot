package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/maximotejeda/tgbot/ollama"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	log.Info("starting bot")
	var msg tgbot.MessageConfig
	botApiToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	//ollamaApi := os.Getenv("OLLAMA_API")

	bot, err := tgbot.NewBotAPI(botApiToken)
	if err != nil {
		log.Error("init bot: ", err.Error(), botApiToken)
		os.Exit(1)
	}
	bot.Debug = true

	log.Info(
		fmt.Sprintf("Authorized on account %s", bot.Self.UserName))

	u := tgbot.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		// we can have updates on messages or commands
		// update.Message.IsCommand()
		// My plan is to pass the chat for user check auth of user on the system
		// then pass it through command and set the commands available to that chat
		// then work in conjunction of that
		if update.Message != nil {
			log.Info(fmt.Sprintf("[%s] %s", update.Message.From.UserName, update.Message.Text))

			ctx, cancel := context.WithCancel(context.Background())

			if ok := DoWeKnowUser(update.Message.From); !ok {
				unknownUserMsg := "sorry this is a private bot\ncontact Maxaltepo to be granted access"
				msg = tgbot.NewMessage(update.Message.Chat.ID, unknownUserMsg)
				msg.ReplyToMessageID = update.Message.MessageID

			} else {
				msg = QueryOllama(ctx, update.Message.Chat.ID, update.Message.Text)

			}

			MessageHandler(msg, update.Message.MessageID)
			if _, err := bot.Send(msg); err != nil {

				panic(err)
			}
			cancel()
		}

	}

}

// query DB to manage users
func DoWeKnowUser(user *tgbot.User) bool {
	fmt.Printf("username: %s with ID: %d\n\tFirstName%s\n\tLastName:%s ", user.UserName, user.ID, user.FirstName, user.LastName)
	return true
}

// Query ollama to manage ollama
func QueryOllama(ctx context.Context, chatID int64, query string) tgbot.MessageConfig {
	var msg tgbot.MessageConfig
	res := ollama.OllamaConsult(ctx, query)
	msg = tgbot.NewMessage(chatID, res)
	return msg
}

// finally decide over the senmding message
func MessageHandler(msg tgbot.MessageConfig, replyID int) tgbot.Chattable {
	msg.ReplyToMessageID = replyID
	return msg
}
