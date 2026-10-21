package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type Storage struct {
	DB *sql.DB
}

func ConnToStorage(ctx context.Context, dbUrl string) (*Storage, error) {
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return nil, err
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	if err = RunMigrations(dbUrl); err != nil {
		return nil, err
	}

	return &Storage{
		DB: db,
	}, nil
}

func RunMigrations(databaseURL string) error {
	m, err := migrate.New(
		"file://migrations",
		databaseURL,
	)

	if err != nil {
		return fmt.Errorf("initialize migrations: %w", err)
	}

	defer func() {
		_, _ = m.Close()
	}()

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
