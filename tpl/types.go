package tpl

import (
	"fmt"
	"os"

	"github.com/BurntSushi/goim/imdb"
	"github.com/BurntSushi/goim/imdb/search"
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

// assert will quit Goim with the specified error if it is not nil.
func assert(err error) {
	if err != nil {
		fatalf("%s", err)
	}
}

// assertTwo will quit Goim with the specified error if it is not nil.
// Otherwise, the first value given is returned.
func assertTwo(v interface{}, err error) interface{} {
	assert(err)
	return v
}

// assertDB makes sure there is a valid DB connection.
func assertDB() {
	if tplDB == nil {
		assert(ef("No database connection found. Please set one with SetDB."))
	}
}

// FromSearchResult translates a search result to an appropriate template type
// in this package. Such values are intended to be used inside Goim templates.
//
// If there was a problem translating the value, Goim will quit with an error
// message.
func FromSearchResult(db *imdb.DB, sr search.Result) interface{} {
	return nil
}

func fromAtom(db *imdb.DB, ent imdb.EntityKind, id imdb.Atom) interface{} {
	return nil
}
