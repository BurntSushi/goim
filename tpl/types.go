package tpl

import (
	"fmt"
	"os"

	"github.com/BurntSushi/csql"
	"github.com/BurntSushi/goim/imdb"
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

// FromSearchResult translates a search result to an appropriate template type
// in this package. Such values are intended to be used inside Goim templates.
//
// If there was a problem translating the value, Goim will quit with an error
// message.
func FromSearchResult(db *imdb.DB, sr imdb.SearchResult) interface{} {
	return fromAtom(db, sr.Entity, sr.Id)
}

func fromAtom(db *imdb.DB, ent imdb.Entity, id imdb.Atom) interface{} {
	switch ent {
	case imdb.EntityMovie:
		m, err := imdb.AtomToMovie(db, id)
		assert(err)
		return Movie{db, m}
	case imdb.EntityTvshow:
		t, err := imdb.AtomToTvshow(db, id)
		assert(err)
		return Tvshow{db, t}
	case imdb.EntityEpisode:
		e, err := imdb.AtomToEpisode(db, id)
		assert(err)
		return Episode{db, e}
	}
	fatalf("Unrecognized entity type: %s", ent)
	panic("unreachable")
}

type Movie struct {
	db *imdb.DB
	imdb.Movie
}

type Tvshow struct {
	db *imdb.DB
	imdb.Tvshow
}

type Episode struct {
	db *imdb.DB
	imdb.Episode
}

func (e Episode) Tvshow() Tvshow {
	return fromAtom(e.db, imdb.EntityTvshow, e.TvshowId).(Tvshow)
}

func releaseDates(
	db *imdb.DB,
	getDates func(csql.Queryer) ([]imdb.ReleaseDate, error),
) []imdb.ReleaseDate {
	dates, err := getDates(db)
	assert(err)
	return dates
}

func (m Movie) ReleaseDates() []imdb.ReleaseDate {
	return releaseDates(m.db, m.Movie.ReleaseDates)
}

func (t Tvshow) ReleaseDates() []imdb.ReleaseDate {
	return releaseDates(t.db, t.Tvshow.ReleaseDates)
}

func (e Episode) ReleaseDates() []imdb.ReleaseDate {
	return releaseDates(e.db, e.Episode.ReleaseDates)
}
