package dbx

import (
	"context"

	"log/slog"
	"strings"
	"time"
)

type History interface {
	Query()
	Add()
}

type history struct {
	ctx          context.Context
	db           *DB
	log          *slog.Logger
	ID           int
	User_ID      int
	Role         string
	Conversation string
	Model_id     int
	Created      string
	Edited       string
	Deleted      string
}

// NewHistory
// returns an instance of history to interact with
func NewHistory(ctx context.Context, db *DB, log *slog.Logger) *history {
	return &history{ctx: ctx, db: db, log: log}
}

// Query
// look for recent user interaction with the assistant
// the default time to remember is 10 minutes
// the default amount of message to retrieve are 5
func (h history) Query(us User) (hList []history, err error) {

	var modelID int
	parts := strings.Split(us.Model, ":")
	model, tag := parts[0], parts[1]
	if err = h.db.QueryRow("SELECT id FROM models WHERE model_name=? AND model_tag=?;", model, tag).Scan(&modelID); err != nil {
		return hList, err
	}
	dt := (time.Now().Add(-10 * time.Minute)).Format(time.DateTime)
	now := time.Now().Format(time.DateTime)
	rows, err := h.db.QueryContext(h.ctx, "SELECT role, conversation, created FROM (SELECT role, conversation, created FROM 'history' WHERE query_mode=? AND user_id=? AND model_id=? AND created BETWEEN ? AND ? ORDER BY created DESC LIMIT 5) AS r ORDER BY created ", us.Mode, us.ID, modelID, dt, now)
	if err != nil {
		h.log.Error("error while query", "error", err.Error())
		// Handle the case of no rows returned.
		return []history{}, err
	}
	defer h.db.Exec("DELETE FROM history WHERE user_id = ? AND created <= ?;", us.ID, dt)
	for rows.Next() {
		hist := history{}
		err := rows.Scan(&hist.Role, &hist.Conversation, &hist.Created)
		if err != nil {
			return nil, err
		}
		hList = append(hList, hist)
	}
	return hList, nil

}

// Add
// add a user, assistant interaction to the db
func (h *history) Add(us User, content, role string) bool {
	var modelID int
	parts := strings.Split(us.Model, ":")
	model, tag := parts[0], parts[1]
	if err := h.db.QueryRow("SELECT id FROM models WHERE model_name=? AND model_tag=?;", model, tag).Scan(&modelID); err != nil {
		h.log.Error("[history.QUERY] error while preparing ADD", "error", err.Error())
		return false
	}
	now := time.Now().Format(time.DateTime)

	smt, err := h.db.PrepareContext(h.ctx, "INSERT INTO 'history' (user_id, 'query_mode', 'role', 'conversation', model_id, created, edited) VALUES(?,?,?,?,?,?,?)")
	if err != nil {
		h.log.Error(err.Error())
		panic(err)
	}
	_, err = smt.Exec(us.ID, us.Mode, role, content, modelID, now, now)
	defer smt.Close()
	if err != nil {
		h.log.Error("[history.ADD] error while ADDing history", "error", err.Error())
		return false
	}
	return true
}

// Reset
// reset a user history for a given model
func (h *history) Reset(us User) bool {
	uid := us.ID
	var modelID int
	parts := strings.Split(us.Model, ":")
	model, tag := parts[0], parts[1]
	if err := h.db.QueryRow("SELECT id FROM models WHERE model_name=? AND model_tag=?;", model, tag).Scan(&modelID); err != nil {
		h.log.Error("[history.QUERY] error while preparing ADD", "error", err.Error())
		return false
	}

	stmt, err := h.db.PrepareContext(h.ctx, "DELETE FROM 'history' WHERE user_id = ? AND model_id = ?;")
	if err != nil {
		h.log.Error("[history.Reset]", "error", err.Error())
		panic(err)

	}
	_, err = stmt.Exec(uid, modelID)
	defer stmt.Close()
	if err != nil {
		h.log.Error("[history.Reset] error while ADDing history", "error", err.Error())
		return false
	}
	return true
}

// ResetAll
// reset user history on all models at a time
func (h *history) ResetAll(us User) bool {

	stmt, err := h.db.PrepareContext(h.ctx, "DELETE FROM 'history' WHERE user_id = ?;")
	if err != nil {
		h.log.Error("[history.Reset]", "error", err.Error())
		panic(err)

	}
	_, err = stmt.Exec(us.ID)
	defer stmt.Close()
	if err != nil {
		h.log.Error("[history.Reset All] error while ADDing history", "error", err.Error())
		return false
	}
	return true
}
