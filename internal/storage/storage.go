package storage

import (
	"database/sql"
	"embed"
	"errors"
	"strings"

	"github.com/VauntDev/tqla"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

//go:embed migrations
var fs embed.FS

// Configuration for SqliteDatastore
type DatastoreConfig struct {
	DatabaseURL   string
	MigrationPath string
	Overwrite     bool
}

var DefaultSqliteConfig = DatastoreConfig{
	DatabaseURL:   "file://./musclemem.sqlite",
	MigrationPath: "migrations",
	Overwrite:     false,
}

type SqlDatastore struct {
	*sql.DB
	placeholder tqla.Option
}

// NewSqliteDatastore creates a new database at dbURL
// and runs the migrations in the defaultMigrations folder
// if overwrite is false, it returns the existing db
func NewSqlDatastore(config DatastoreConfig) (*SqlDatastore, error) {
	scheme, _, ok := strings.Cut(config.DatabaseURL, "://")
	if !ok {
		return nil, errors.New("invalid url " + config.DatabaseURL)
	}

	switch scheme {
	case "file":
		return newLocalSqlite(config.DatabaseURL, config.MigrationPath, config.Overwrite)
	case "postgresql":
		return newCockroach(config.DatabaseURL, config.MigrationPath)
	default:
		err := errors.New("invalid scheme " + scheme)
		return nil, err
	}
}

// CompileStatement prepares a SQL statement using tqla.Template
// statement will be populated with the data argument like {{ .ID }}
func (ds *SqlDatastore) CompileStatement(stmt string, data any) (string, []any, error) {
	tmpl, err := tqla.New(ds.placeholder)
	if err != nil {
		return "", []any{}, err
	}

	return tmpl.Compile(stmt, data)
}
