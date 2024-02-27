package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/VauntDev/tqla"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func newLocalSqlite(dburl, path string, overwrite bool) (*SqlDatastore, error) {
	dbpath, _ := strings.CutPrefix(dburl, "file://")
	if overwrite {
		os.Remove(path)
	}

	dbDNS := fmt.Sprintf("file:%s?_foreign_keys=%t", dbpath, true)
	db, err := sql.Open("sqlite", dbDNS)
	if err != nil {
		return nil, err
	}

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
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

	placeholder := tqla.WithPlaceHolder(tqla.Question)

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return &SqlDatastore{db, placeholder}, nil
		}
		return nil, err
	}

	return &SqlDatastore{db, placeholder}, nil
}
