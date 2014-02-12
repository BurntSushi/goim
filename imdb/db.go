package imdb

import (
	"database/sql"
	"fmt"
	"os"

	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	_ "github.com/lib/pq"

	"github.com/BurntSushi/csql"
	"github.com/BurntSushi/migration"
)

var (
	sf     = fmt.Sprintf
	ef     = fmt.Errorf
	pf     = fmt.Printf
	fatalf = func(f string, v ...interface{}) { pef(f, v...); os.Exit(1) }
	pef    = func(f string, v ...interface{}) {
		fmt.Fprintf(os.Stderr, f+"\n", v...)
	}
	logf = func(format string, v ...interface{}) {
		pef(format, v...)
	}
)

// DB represents a database containing information from the Internet
// Movie DataBase. The underlying database connection is exposed so that
// clients may run their own queries.
type DB struct {
	*sql.DB
	Driver string
}

func Open(driver, dsn string) (*DB, error) {
	db, err := migration.Open(driver, dsn, migrations[driver])
	if err != nil {
		return nil, err
	}
	if driver == "postgres" {
		if _, err := db.Exec("SET timezone = UTC"); err != nil {
			return nil, fmt.Errorf("Could set timezone to UTC: %s", err)
		}
	}
	return &DB{db, driver}, nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) Clean() error {
	tables := []string{"atom", "movie", "tvshow", "episode", "release"}
	return csql.Safe(func() {
		for _, table := range tables {
			csql.Panic(csql.Truncate(db, db.Driver, table))
		}
	})
}

// Empty returns true if and only if the database does not have any data.
// (At the moment, it determines this by only checking the movie table.)
func (db *DB) Empty() bool {
	empty := true
	csql.Safe(func() { // ignore the error, return true
		var count int
		r := db.QueryRow("SELECT COUNT(*) AS count FROM movie")
		csql.Scan(r, &count)
		if count > 0 {
			empty = false
		}
	})
	return empty
}

func (db *DB) IsFuzzyEnabled() bool {
	_, err := db.Exec("SELECT similarity('a', 'a')")
	if err == nil {
		return true
	}
	return false
}
