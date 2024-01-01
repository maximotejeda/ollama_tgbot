package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/maximotejeda/ollama_tgbot/dbx"
	"github.com/maximotejeda/ollama_tgbot/ollama"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	db := dbx.Dial(ctx, "sqlite", "tg.db")
	defer cancel()

	log.Info("starting bot")

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
		// before anything we verify user is known
		// if messageis != nil we have 3 options
		//    1. is a command: we make system things here
		//    2. Is a response from the command a key board or something: only admins
		//    3. is a prompt: query ollama with the params set on config
		ctx, cancel := context.WithCancel(context.Background())
		msg := tgbot.NewMessage(update.SentFrom().ID, "")
		if ok := DoWeKnowUser(db, update.SentFrom()); !ok { // if the user is not known
			unknownUserMsg := "sorry this is a private bot\ncontact Maxaltepo to be granted access"

			if update.Message.From.UserName != "" { // without username no query to database is posible
				msgAdmin := ConsultAdmin(ctx, db, log, update.Message.Chat.ID, *update.Message.From)
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
			//log.Info(fmt.Sprintf("[%s] %s", update.Message.From.UserName, update.Message.Text))

			if update.Message.Text != "" && !update.Message.IsCommand() {
				msg = QueryOllama(ctx, db, log, update.Message.Chat.ID, update.Message.Text, update.Message.From, user)
			} else if update.Message.IsCommand() {
				msg = CommandHandler(ctx, db, log, update.Message.Command(), update.Message.Text, update.Message.Chat.ID)

			}
			MessageHandler(msg, update.Message.MessageID)
			if _, err := bot.Send(msg); err != nil {
				panic(err)
			}
			cancel()
		} else if update.CallbackQuery != nil {
			//msg = tgbot.NewMessage(update.CallbackQuery.From.ID, "")
			// callback := tgbot.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			//	log.Warn("====> message", "value", update.CallbackQuery.Data)

			QueryHandler(ctx, db, log, update.CallbackQuery)
			del := tgbot.NewDeleteMessage(update.CallbackQuery.From.ID, update.CallbackQuery.Message.MessageID)

			if _, err := bot.Send(del); err != nil {
				log.Error(err.Error())
				continue
			}
		}

	}

}

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
func QueryOllama(ctx context.Context, db *dbx.DB, log *slog.Logger, chatID int64, query string,us *tgbot.User, user *dbx.User) tgbot.MessageConfig {
	oClient := ollama.NewOllamaClient(ctx, db, log, user)
	res := oClient.Do(query)
	msg := tgbot.NewMessage(chatID, res)
	return msg
}

// MessageHandler
// mainly only reply text sent
func MessageHandler(msg tgbot.MessageConfig, replyID int) tgbot.Chattable {
	msg.ReplyToMessageID = replyID
	return msg
}

// CommandHandler
// Options for user to create queries on the bot
func CommandHandler(ctx context.Context, db *dbx.DB, log *slog.Logger, command, msgTxt string, chatID int64) tgbot.MessageConfig {
	var (
		msg tgbot.MessageConfig
		usr = dbx.NewUser(ctx,db,log)
	)
	
	usr.Query(chatID)
	value := strings.Split(msgTxt, " ")
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
		keyboard := CreateKeyboard(models)
		if len(value) < 2 {
			msg.Text = "Models available to chat or generate prompts."
			msg.ReplyMarkup = keyboard
		} else {
			msg.Text = "changing model " + value[1]
		}
	case "status", "info":
		btnSTR := map[string]string{"ok": "status=true"}
		keyboard := CreateKeyboard(btnSTR)
		msg.ReplyMarkup = keyboard
		msg.Text = fmt.Sprintf("User config information\n\tmodel: %s\n\tmode: %s",usr.Model, usr.Mode)
	case "reset":
		reset := map[string]string{"Reset": "reset=true", "Reset All": "resetAll=true"}
		keyboard := CreateKeyboard(reset)
		msg.ReplyMarkup = keyboard
		msg.Text = "Reset user chat interaction of a concrete model or all interactions with all models."
	case "listusers":
		msg.Text = "Listing distinct users"
	case "mode", "modo":
		modes := map[string]string{"chat": "mode=true&name=chat", "generate": "mode=true&name=generate"}
		keyboard := CreateKeyboard(modes)
		msg.ReplyMarkup = keyboard
		msg.Text = "Query mode will be changed"
	case "help", "start", "ayuda", "h":
		msg.Text = "Welcome to the bot\nHere are some command to interact with the different models \n\t/help: print this message\n\t/status: print actual query info\n\t/reset: reset chat context\n\t/model: specify the model to work with\n\t/mode: change the query without context"
	default:
		msg.Text = "unknown command try with \n/help: to get bot info."
	}
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

		if (counter!= 0 && counter %2 == 0) {
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, rows)
			rows = tgbot.NewInlineKeyboardRow() 
		}
		rows = append(rows, tgbot.NewInlineKeyboardButtonData(key, val))
		if counter >= len(data)-1{
			keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, rows)
		}
		counter ++
	}
	return keyboard
}

// ConsultAdmin
// ask admin hardcoded on env to grant access
func ConsultAdmin(ctx context.Context, db *dbx.DB, log *slog.Logger, chatID int64, user tgbot.User) *tgbot.MessageConfig {
	adminID, err := strconv.ParseInt(os.Getenv("TELEGRAM_ADMIN_ID"), 10, 64)
	if err != nil {
		log.Error("needs an admin on db to manage users", "error", err.Error())
		return nil
	}
	admin := dbx.NewUser(ctx, db, log)
	txt := fmt.Sprintf("user %s \nID %d\nfirstName %s\nlastname: %s\n is trying to enter the chat with bot, acess was requested", user.UserName, user.ID, user.FirstName, user.LastName)

	msg := tgbot.NewMessage(adminID, txt)

	if ok := admin.Query(adminID); ok {
		queryData := fmt.Sprintf("tid=%d&uname=%s", user.ID, user.UserName)

		data := map[string]string{"delete": "del=true&" + queryData, "add": "add=true&" + queryData}
		replyKeyboard := CreateKeyboard(data)
		msg.ReplyMarkup = replyKeyboard
	}
	return &msg
}

// QueryHandler
// Manage queries to execute user commands
func QueryHandler(ctx context.Context, db *dbx.DB, log *slog.Logger, query *tgbot.CallbackQuery) (msg *tgbot.MessageConfig) {
	tUser := query.From
	user := dbx.NewUser(ctx, db, log)
	user.Query(tUser.ID)
	data := query.Data
	dataList := strings.Split(data, "&")
	dataMap := map[string]string{}
	for _, val := range dataList {
		subData := strings.Split(val, "=")
		//log.Warn(subData[0])
		dataMap[subData[0]] = subData[1]
	}

	switch {
	case dataMap["del"] != "" || dataMap["add"] != "":
		user1 := dbx.NewUser(ctx, db, log)
		telegramID, ok := dataMap["tid"]
		
		id, err := strconv.ParseInt(telegramID, 10, 64)
		if err != nil {
			log.Error("converting to in64", "error", err.Error())
		}
		if !ok {
			msg.Text = "no tID provided"
		}
		username, ok := dataMap["uname"]
		if !ok {
			msg.Text = "no username provided"
		}
		if dataMap["add"] != "" {
			//log.Warn("info from map", "username:", username, "tgID:", id, dataMap)
			if id == 0 {
				log.Error("bad key conversion", "id", id)
				break
			}
			res := user1.Create(id, username, "", "")
			if !res {
				log.Error("error Creating user on main function")
			}
		} else if dataMap["del"] != "" {
			log.Info("blocking user")
		}

	case dataMap["models"] != "":
		user.Query(tUser.ID)
		_, err := user.SetModel(dataMap["name"])
		if err != nil {
			log.Error("", "error", err.Error())
		}
	case dataMap["mode"] != "":
		user.Query(tUser.ID)
		_, err := user.SetMode(dataMap["name"])
		if err != nil {
			log.Error("", "error", err.Error())
		}
	case dataMap["reset"] != "":
		h := dbx.NewHistory(ctx, db, log)
		ok := h.Reset(*user)
		if !ok {
			log.Error("[query reset] error")
		}
	case dataMap["resetAll"] != "":
		h := dbx.NewHistory(ctx, db, log)
		ok := h.ResetAll(*user)
		if !ok {
			log.Error("[query reset] error")
		}

	}

	return msg
}
