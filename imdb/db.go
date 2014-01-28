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
	Driver    string
	inserters []*Inserter
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
	return &DB{db, driver, nil}, nil
}

func (db *DB) Close() error {
	for _, ins := range db.inserters {
		if err := ins.Exec(); err != nil {
			return err
		}
	}
	for _, ins := range db.inserters {
		if err := ins.Close(); err != nil {
			return err
		}
	}
	return db.DB.Close()
}

func (db *DB) Clean() error {
	tables := []string{"atom", "movie", "tvshow", "episode", "release"}
	return csql.Safe(func() {
		for _, table := range tables {
			csql.SQLPanic(csql.Truncate(db, db.Driver, table))
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

func (db *DB) Begin() (*Tx, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{db, false, tx}, nil
}

type Tx struct {
	db     *DB
	closed bool
	*sql.Tx
}

func (tx *Tx) Another() (*Tx, error) {
	if tx.db.Driver == "sqlite3" {
		return tx, nil
	}
	txx, err := tx.db.Begin()
	if err != nil {
		return nil, err
	}
	return txx, nil
}

func (tx *Tx) Commit() error {
	if tx.closed {
		return nil
	}
	tx.closed = true
	return tx.Tx.Commit()
}

func (tx *Tx) Rollback() error {
	if tx.closed {
		return nil
	}
	tx.closed = true
	return tx.Tx.Rollback()
}
