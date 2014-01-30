package imdb

import (
	"time"

	"github.com/BurntSushi/csql"
)

type Entity int

const (
	EntityMovie Entity = iota
	EntityTvshow
	EntityEpisode
)

var Entities = map[string]Entity{
	"movie":   EntityMovie,
	"tvshow":  EntityTvshow,
	"episode": EntityEpisode,
}

func EntityFromString(e string) Entity {
	ent, ok := Entities[e]
	if !ok {
		fatalf("unrecognized entity %s", e)
	}
	return ent
}

func (e Entity) String() string {
	switch e {
	case EntityMovie:
		return "movie"
	case EntityTvshow:
		return "tvshow"
	case EntityEpisode:
		return "episode"
	}
	fatalf("unrecognized entity %d", e)
	panic("unreachable")
}

type Movie struct {
	Id       Atom
	Title    string
	Year     int
	Sequence string
	Tv       bool
	Video    bool
}

type Tvshow struct {
	Id                 Atom
	Title              string
	Year               int
	Sequence           string
	YearStart, YearEnd int
}

type Episode struct {
	Id                 Atom
	TvshowId           Atom
	Title              string
	Year               int
	Season, EpisodeNum int
}

func (m Movie) String() string {
	return sf("%s (%d)", m.Title, m.Year)
}

func (t Tvshow) String() string {
	return sf("%s (%d)", t.Title, t.Year)
}

func (e Episode) String() string {
	return sf("%s %d", e.Title, e.Year)
}

func ScanMovie(rs csql.RowScanner) (Movie, error) {
	m := Movie{}
	err := rs.Scan(&m.Id, &m.Title, &m.Year, &m.Sequence, &m.Tv, &m.Video)
	return m, err
}

func ScanTvshow(rs csql.RowScanner) (Tvshow, error) {
	t := Tvshow{}
	err := rs.Scan(&t.Id, &t.Title, &t.Year, &t.Sequence,
		&t.YearStart, &t.YearEnd)
	return t, err
}

func ScanEpisode(rs csql.RowScanner) (Episode, error) {
	e := Episode{}
	err := rs.Scan(&e.Id, &e.TvshowId, &e.Title,
		&e.Year, &e.Season, &e.EpisodeNum)
	return e, err
}

func AtomToMovie(db csql.Queryer, id Atom) (Movie, error) {
	return ScanMovie(db.QueryRow(`
		SELECT id, title, year, sequence, tv, video
		FROM movie WHERE id = $1`, id))
}

func AtomToTvshow(db csql.Queryer, id Atom) (Tvshow, error) {
	return ScanTvshow(db.QueryRow(`
		SELECT id, title, year, sequence, year_start, year_end
		FROM tvshow WHERE id = $1`, id))
}

func AtomToEpisode(db csql.Queryer, id Atom) (Episode, error) {
	return ScanEpisode(db.QueryRow(`
		SELECT id, tvshow_id, title, year, season, episode_num
		FROM episode WHERE id = $1`, id))
}

func (e Episode) Tvshow(db csql.Queryer) (tv Tvshow, err error) {
	r := db.QueryRow(`
		SELECT id, title, year, sequence, year_start, year_end
		FROM tvshow
		WHERE id = $1`, e.TvshowId)
	err = r.Scan(&tv)
	return
}

type ReleaseDate struct {
	Country  string
	Released time.Time
	Attrs    string
}

func (r ReleaseDate) String() string {
	var date string
	if !r.Released.IsZero() {
		date = r.Released.Format("2006-01-02")
	}
	var full string
	switch {
	case len(r.Country) > 0 && len(date) > 0:
		full = sf("%s:%s", r.Country, date)
	case len(r.Country) > 0:
		full = r.Country
	case len(date) > 0:
		full = date
	}
	if len(r.Attrs) > 0 {
		full += " " + r.Attrs
	}
	return full
}

func (m Movie) ReleaseDates(db csql.Queryer) ([]ReleaseDate, error) {
	return releaseDates(db, m.Id, EntityMovie)
}

func (t Tvshow) ReleaseDates(db csql.Queryer) ([]ReleaseDate, error) {
	return releaseDates(db, t.Id, EntityTvshow)
}

func (e Episode) ReleaseDates(db csql.Queryer) ([]ReleaseDate, error) {
	return releaseDates(db, e.Id, EntityEpisode)
}

func releaseDates(db csql.Queryer, id Atom, ent Entity) ([]ReleaseDate, error) {
	var dates []ReleaseDate
	err := csql.Safe(func() {
		rs := csql.Query(db, `
			SELECT country, released, attrs
			FROM release
			WHERE atom_id = $1 AND outlet = $2
			ORDER BY released ASC
		`, id, ent.String())
		csql.SQLPanic(csql.ForRow(rs, func(s csql.RowScanner) {
			var d ReleaseDate
			csql.SQLPanic(s.Scan(&d.Country, &d.Released, &d.Attrs))
			dates = append(dates, d)
		}))
	})
	return dates, err
}
