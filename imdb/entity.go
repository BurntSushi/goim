package imdb

import (
	"github.com/BurntSushi/csql"
)

// Atom corresponds to a unique identifier for an entity. If any two entities
// have different atoms, then they are considered logically distinct.
type Atom int32

func (a Atom) String() string {
	return sf("%d", a)
}

// EntityKind represents all possible types of entities supported by this
// package.
type EntityKind int

// All possible entities.
const (
	EntityMovie EntityKind = iota
	EntityTvshow
	EntityEpisode
	EntityActor
)

// Entities is a map from a string representation of an entity type to a Goim
// entity type.
var Entities = map[string]EntityKind{
	"movie":   EntityMovie,
	"tvshow":  EntityTvshow,
	"episode": EntityEpisode,
	"actor":   EntityActor,
}

func entityKindFromString(e string) EntityKind {
	ent, ok := Entities[e]
	if !ok {
		panic(sf("BUG: unrecognized entity %s", e))
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

// Entity is an interface that all types claiming to be an entity must satisfy.
type Entity interface {
	// Returns a unique atom identifier for this entity.
	Ident() Atom

	// The type of this entity.
	Type() EntityKind

	// A name representing this entity. It need not be unique among all
	// entities.
	Name() string

	// Returns the year associated with this entity. If no such year exists
	// or is not relevant, it may be 0.
	EntityYear() int

	// Attrs uses double dispatch to load all attribute values for the given
	// Attributer for this entity.
	Attrs(csql.Queryer, Attributer) error

	// Scan loads an entity from a row in the database.
	Scan(rs csql.RowScanner) error
}

// FromAtom returns an entity given its type and its unique identifier.
func FromAtom(db csql.Queryer, ent EntityKind, id Atom) (Entity, error) {
	switch ent {
	case EntityMovie:
		return atomToMovie(db, id)
	case EntityTvshow:
		return atomToTvshow(db, id)
	case EntityEpisode:
		return atomToEpisode(db, id)
	case EntityActor:
		return atomToActor(db, id)
	}
	return nil, ef("Unrecognized entity type: %s", ent)
}

// fromAtomGuess is just like FromAtom, except it doesn't use an entity type
// as a hint for which table to select from. Therefore, it tries all entity
// types until it gets a hit. If no entities could be found matching the
// identifier given, an error is returned.
func fromAtomGuess(db csql.Queryer, id Atom) (e Entity, err error) {
	e, err = atomToMovie(db, id)
	if err == nil {
		return e, nil
	}
	e, err = atomToTvshow(db, id)
	if err == nil {
		return e, nil
	}
	e, err = atomToEpisode(db, id)
	if err == nil {
		return e, nil
	}
	e, err = atomToActor(db, id)
	if err == nil {
		return e, nil
	}
	return nil, ef("Could not find any entity corresponding to atom %d", id)
}

// Movie represents a single movie in IMDb. This includes "made for tv" and
// "made for video" movies.
type Movie struct {
	Id       Atom
	Title    string
	Year     int    // Year released.
	Sequence string // Non-data. Used by IMDb for unique entity strings.
	Tv       bool
	Video    bool
}

// Tvshow represents a single TV show in IMDb. Typically TV shows lack
// attribute data in lieu of individual episodes containing the data, and are 
// instead a way of connecting episodes together.
type Tvshow struct {
	Id        Atom
	Title     string
	Year      int    // Year started.
	Sequence  string // Non-data. Used by IMDb for unique entity strings.
	YearStart int
	YearEnd   int // Year ended or 0 if still on air.
}

// Episode represents a single episode for a single TV show in IMDb.
type Episode struct {
	Id                 Atom
	TvshowId           Atom
	Title              string
	Year               int
	Season, EpisodeNum int // May be 0!
}

// Actor represents a single cast member that has appeared in the credits of
// at least one movie, TV show or episode in IMDb.
type Actor struct {
	Id       Atom
	FullName string
	Sequence  string // Non-data. Used by IMDb for unique entity strings.
}

func entityString(title string, year int) string {
	var s string
	if len(title) > 0 {
		s = title
	} else {
		s = "N/A"
	}
	if year > 0 {
		s += sf(" (%d)", year)
	}
	return s
}

func (e *Movie) Ident() Atom      { return e.Id }
func (e *Movie) Type() EntityKind { return EntityMovie }
func (e *Movie) Name() string     { return e.Title }
func (e *Movie) EntityYear() int  { return e.Year }
func (e *Movie) String() string   { return entityString(e.Title, e.Year) }
func (e *Movie) Attrs(db csql.Queryer, attrs Attributer) error {
	return attrs.ForEntity(db, e)
}

func (e *Tvshow) Ident() Atom      { return e.Id }
func (e *Tvshow) Type() EntityKind { return EntityTvshow }
func (e *Tvshow) Name() string     { return e.Title }
func (e *Tvshow) EntityYear() int  { return e.Year }
func (e *Tvshow) String() string   { return entityString(e.Title, e.Year) }
func (e *Tvshow) Attrs(db csql.Queryer, attrs Attributer) error {
	return attrs.ForEntity(db, e)
}

func (e *Episode) Ident() Atom      { return e.Id }
func (e *Episode) Type() EntityKind { return EntityEpisode }
func (e *Episode) Name() string     { return e.Title }
func (e *Episode) EntityYear() int  { return e.Year }
func (e *Episode) String() string   { return entityString(e.Title, e.Year) }
func (e *Episode) Attrs(db csql.Queryer, attrs Attributer) error {
	return attrs.ForEntity(db, e)
}

func (e *Actor) Ident() Atom      { return e.Id }
func (e *Actor) Type() EntityKind { return EntityActor }
func (e *Actor) Name() string     { return e.FullName }
func (e *Actor) EntityYear() int  { return 0 }
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

func atomToMovie(db csql.Queryer, id Atom) (*Movie, error) {
	e := new(Movie)
	err := e.Scan(db.QueryRow(`
		SELECT m.atom_id, n.name, m.year, m.sequence, m.tv, m.video
		FROM movie AS m
		LEFT JOIN name AS n ON n.atom_id = m.atom_id
		WHERE m.atom_id = $1
		`, id))
	return e, err
}

func atomToTvshow(db csql.Queryer, id Atom) (*Tvshow, error) {
	e := new(Tvshow)
	err := e.Scan(db.QueryRow(`
		SELECT t.atom_id, n.name, t.year, t.sequence, t.year_start, t.year_end
		FROM tvshow AS t
		LEFT JOIN name AS n ON n.atom_id = t.atom_id
		WHERE t.atom_id = $1
		`, id))
	return e, err
}

func atomToEpisode(db csql.Queryer, id Atom) (*Episode, error) {
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

func atomToActor(db csql.Queryer, id Atom) (*Actor, error) {
	e := new(Actor)
	err := e.Scan(db.QueryRow(`
		SELECT a.atom_id, n.name, a.sequence
		FROM actor AS a
		LEFT JOIN name AS n ON n.atom_id = a.atom_id
		WHERE a.atom_id = $1
		`, id))
	return e, err
}

// Tvshow returns a TV show entity that corresponds to this episode.
func (e *Episode) Tvshow(db csql.Queryer) (*Tvshow, error) {
	return atomToTvshow(db, e.TvshowId)
}
