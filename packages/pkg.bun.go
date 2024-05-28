package packages

import (
	"database/sql"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

func Database(dsn string) (*bun.DB, error) {
	db := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(dsn),
		pgdriver.WithWriteTimeout(time.Minute*time.Duration(10)),
		pgdriver.WithTimeout(time.Minute*time.Duration(1)),
		pgdriver.WithDialTimeout(time.Minute*time.Duration(1)),
		pgdriver.WithReadTimeout(time.Minute*time.Duration(1)),
		pgdriver.WithInsecure(true),
	))

	if err := db.Ping(); err != nil {
		defer Logrus("error", "Database connection error: %v", err)
		return nil, err
	}

	if db != nil {
		defer Logrus("info", "Database connection success")
		db.SetConnMaxIdleTime(time.Duration(time.Second * time.Duration(30)))
		db.SetConnMaxLifetime(time.Duration(time.Second * time.Duration(30)))
	}

	bundb := bun.NewDB(db, pgdialect.New())

	if env := os.Getenv("GO_ENV"); env != "production" {
		bundb.AddQueryHook(bundebug.NewQueryHook(bundebug.WithEnabled(true), bundebug.WithVerbose(true), bundebug.FromEnv("BUNDEBUG")))
	}

	return bundb, nil
}
