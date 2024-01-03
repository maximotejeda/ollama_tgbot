// Verify and authenticates users with the database
// we will have different ty0pes of users
// admin users the users that can do anything and correct delete and update anything
// sellers users will be on charge of billing and receiving money
// delivery users will be the ones on sending products and transport goods
//
// each user will have different type of commands
// for example when a delivery user receive a shipping will be marked as on delivery
// when the goods are delivered we mark them as done
package user

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/maximotejeda/ollama_tgbot/dbx"
	"github.com/maximotejeda/ollama_tgbot/helpers"
)

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
		replyKeyboard := helpers.CreateKeyboard(data)
		msg.ReplyMarkup = replyKeyboard
	}
	return &msg
}
