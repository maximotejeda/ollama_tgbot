package query

import (
	"context"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/maximotejeda/ollama_tgbot/dbx"
	"log/slog"
	"strconv"
	"strings"
)

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
		// mixtral context breaks my PC 32GB no gpu
		if dataMap["name"] == "mistral:latest" {
			user.SetMode("generate")
		}
	case dataMap["mode"] != "":
		user.Query(tUser.ID)
		if user.Model != "mistral:latest" {
			_, err := user.SetMode(dataMap["name"])
			if err != nil {
				log.Error("", "error", err.Error())
			}
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
