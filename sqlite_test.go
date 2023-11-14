package musclememapi

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestCreateLocalDatabasePingableOnValidPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := newLocalDatabase(path)
	if err != nil {
		t.Fatalf("expected no error but got %s", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("expected no error but got %s", err)
	}
}

func TestCreateLocalDatabaseErrOnEmptyPath(t *testing.T) {
	path := ""
	_, err := newLocalDatabase(path)
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
}

func TestCreateLocalDatabaseErrOnInvalidPath(t *testing.T) {
	path := "///test.db"
	_, err := newLocalDatabase(path)
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
}

func TestCreateLocalDatabaseErrOnURL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	url := "sqlite3://" + path
	_, err := newLocalDatabase(url)
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
}

func tablesEqual(t *testing.T, db *sql.DB, wantTables []string) ([]string, bool) {
	t.Helper()

	stmt := `
  SELECT 
    name
  FROM 
    sqlite_schema
  WHERE 
    type = 'table' AND 
    name != 'schema_migrations' AND
    name NOT LIKE 'sqlite_%';
  `

	rows, err := db.Query(stmt)
	if err != nil {
		t.Fatalf("expected no error but got %s", err)
	}

	var tables []string
	for rows.Next() {
		var table string
		rows.Scan(&table)
		tables = append(tables, table)
	}

	less := func(a, b string) bool { return a < b }
	if cmp.Equal(tables, wantTables, cmpopts.SortSlices(less)) {
		return tables, true
	}

	return tables, false
}

func TestNewSqliteDatastore(t *testing.T) {
	dbURL := fmt.Sprintf("file://%s/%s", t.TempDir(), "test.db")
	db, err := NewSqliteDatastore(dbURL, true)
	if err != nil {
		t.Errorf("expected no error but got %s", err)
	}

	t.Run("DeleteExistingIfOverwrite", func(t *testing.T) {
		db.Exec("DROP TABLE IF EXISTS exercises;")

		db, err = NewSqliteDatastore(dbURL, true)
		if err != nil {
			t.Errorf("expected no error but got %s", err)
		}

		wantTables := []string{"exercises", "users"}
		if gotTables, eq := tablesEqual(t, db.DB, wantTables); !eq {
			t.Errorf("expected tables %s but got %s", wantTables, gotTables)
		}
	})

	t.Run("OpenExistingIfNotOverwrite", func(t *testing.T) {
		db.Exec("DROP TABLE IF EXISTS exercises;")

		db, err = NewSqliteDatastore(dbURL, false)
		if err != nil {
			t.Errorf("expected no error but got %s", err)
		}

		wantTables := []string{"users"}
		if gotTables, eq := tablesEqual(t, db.DB, wantTables); !eq {
			t.Errorf("expected tables %s but got %s", wantTables, gotTables)
		}
	})
}
