package main

import (
	"context"

	"github.com/uptrace/bun"

	"restuwahyu13/csv-stream/models"
	"restuwahyu13/csv-stream/packages"
)

type service struct {
	ctx    context.Context
	db     *bun.DB
	broker packages.Ikafka
}

func NewService(ctx context.Context, db *bun.DB, broker packages.Ikafka) IService {
	return &service{ctx: ctx, db: db, broker: broker}
}

func (s *service) UpsertUser(req []models.User) error {
	var (
		users     []models.User = []models.User{}
		queueName string        = "csv.stream"
	)

	result, err := s.db.NewInsert().
		Model(&users).
		On("CONFLICT (email) DO UPDATE").
		Set("first_name = EXCLUDED.first_name").
		Set("last_name = EXCLUDED.last_name").
		Set("email = EXCLUDED.email").
		Exec(s.ctx)

	if err != nil {
		packages.Logrus("error", "s.db.NewInsert is error: %s", err.Error())
		users = req
	}

	if rows, err := result.RowsAffected(); err != nil || rows <= 0 {
		packages.Logrus("error", "result.RowsAffected is error: %s", err.Error())
		users = req
	}

	if len(users) > 0 {
		if err := s.broker.Publisher(queueName, nil, users); err != nil {
			return err
		}
	}

	return nil
}
