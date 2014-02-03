package imdb

import (
	"github.com/BurntSushi/csql"
)

type EntityKind int

const (
	EntityMovie EntityKind = iota
	EntityTvshow
	EntityEpisode
)

var Entities = map[string]EntityKind{
	"movie":   EntityMovie,
	"tvshow":  EntityTvshow,
	"episode": EntityEpisode,
}

func EntityKindFromString(e string) EntityKind {
	ent, ok := Entities[e]
	if !ok {
		fatalf("unrecognized entity %s", e)
	}
	return ent
}

func (e EntityKind) String() string {
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

type Entity interface {
	Ident() Atom
	Type() EntityKind
	Name() string
	Scan(rs csql.RowScanner) error
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

func (e Movie) Ident() Atom      { return e.Id }
func (e Movie) Type() EntityKind { return EntityMovie }
func (e Movie) Name() string     { return e.Title }
func (e Movie) String() string   { return sf("%s (%d)", e.Title, e.Year) }

func (e Tvshow) Ident() Atom      { return e.Id }
func (e Tvshow) Type() EntityKind { return EntityTvshow }
func (e Tvshow) Name() string     { return e.Title }
func (e Tvshow) String() string   { return sf("%s (%d)", e.Title, e.Year) }

func (e Episode) Ident() Atom      { return e.Id }
func (e Episode) Type() EntityKind { return EntityEpisode }
func (e Episode) Name() string     { return e.Title }
func (e Episode) String() string   { return sf("%s %d", e.Title, e.Year) }

func (e *Movie) Scan(rs csql.RowScanner) error {
	if e == nil {
		e = new(Movie)
	}
	return rs.Scan(&e.Id, &e.Title, &e.Year, &e.Sequence, &e.Tv, &e.Video)
}

func (e *Tvshow) Scan(rs csql.RowScanner) error {
	if e == nil {
		e = new(Tvshow)
	}
	return rs.Scan(&e.Id, &e.Title, &e.Year, &e.Sequence,
		&e.YearStart, &e.YearEnd)
}

func (e *Episode) Scan(rs csql.RowScanner) error {
	if e == nil {
		e = new(Episode)
	}
	return rs.Scan(&e.Id, &e.TvshowId, &e.Title,
		&e.Year, &e.Season, &e.EpisodeNum)
}

func AtomToMovie(db csql.Queryer, id Atom) (Movie, error) {
	e := new(Movie)
	err := e.Scan(db.QueryRow(`
		SELECT atom_id, title, year, sequence, tv, video
		FROM movie WHERE atom_id = $1`, id))
	return *e, err
}

func AtomToTvshow(db csql.Queryer, id Atom) (Tvshow, error) {
	e := new(Tvshow)
	err := e.Scan(db.QueryRow(`
		SELECT atom_id, title, year, sequence, year_start, year_end
		FROM tvshow WHERE atom_id = $1`, id))
	return *e, err
}

func AtomToEpisode(db csql.Queryer, id Atom) (Episode, error) {
	e := new(Episode)
	err := e.Scan(db.QueryRow(`
		SELECT atom_id, tvshow_atom_id, title, year, season, episode_num
		FROM episode WHERE atom_id = $1`, id))
	return *e, err
}

func (e Episode) Tvshow(db csql.Queryer) (Tvshow, error) {
	r := db.QueryRow(`
		SELECT atom_id, title, year, sequence, year_start, year_end
		FROM tvshow
		WHERE atom_id = $1`, e.TvshowId)
	tv := new(Tvshow)
	err := tv.Scan(r)
	return *tv, err
}
