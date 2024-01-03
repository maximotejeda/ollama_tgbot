package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/maximotejeda/ollama_tgbot/commands"
	"github.com/maximotejeda/ollama_tgbot/dbx"
	"github.com/maximotejeda/ollama_tgbot/helpers"
	"github.com/maximotejeda/ollama_tgbot/query"
	"github.com/maximotejeda/ollama_tgbot/user"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	db := dbx.Dial(ctx, "sqlite", "tg.db")
	defer cancel()

	log.Info("starting bot")

	botApiToken := os.Getenv("TELEGRAM_BOT_TOKEN")
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
		// before anything we verify user is known
		// if messageis != nil we have 3 options
		//    1. is a command: we make system things here
		//    2. Is a response from the command a key board or something: only admins
		//    3. is a prompt: query ollama with the params set on config
		ctx, cancel := context.WithCancel(context.Background())
		msg := tgbot.NewMessage(update.SentFrom().ID, "")
		if ok := helpers.DoWeKnowUser(db, update.SentFrom()); !ok { // if the user is not known
			unknownUserMsg := "sorry this is a private bot\ncontact Maxaltepo to be granted access"

			if update.Message.From.UserName != "" { // without username no query to database is posible
				msgAdmin := user.ConsultAdmin(ctx, db, log, update.Message.Chat.ID, *update.Message.From)
				// send me a message if not found
				if msgAdmin != nil {
					if _, err := bot.Send(*msgAdmin); err != nil {
						panic(err)
					}
				}
			} else {
				unknownUserMsg = unknownUserMsg + "\nremember to set a user name to be able to use the bot"
			}
			msg.Text = unknownUserMsg
			if _, err := bot.Send(msg); err != nil {

				panic(err)
			}
			cancel()
			continue

		}
		if update.Message != nil {
			user := dbx.NewUser(ctx, db, log)
			user.Query(update.Message.From.ID)

			if update.Message.Text != "" && !update.Message.IsCommand() {
				msg = helpers.QueryOllama(ctx, db, log, update.Message.Chat.ID, update.Message.Text, update.Message.From, user)
			} else if update.Message.IsCommand() {
				msg = commands.CommandHandler(ctx, db, log, update.Message.Command(), update.Message.Text, update.Message.Chat.ID)

			}
			if _, err := bot.Send(msg); err != nil {
				panic(err)
			}
			cancel()
		} else if update.CallbackQuery != nil {
			go func(update tgbot.Update){
				query.QueryHandler(ctx, db, log, update.CallbackQuery)
				del := tgbot.NewDeleteMessage(update.CallbackQuery.From.ID, update.CallbackQuery.Message.MessageID)
				
				if _, err := bot.Send(del); err != nil {
					log.Error(err.Error())
				}
				cancel()
			}(update)
		}

	}

}
