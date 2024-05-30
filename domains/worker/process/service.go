package main

import (
	"context"
	"os"

	"github.com/panjf2000/ants/v2"
	"github.com/uptrace/bun"

	"restuwahyu13/csv-stream/helpers"
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
		parser    helpers.IParser = helpers.NewParser()
		users     []models.User   = []models.User{}
		queueName string          = "csv.stream"
	)

	gorutinePoolSize, err := parser.ToInt(os.Getenv("GORUTINE_POOL_SIZE"))
	if err != nil {
		return err
	}

	pool, err := ants.NewPoolWithFunc(gorutinePoolSize, func(data interface{}) {
		body := data.([]models.User)

		result, err := s.db.NewInsert().
			Model(&body).
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
				packages.Logrus("error", err)
				return
			}
		}
	},
		ants.WithPreAlloc(true),
		ants.WithNonblocking(true),
	)

	defer pool.Release()
	if err := pool.Invoke(req); err != nil {
		return err
	}

	return nil
}
