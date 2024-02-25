package storage

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/cockroachdb"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func newCockroach(dburl, path string) (*SqlDatastore, error) {
	db, err := sql.Open("pgx", dburl)
	if err != nil {
		return nil, err
	}

	driver, err := cockroachdb.WithInstance(db, &cockroachdb.Config{})
	if err != nil {
		return nil, err
	}

	source, err := iofs.New(fs, path)
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance("embedded-migrations", source, "cockroach-db", driver)
	if err != nil {
		return nil, err
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return &SqlDatastore{db}, nil
		}
		return nil, err
	}

	return &SqlDatastore{db}, nil
}
