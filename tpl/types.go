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

// assertTwo will quit Goim with the specified error if it is not nil.
// Otherwise, the first value given is returned.
func assertTwo(v interface{}, err error) interface{} {
	assert(err)
	return v
}

// FromSearchResult translates a search result to an appropriate template type
// in this package. Such values are intended to be used inside Goim templates.
//
// If there was a problem translating the value, Goim will quit with an error
// message.
func FromSearchResult(db *imdb.DB, sr imdb.SearchResult) interface{} {
	return fromAtom(db, sr.Entity, sr.Id)
}

func fromAtom(db *imdb.DB, ent imdb.EntityKind, id imdb.Atom) interface{} {
	switch ent {
	case imdb.EntityMovie:
		m, err := imdb.AtomToMovie(db, id)
		assert(err)
		return Movie{db, &m}
	case imdb.EntityTvshow:
		t, err := imdb.AtomToTvshow(db, id)
		assert(err)
		return Tvshow{db, &t}
	case imdb.EntityEpisode:
		e, err := imdb.AtomToEpisode(db, id)
		assert(err)
		return Episode{db, &e}
	}
	fatalf("Unrecognized entity type: %s", ent)
	panic("unreachable")
}

type Movie struct {
	db *imdb.DB
	*imdb.Movie
}

type Tvshow struct {
	db *imdb.DB
	*imdb.Tvshow
}

type Episode struct {
	db *imdb.DB
	*imdb.Episode
}

func (e Episode) Tvshow() Tvshow {
	tv, err := e.Episode.Tvshow(e.db)
	assert(err)
	return Tvshow{e.db, &tv}
}

func (e Tvshow) CountSeasons() (count int) {
	assert(csql.Safe(func() {
		count = csql.Count(e.db, `
			SELECT COUNT(*) AS count
			FROM (
				SELECT DISTINCT season
				FROM episode
				WHERE tvshow_atom_id = $1 AND season > 0
			) AS s
		`, e.Id)
	}))
	return
}

func (e Tvshow) CountEpisodes() (count int) {
	assert(csql.Safe(func() {
		count = csql.Count(e.db, `
			SELECT COUNT(*) AS count
			FROM episode
			WHERE tvshow_atom_id = $1 AND season > 0
		`, e.Id)
	}))
	return
}

func (e Movie) ReleaseDates() []imdb.ReleaseDate {
	return assertTwo(imdb.ReleaseDates(e.db, e)).([]imdb.ReleaseDate)
}

func (e Tvshow) ReleaseDates() []imdb.ReleaseDate {
	return assertTwo(imdb.ReleaseDates(e.db, e)).([]imdb.ReleaseDate)
}

func (e Episode) ReleaseDates() []imdb.ReleaseDate {
	return assertTwo(imdb.ReleaseDates(e.db, e)).([]imdb.ReleaseDate)
}

func (e Movie) RunningTimes() []imdb.RunningTime {
	return assertTwo(imdb.RunningTimes(e.db, e)).([]imdb.RunningTime)
}

func (e Tvshow) RunningTimes() []imdb.RunningTime {
	return assertTwo(imdb.RunningTimes(e.db, e)).([]imdb.RunningTime)
}

func (e Episode) RunningTimes() []imdb.RunningTime {
	return assertTwo(imdb.RunningTimes(e.db, e)).([]imdb.RunningTime)
}

func (e Movie) AkaTitles() []imdb.AkaTitle {
	return assertTwo(imdb.AkaTitles(e.db, e)).([]imdb.AkaTitle)
}

func (e Tvshow) AkaTitles() []imdb.AkaTitle {
	return assertTwo(imdb.AkaTitles(e.db, e)).([]imdb.AkaTitle)
}

func (e Episode) AkaTitles() []imdb.AkaTitle {
	return assertTwo(imdb.AkaTitles(e.db, e)).([]imdb.AkaTitle)
}

func (e Movie) AlternateVersions() []imdb.AlternateVersion {
	return assertTwo(imdb.AlternateVersions(e.db, e)).([]imdb.AlternateVersion)
}

func (e Tvshow) AlternateVersions() []imdb.AlternateVersion {
	return assertTwo(imdb.AlternateVersions(e.db, e)).([]imdb.AlternateVersion)
}

func (e Episode) AlternateVersions() []imdb.AlternateVersion {
	return assertTwo(imdb.AlternateVersions(e.db, e)).([]imdb.AlternateVersion)
}

func (e Movie) ColorInfos() []imdb.ColorInfo {
	return assertTwo(imdb.ColorInfos(e.db, e)).([]imdb.ColorInfo)
}

func (e Tvshow) ColorInfos() []imdb.ColorInfo {
	return assertTwo(imdb.ColorInfos(e.db, e)).([]imdb.ColorInfo)
}

func (e Episode) ColorInfos() []imdb.ColorInfo {
	return assertTwo(imdb.ColorInfos(e.db, e)).([]imdb.ColorInfo)
}

func (e Movie) MPAARating() imdb.RatingReason {
	return assertTwo(imdb.MPAARating(e.db, e)).(imdb.RatingReason)
}

func (e Tvshow) MPAARating() imdb.RatingReason {
	return assertTwo(imdb.MPAARating(e.db, e)).(imdb.RatingReason)
}

func (e Episode) MPAARating() imdb.RatingReason {
	return assertTwo(imdb.MPAARating(e.db, e)).(imdb.RatingReason)
}

func (e Movie) SoundMixes() []imdb.SoundMix {
	return assertTwo(imdb.SoundMixes(e.db, e)).([]imdb.SoundMix)
}

func (e Tvshow) SoundMixes() []imdb.SoundMix {
	return assertTwo(imdb.SoundMixes(e.db, e)).([]imdb.SoundMix)
}

func (e Episode) SoundMixes() []imdb.SoundMix {
	return assertTwo(imdb.SoundMixes(e.db, e)).([]imdb.SoundMix)
}

func (e Movie) Quotes() []imdb.Quote {
	return assertTwo(imdb.Quotes(e.db, e)).([]imdb.Quote)
}

func (e Tvshow) Quotes() []imdb.Quote {
	return assertTwo(imdb.Quotes(e.db, e)).([]imdb.Quote)
}

func (e Episode) Quotes() []imdb.Quote {
	return assertTwo(imdb.Quotes(e.db, e)).([]imdb.Quote)
}
