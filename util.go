package main

import (
	"fmt"

	"github.com/BurntSushi/goim/imdb"

	"os"
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
		if flagQuiet {
			return
		}
		pef(format, v...)
	}
)

func createFile(fpath string) *os.File {
	f, err := os.Create(fpath)
	if err != nil {
		fatalf(err.Error())
	}
	return f
}

func openFile(fpath string) *os.File {
	f, err := os.Open(fpath)
	if err != nil {
		fatalf(err.Error())
	}
	return f
}

func openDb(driver, dsn string) *imdb.DB {
	db, err := imdb.Open(driver, dsn)
	if err != nil {
		fatalf("Could not open '%s:%s': %s", driver, dsn, err)
	}
	return db
}

func closeDb(db *imdb.DB) {
	if err := db.Close(); err != nil {
		fatalf("Could not close database: %s", err)
	}
}
