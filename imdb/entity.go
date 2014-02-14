package imdb

import (
	"github.com/BurntSushi/csql"
)

type Atom int32

func (a Atom) String() string {
	return sf("%d", a)
}

type EntityKind int

const (
	EntityNone EntityKind = iota
	EntityMovie
	EntityTvshow
	EntityEpisode
	EntityActor
)

var Entities = map[string]EntityKind{
	"movie":   EntityMovie,
	"tvshow":  EntityTvshow,
	"episode": EntityEpisode,
	"actor":   EntityActor,
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
	case EntityActor:
		return "actor"
	}
	panic(sf("unrecognized entity %d", e))
}

type Entity interface {
	Ident() Atom
	Type() EntityKind
	Name() string
	Attrs(csql.Queryer, Attributer) error
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

type Actor struct {
	Id       Atom
	FullName string
	Sequence string
}

func (e *Movie) Ident() Atom      { return e.Id }
func (e *Movie) Type() EntityKind { return EntityMovie }
func (e *Movie) Name() string     { return e.Title }
func (e *Movie) String() string   { return sf("%s (%d)", e.Title, e.Year) }
func (e *Movie) Attrs(db csql.Queryer, attrs Attributer) error {
	return attrs.ForEntity(db, e)
}

func (e *Tvshow) Ident() Atom      { return e.Id }
func (e *Tvshow) Type() EntityKind { return EntityTvshow }
func (e *Tvshow) Name() string     { return e.Title }
func (e *Tvshow) String() string   { return sf("%s (%d)", e.Title, e.Year) }
func (e *Tvshow) Attrs(db csql.Queryer, attrs Attributer) error {
	return attrs.ForEntity(db, e)
}

func (e *Episode) Ident() Atom      { return e.Id }
func (e *Episode) Type() EntityKind { return EntityEpisode }
func (e *Episode) Name() string     { return e.Title }
func (e *Episode) String() string   { return sf("%s %d", e.Title, e.Year) }
func (e *Episode) Attrs(db csql.Queryer, attrs Attributer) error {
	return attrs.ForEntity(db, e)
}

func (e *Actor) Ident() Atom      { return e.Id }
func (e *Actor) Type() EntityKind { return EntityActor }
func (e *Actor) Name() string     { return e.FullName }
func (e *Actor) String() string   { return e.FullName }
func (e *Actor) Attrs(db csql.Queryer, attrs Attributer) error {
	return attrs.ForEntity(db, e)
}

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

func (e *Actor) Scan(rs csql.RowScanner) error {
	if e == nil {
		e = new(Actor)
	}
	return rs.Scan(&e.Id, &e.FullName, &e.Sequence)
}

func AtomToMovie(db csql.Queryer, id Atom) (*Movie, error) {
	e := new(Movie)
	err := e.Scan(db.QueryRow(`
		SELECT m.atom_id, n.name, m.year, m.sequence, m.tv, m.video
		FROM movie AS m
		LEFT JOIN name AS n ON n.atom_id = m.atom_id
		WHERE m.atom_id = $1
		`, id))
	return e, err
}

func AtomToTvshow(db csql.Queryer, id Atom) (*Tvshow, error) {
	e := new(Tvshow)
	err := e.Scan(db.QueryRow(`
		SELECT t.atom_id, n.name, t.year, t.sequence, t.year_start, t.year_end
		FROM tvshow AS t
		LEFT JOIN name AS n ON n.atom_id = t.atom_id
		WHERE t.atom_id = $1
		`, id))
	return e, err
}

func AtomToEpisode(db csql.Queryer, id Atom) (*Episode, error) {
	e := new(Episode)
	err := e.Scan(db.QueryRow(`
		SELECT e.atom_id, e.tvshow_atom_id, n.name,
			   e.year, e.season, e.episode_num
		FROM episode AS e
		LEFT JOIN name AS n ON n.atom_id = e.atom_id
		WHERE e.atom_id = $1
		`, id))
	return e, err
}

func AtomToActor(db csql.Queryer, id Atom) (*Actor, error) {
	e := new(Actor)
	err := e.Scan(db.QueryRow(`
		SELECT a.atom_id, n.name, a.sequence
		FROM actor AS a
		LEFT JOIN name AS n ON n.atom_id = a.atom_id
		WHERE a.atom_id = $1
		`, id))
	return e, err
}

func (e Episode) Tvshow(db csql.Queryer) (*Tvshow, error) {
	return AtomToTvshow(db, e.TvshowId)
}
