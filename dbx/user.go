// operations internal or admin related
package dbx

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

type user struct {
	Ctx        context.Context
	log        *slog.Logger
	db         *DB
	dbTable    string
	Error      error
	ID         int
	TelegramID int64
	Username   string
	FirstName  string
	LastName   string
	Auth       string
	Model      string
	Mode       string
	Created    string
	Edited     string
	Deleted    string
}

func NewUser(ctx context.Context, db *DB, log *slog.Logger) *user {
	if ctx == nil || db == nil || log == nil {
		slog.Default().Error("[NewUser] struct passed is missing")
		return nil
	}
	return &user{Ctx: ctx, log: log, db: db, dbTable: "users"}
}

// create an user on the database
func (u *user) Create(telegramID int64, username, firstname, lastname string) bool {
	slog.Info("")
	if telegramID == 0 || username == "" {
		err := fmt.Errorf("telegram_ID or username can't be empty on user telegram_ID: %d, username: %s", telegramID, username)
		u.log.Error("[user.Create]", "error", err.Error())
		return false
	}
	stmt, err := u.db.PrepareContext(u.Ctx, "INSERT INTO users ('t_id', 'username','first_name', 'last_name', 'auth', model_id, mode, created, edited) VALUES(?,?,?,?,?,?,?,?,?)")
	if err != nil {
		u.log.Error("[user.Create]", "error", err.Error())
		return false
	}
	now := time.Now().Format(time.RFC3339)
	_, err = stmt.ExecContext(u.Ctx, telegramID, username, firstname, lastname, "user", 1, "chat", now, now)
	if err != nil {
		u.log.Error("[user.Create]", "error", err.Error())
		return false
	}
	// we need username first and last name and telegramID
	return true
}

// Look for user on DB
// will populate
// will use only userId
func (u *user) Query(telID int64) bool {
	if telID == 0 {
		err := fmt.Errorf("telegram_ID can't be empty on user search telegram_ID: %d", telID)
		u.log.Error(err.Error())
		return false
	}
	stmt, err := u.db.PrepareContext(u.Ctx, "SELECT users.id, users.t_id, users.username, users.first_name, users.last_name, users.auth, models.model_name || ':' || models.model_tag, users.mode, users.created, users.edited FROM 'users' INNER JOIN models ON users.model_id=models.id WHERE t_id = ?")
	if err != nil {
		u.log.Error("[user.Query stmt]", "error", err.Error())
		panic(err)
	}
	err = stmt.QueryRowContext(u.Ctx, telID).Scan(&u.ID, &u.TelegramID, &u.Username, &u.FirstName, &u.LastName, &u.Auth, &u.Model, &u.Mode, &u.Created, &u.Edited)
	if err != nil {
		if err == sql.ErrNoRows {
			u.log.Warn("no rows found")
			// Handle the case of no rows returned.
		}
		u.log.Error("[user.Query]", "error", err.Error())
		return false
	}
	return true

}

func (u *user) SetModel(modelName string) (bool, error) {
	var modelID int
	parts := strings.Split(modelName, ":")
	model, tag := parts[0], parts[1]
	if err := u.db.QueryRow("SELECT id FROM models WHERE model_name=? AND model_tag=?;", model, tag).Scan(&modelID); err != nil {
		return false, fmt.Errorf("[querying model] model=%s, tag=%s: %w", model, tag, err)
	}
	_, err := u.db.ExecContext(u.Ctx, "UPDATE users SET model_id=? WHERE t_id=?;", modelID, u.TelegramID)
	if err != nil {
		return false, fmt.Errorf("[updating user] %w", err)
	}
	return true, nil
}

//	SetMode
//
// set chat or generate mode on the bot user
func (u *user) SetMode(modeName string) (bool, error) {
	_, err := u.db.ExecContext(u.Ctx, "UPDATE users SET mode=? WHERE t_id=?;", modeName, u.TelegramID)
	if err != nil {
		return false, fmt.Errorf("[updating user] %w", err)
	}
	return true, nil
}

// Edit user on DB
func (u *user) Edit() {}

// Delete user on DB
func (u *user) Delete() {}
