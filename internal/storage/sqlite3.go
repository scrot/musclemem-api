package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/VauntDev/tqla"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "modernc.org/sqlite"
)

// Configuration for SqliteDatastore
type SqliteDatastoreConfig struct {
	DatabaseURL        string
	MigrationURL       string
	Overwrite          bool
	ForeignKeyEnforced bool
}

var DefaultSqliteConfig = SqliteDatastoreConfig{
	DatabaseURL:        "file://./musclemem.sqlite",
	MigrationURL:       "file://./migrations/",
	Overwrite:          false,
	ForeignKeyEnforced: true,
}

type SqliteDatastore struct {
	*sql.DB
}

// NewSqliteDatastore creates a new database at dbURL
// and runs the migrations in the defaultMigrations folder
// if overwrite is false, it returns the existing db
func NewSqliteDatastore(config SqliteDatastoreConfig) (*SqliteDatastore, error) {
	path, hasFileScheme := strings.CutPrefix(config.DatabaseURL, "file://")
	if !hasFileScheme {
		err := errors.New("invalid scheme expected file://")
		return nil, err
	}

	if config.Overwrite {
		os.Remove(path)
	}

	dbDNS := fmt.Sprintf("file:%s?_foreign_keys=%t", path, config.ForeignKeyEnforced)
	// dbDNS := "file://.musclemem.sqlite?_pragma=foreign_keys(1)&_time_format=sqlite"
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
		if errors.Is(err, migrate.ErrNoChange) {
			return &SqliteDatastore{db}, nil
		}
		return nil, err
	}

	return &SqliteDatastore{db}, nil
}

// CompileStatement prepares a SQL statement using tqla.Template
// statement will be populated with the data argument like {{ .ID }}
func (ds *SqliteDatastore) CompileStatement(stmt string, data any) (string, []any, error) {
	tmpl, err := tqla.New()
	if err != nil {
		return "", []any{}, err
	}

	return tmpl.Compile(stmt, data)
}
