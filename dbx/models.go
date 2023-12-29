package dbx

import (
	"context"
	"log/slog"
)

type Model struct {
	ctx       context.Context
	db        *DB
	log       *slog.Logger
	ID        int
	ModelName string
	ModelTag  string
	Created   string
	Edited    string
	Deleted   string
}

func NewModel(ctx context.Context, db *DB, log *slog.Logger) *Model {
	return &Model{ctx: ctx, db: db, log: log}
}

func (m *Model) Query() ([]Model, error) {
	stmt, err := m.db.PrepareContext(m.ctx, "SELECT model_name, model_tag FROM models")
	if err != nil {
		panic(err)
	}
	models := []Model{}
	res, err := stmt.Query()
	if err != nil {
		panic(err)
	}
	for res.Next() {
		var model Model
		if err := res.Scan(&model.ModelName, &model.ModelTag); err != nil {
			return models, err
		}
		models = append(models, model)
	}
	if err = res.Err(); err != nil {
		return models, err
	}
	return models, nil
}
