package dbx

import (
	"context"
	"log/slog"
)

type Model interface {
	Query()
}

type model struct {
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

func NewModel(ctx context.Context, db *DB, log *slog.Logger) *model {
	return &model{ctx: ctx, db: db, log: log}
}

func (m *model) Query() ([]model, error) {
	stmt, err := m.db.PrepareContext(m.ctx, "SELECT model_name, model_tag FROM models")
	if err != nil {
		panic(err)
	}
	models := []model{}
	res, err := stmt.Query()
	if err != nil {
		panic(err)
	}
	for res.Next() {
		var model model
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
