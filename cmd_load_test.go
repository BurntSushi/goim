package main

// The testing is pretty pathetic at the moment, but at least there's
// something to build on.
//
// While there's only one test at the moment, it's actually testing a fair
// amount (not exactly an exemplary unit test):

import (
	"io"
	"log"
	"strings"
	"testing"

	"github.com/BurntSushi/csql"

	"github.com/BurntSushi/goim/imdb"
)

var (
	testDB              *imdb.DB
	testDriver, testDsn = "sqlite3", "/tmp/goim-test.sqlite"
)

var (
	testLists = mapFetcher{
		"movies": `
MOVIES LIST
===========
The Matrix (1999)					1999
The Matrix Reloaded (2003)				2003
The Matrix Revolutions (2003)				2003
V for Vendetta (2005)					2005
"The Simpsons" (1989)					1989-????
"The Simpsons" (1989) {Lisa the Iconoclast (#7.16)}	1996
"The Simpsons" (1989) {HOMR (#12.9)}			2001
`,
	}
)

type mapFetcher map[string]string

type readCloser struct {
	io.Reader
}

func (rc readCloser) Close() error {
	return nil
}

func (mf mapFetcher) list(name string) (io.ReadCloser, error) {
	return readCloser{strings.NewReader(mf[name])}, nil
}

func init() {
	var err error
	testDB, err = imdb.Open(testDriver, testDsn)
	if err != nil {
		log.Fatal(err)
	}
}

func TestLoadMovies(t *testing.T) {
	var exp = map[string]int{
		"movies": 4, "tvs": 1, "episodes": 2,
	}
	if err := loadMovies(testDriver, testDsn, testLists); err != nil {
		t.Fatal(err)
	}
	movies := csql.Count(testDB, "SELECT COUNT(*) FROM movie")
	tvs := csql.Count(testDB, "SELECT COUNT(*) FROM tvshow")
	episodes := csql.Count(testDB, "SELECT COUNT(*) FROM episode")
	if movies != exp["movies"] {
		t.Fatalf("Expected %d movies but got %d", exp["movies"], movies)
	}
	if tvs != exp["tvs"] {
		t.Fatalf("Expected %d tvs but got %d", exp["tvs"], tvs)
	}
	if episodes != exp["episodes"] {
		t.Fatalf("Expected %d episodes but got %d", exp["episodes"], episodes)
	}
}
