package internal

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

// Configuration for SqliteDatastore
type SqliteDatastoreConfig struct {
	DatabaseURL        string
	MigrationURL       string
	Overwrite          bool
	ForeignKeyEnforced bool
}

var DefaultSqliteConfig = SqliteDatastoreConfig{
	DatabaseURL:        "file://./musclemem.db",
	MigrationURL:       "file://./migrations/",
	Overwrite:          false,
	ForeignKeyEnforced: true,
}

// NewSqliteDatastore creates a new database at dbURL
// and runs the migrations in the defaultMigrations folder
// if overwrite is false, it returns the existing db
func NewSqliteDatastore(config SqliteDatastoreConfig) (*sql.DB, error) {
	path, hasFileScheme := strings.CutPrefix(config.DatabaseURL, "file://")
	if !hasFileScheme {
		err := errors.New("invalid scheme expected file://")
		return nil, err
	}

	if config.Overwrite {
		os.Remove(path)
	}

	dbDNS := fmt.Sprintf("file:%s?_foreign_keys=%t", path, config.ForeignKeyEnforced)
	db, err := sql.Open("sqlite3", dbDNS)
	if err != nil {
		return nil, err
	}

	drv, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(config.MigrationURL, config.DatabaseURL, drv)
	if err != nil {
		return nil, err
	}

	if err := m.Up(); err != nil {
		return nil, err
	}

	return db, nil
}
