package imdb

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/BurntSushi/csql"
	"github.com/BurntSushi/migration"
)

var (
	sf = fmt.Sprintf
	ef = fmt.Errorf
)

// DB represents a database containing information from the Internet
// Movie DataBase. The underlying database connection is exposed so that
// clients may run their own queries.
type DB struct {
	*sql.DB

	// Since this package attempts to support multiple databases, there are
	// areas where the type of driver being used is important.
	// For example, PostgreSQL supports simultaneous transactions updating the
	// database but SQLite does not.
	Driver string
}

// Open opens a connection to an IMDb relational database. The driver may
// either be "sqlite3" or "postgres". The dsn (data source name) is dependent
// upon the driver. For example, for the sqlite3 driver, the dsn is just a
// path to a file (that may not exist).
//
// In general, the 'driver' and 'dsn' should be exactly the same as used in
// the 'database/sql' package.
//
// Whenever an imdb database is opened, it is checked to make sure its schema
// is up to date with the current library. If it isn't, it will be updated.
func Open(driver, dsn string) (*DB, error) {
	db, err := migration.Open(driver, dsn, migrations[driver])
	if err != nil {
		return nil, err
	}
	if driver == "postgres" {
		if _, err := db.Exec("SET timezone = UTC"); err != nil {
			return nil, fmt.Errorf("Could not set timezone to UTC: %s", err)
		}
	}
	return &DB{db, driver}, nil
}

// Close closes the connection to the database.
func (db *DB) Close() error {
	return db.DB.Close()
}

// Tables returns the names of all tables in the database sorted
// alphabetically in ascending order.
func (db *DB) Tables() (tables []string, err error) {
	defer csql.Safe(&err)

	var q string
	switch db.Driver {
	case "postgres":
		q = `
			SELECT tablename FROM pg_tables
			WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
			ORDER BY tablename ASC
		`
	case "sqlite3":
		q = `
			SELECT tbl_name FROM sqlite_master
			WHERE type = 'table'
			ORDER BY tbl_name ASC
		`
	default:
		return nil, ef("Unrecognized database driver: %s", db.Driver)
	}
	rows := csql.Query(db, q)
	csql.Panic(csql.ForRow(rows, func(rs csql.RowScanner) {
		var table string
		csql.Scan(rs, &table)
		if table != "migration_version" {
			tables = append(tables, table)
		}
	}))
	return
}

// IsFuzzyEnabled returns true if and only if the database is a Postgres
// database with the 'pg_trgm' extension enabled.
func (db *DB) IsFuzzyEnabled() bool {
	_, err := db.Exec("SELECT similarity('a', 'a')")
	if err == nil {
		return true
	}
	return false
}
