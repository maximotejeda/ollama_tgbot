package dbx

import (
	"context"

	"log/slog"
	"strings"
	"time"
)

type history struct {
	ctx          context.Context
	db           *DB
	log          *slog.Logger
	ID           int
	User_ID      int
	Conversation string
	Model_id     int
	Created      string
	Edited       string
	Deleted      string
}

func NewHistory(ctx context.Context, db *DB, log *slog.Logger) *history {
	return &history{ctx: ctx, db: db, log: log}
}

func (h history) QueryHistory(us user) (hList []history, err error) {
	var modelID int
	parts := strings.Split(us.Model, ":")
	model, tag := parts[0], parts[1]
	if err = h.db.QueryRow("SELECT id FROM models WHERE model_name=? AND model_tag=?;", model, tag).Scan(&modelID); err != nil {
		return hList, err
	}
	dt := (time.Now().Add(-1 * time.Hour)).Format(time.RFC3339)
	now := time.Now().Format(time.RFC3339)
	rows, err := h.db.QueryContext(h.ctx, "SELECT conversation FROM 'history' WHERE query_mode=? AND user_id=? AND model_id=? AND created BETWEEN ? AND ?", us.Mode, us.ID, modelID, dt, now)
	if err != nil {
		h.log.Error("error while query", "error", err.Error())
		// Handle the case of no rows returned.
		return []history{}, err
	}
	for rows.Next() {
		hist := history{}
		err := rows.Scan(&hist.Conversation)
		if err != nil {
			return nil, err
		}
		hList = append(hList, hist)
	}
	return hList, nil

}

func (h *history) Add(us user, content string) bool {
	var modelID int
	parts := strings.Split(us.Model, ":")
	model, tag := parts[0], parts[1]
	if err := h.db.QueryRow("SELECT id FROM models WHERE model_name=? AND model_tag=?;", model, tag).Scan(&modelID); err != nil {
		h.log.Error("[history.QUERY] error while preparing ADD", "error", err.Error())
		return false
	}
	now := time.Now().Format(time.RFC3339)

	smt, err := h.db.PrepareContext(h.ctx, "INSERT INTO 'history' (user_id, 'query_mode', 'conversation', model_id, created, edited) VALUES(?,?,?,?,?,?)")
	if err != nil {
		h.log.Error(err.Error())
		panic(err)
	}
	_, err = smt.Exec(us.ID, us.Mode, content, modelID, now, now)
	defer smt.Close()
	if err != nil {
		h.log.Error("[history.ADD] error while ADDing history", "error", err.Error())
		return false
	}

	return true

}
