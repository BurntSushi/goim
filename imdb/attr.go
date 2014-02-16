package imdb

import (
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/csql"
)

// Attributer describes types that correspond to one or more attribute values
// of an entity. Namely, values that satisfy this interface can load those
// attribute values from a database.
type Attributer interface {
	ForEntity(csql.Queryer, Entity) error
	Len() int
}

// attrs uses reflection to automatically construct a list of simple attribute
// rows from the database based on information in the attribute's struct.
// This includes building the SELECT query and the slice itself.
//
// zero MUST be a pointer to a simple struct. A simple struct MUST ONLY contain
// fields that can be encoded/decoded as declared by the 'database/sql'
// package. Column names are the lowercase version of their struct field name
// unless the 'imdb_name' struct tag is set, in which case, that name is used.
//
// extra is passed to the end of the query executed. Useful for specifying
// ORDER BY or LIMIT clauses.
func attrs(
	zero interface{},
	db csql.Queryer,
	e Entity,
	tableName string,
	idColumn string,
	extra string,
) (v interface{}, err error) {
	defer csql.Safe(&err)

	rz := reflect.ValueOf(zero).Elem()
	tz := rz.Type()
	nfields := tz.NumField()
	columns := make([]string, nfields)
	for i := 0; i < nfields; i++ {
		f := tz.Field(i)
		column := f.Tag.Get("imdb_name")
		if len(column) == 0 {
			column = strings.ToLower(f.Name)
		}
		columns[i] = column
	}
	tattrs := reflect.SliceOf(tz)
	vattrs := reflect.MakeSlice(tattrs, 0, 10)
	v = vattrs.Interface()

	q := sf("SELECT %s FROM %s WHERE %s = $1 %s",
		strings.Join(columns, ", "), tableName, idColumn, extra)
	rs := csql.Query(db, q, e.Ident())
	csql.Panic(csql.ForRow(rs, func(s csql.RowScanner) {
		loadCols := make([]interface{}, nfields)
		for i := 0; i < nfields; i++ {
			loadCols[i] = reflect.New(tz.Field(i).Type).Interface()
		}
		csql.Scan(s, loadCols...)

		row := reflect.New(tz).Elem()
		for i := 0; i < nfields; i++ {
			row.Field(i).Set(reflect.ValueOf(loadCols[i]).Elem())
		}
		vattrs = reflect.Append(vattrs, row)
	}))
	v = vattrs.Interface() // not sure if this is necessary.
	return
}

// RunningTime represents the running time of an entity in minutes. It may
// also include a country and some miscellaneous attributes.
// A given entity may have more than one running time because running times
// may differ depending upon the country they were released in.
// IMDb's data guides claim that more than one running time should only exist
// if there is a significant (> 5 minutes) difference, but in practice, this
// does not appear true.
type RunningTime struct {
	Country string
	Minutes int
	Attrs   string
}

func (r RunningTime) String() string {
	country, attrs := "", ""
	if len(r.Country) > 0 {
		country = sf(" (%s)", r.Country)
	}
	if len(r.Attrs) > 0 {
		attrs = sf(" %s", r.Attrs)
	}
	return sf("%d minutes%s%s", r.Minutes, country, attrs)
}

// RunningTimes corresponds to a list of running times, usually for one
// particular entity.
// *RunningTimes satisfies the Attributer interface.
type RunningTimes []RunningTime

func (as *RunningTimes) Len() int { return len(*as) }

// ForEntity fills 'as' with all running times corresponding to the entity
// given.
// Note that the list returned is ordered by country. As a result, the running
// time without a country comes first---which IMDb claims *should* be the
// default.
func (as *RunningTimes) ForEntity(db csql.Queryer, e Entity) error {
	rows, err := attrs(new(RunningTime), db, e, "running_time",
		"atom_id", "ORDER BY country ASC")
	*as = rows.([]RunningTime)
	return err
}

// ReleaseDate represents the date that a media item was released, along with
// the region and miscellaneous attributes.
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
		full = sf("%s (%s)", date, r.Country)
	case len(r.Country) > 0:
		full = r.Country
	case len(date) > 0:
		full = date
	}
	if len(r.Attrs) > 0 {
		full += sf(" %s", r.Attrs)
	}
	return full
}

// ReleaseDates corresponds to a list of release dates, usually for one
// particular entity.
// *ReleaseDates satisfies the Attributer interface.
type ReleaseDates []ReleaseDate

func (as *ReleaseDates) Len() int { return len(*as) }

// ForEntity fills 'as' with all release dates corresponding to the entity
// given.
// Note that the list returned is sorted by release date in ascending order.
func (as *ReleaseDates) ForEntity(db csql.Queryer, e Entity) error {
	rows, err := attrs(new(ReleaseDate), db, e, "release_date", "atom_id",
		"ORDER BY released")
	*as = rows.([]ReleaseDate)
	return err
}

// AkaTitle represents the alternative title of a media item with optional
// attributes.
type AkaTitle struct {
	Title string
	Attrs string
}

func (at AkaTitle) String() string {
	s := at.Title
	if len(at.Attrs) > 0 {
		s += " " + at.Attrs
	}
	return s
}

// AkaTitles corresponds to a list of AKA titles, usually for one particular
// entity.
// *AkaTitles satisfies the Attributer interface.
type AkaTitles []AkaTitle

func (as *AkaTitles) Len() int { return len(*as) }

// ForEntity fills 'as' with all AKA titles corresponding to the entity given.
// The list returned is sorted alphabetically in ascending order.
func (as *AkaTitles) ForEntity(db csql.Queryer, e Entity) error {
	rows, err := attrs(new(AkaTitle), db, e, "aka_title", "atom_id",
		"ORDER BY title")
	*as = rows.([]AkaTitle)
	return err
}

// AlternateVersion represents a description of an alternative version of
// an entity.
type AlternateVersion struct {
	About string
}

func (av AlternateVersion) String() string {
	return av.About
}

// AlternativeVersions corresponds to a list of alternative versions, usually
// for one particular entity.
// *AlternateVersions satisfies the Attributer interface.
type AlternateVersions []AlternateVersion

func (as *AlternateVersions) Len() int { return len(*as) }

// ForEntity fills 'as' with all alternative versions corresponding to the
// entity given.
func (as *AlternateVersions) ForEntity(db csql.Queryer, e Entity) error {
	rows, err := attrs(new(AlternateVersion), db, e, "alternate_version",
		"atom_id", "")
	*as = rows.([]AlternateVersion)
	return err
}

// ColorInfo represents the color information of media. Generally this
// indicates whether the film is in black and white or not, along with some
// miscellaneous attributes.
type ColorInfo struct {
	Color bool
	Attrs string
}

func (ci ColorInfo) String() string {
	s := "Black and White"
	if ci.Color {
		s = "Color"
	}
	if len(ci.Attrs) > 0 {
		s += " " + ci.Attrs
	}
	return s
}

// ColorInfos corresponds to a list of color information, usually for one
// particular entity.
// *ColorInfos satisfies the Attributer interface.
type ColorInfos []ColorInfo

func (as *ColorInfos) Len() int { return len(*as) }

// ForEntity fills 'as' with all color information corresponding to the entity
// given.
func (as *ColorInfos) ForEntity(db csql.Queryer, e Entity) error {
	rows, err := attrs(new(ColorInfo), db, e, "color_info", "atom_id", "")
	*as = rows.([]ColorInfo)
	return err
}

// RatingReason represents an MPAA standard rating and the reason for which
// that rating was given.
// *RatingReason satisfies the Attributer interface.
type RatingReason struct {
	Rating string
	Reason string
}

// Unrated returns true if and only if there is no MPAA rating.
func (mr RatingReason) Unrated() bool {
	return len(mr.Rating) == 0
}

func (mr RatingReason) String() string {
	if mr.Unrated() {
		return "Not rated"
	}
	reason := ""
	if len(mr.Reason) > 0 {
		reason = sf(" (%s)", mr.Reason)
	}
	return sf("Rated %s%s", mr.Rating, reason)
}

// Len is 0 if there is no rating or if it is unrated. Otherwise, the Len is 1.
func (mr *RatingReason) Len() int {
	if mr == nil || mr.Unrated() {
		return 0
	} else {
		return 1
	}
}

// ForEntity fills 'mr' with an MPAA rating if it exists. Otherwise, it remains
// nil.
func (mr *RatingReason) ForEntity(db csql.Queryer, e Entity) error {
	rows, err := attrs(new(RatingReason), db, e, "mpaa_rating", "atom_id",
		"LIMIT 1")
	reasons := rows.([]RatingReason)
	if len(reasons) > 0 {
		*mr = reasons[0]
	}
	return err
}

// SoundMix represents the type of sound mix used for a particular entity, like
// "Stereo" or "Dolby Digital". A sound mix may also have miscellaneous
// attributes.
type SoundMix struct {
	Mix   string
	Attrs string
}

func (sm SoundMix) String() string {
	s := sm.Mix
	if len(sm.Attrs) > 0 {
		s += " " + sm.Attrs
	}
	return s
}

// SoundMixes corresponds to a list of sound mixes, usually for one particular
// entity.
// *SoundMixes satisfies the Attributer interface.
type SoundMixes []SoundMix

func (as *SoundMixes) Len() int { return len(*as) }

// ForEntity fills 'as' with all sound mixes corresponding to the entity given.
func (as *SoundMixes) ForEntity(db csql.Queryer, e Entity) error {
	rows, err := attrs(new(SoundMix), db, e, "sound_mix", "atom_id", "")
	*as = rows.([]SoundMix)
	return err
}

// Tagline represents one tagline about an entity, which is usually a very
// short quip.
type Tagline struct {
	Tag string
}

func (t Tagline) String() string {
	return t.Tag
}

// Taglines corresponds to a list of taglines, usually for one particular
// entity.
// *Taglines satisfies the Attributer interface.
type Taglines []Tagline

func (as *Taglines) Len() int { return len(*as) }

// ForEntity fills 'as' with all taglines corresponding to the entity given.
func (as *Taglines) ForEntity(db csql.Queryer, e Entity) error {
	rows, err := attrs(new(Tagline), db, e, "tagline", "atom_id", "")
	*as = rows.([]Tagline)
	return err
}

// Trivia corresponds to a single piece of trivia about an entity. The text
// is guaranteed not to have any new lines.
type Trivia struct {
	Entry string
}

func (t Trivia) String() string {
	return t.Entry
}

// Trivias corresponds to a list of trivia, usually for one particular entity.
// *Trivias satisfies the Attributer interface.
type Trivias []Trivia

func (as *Trivias) Len() int { return len(*as) }

// ForEntity fills 'as' with all trivia corresponding to the entity given.
func (as *Trivias) ForEntity(db csql.Queryer, e Entity) error {
	rows, err := attrs(new(Trivia), db, e, "trivia", "atom_id", "")
	*as = rows.([]Trivia)
	return err
}

// Genre represents a single genre tag for an entity.
type Genre struct {
	Name string
}

func (g Genre) String() string {
	return g.Name
}

// Genres corresponds to a list of genre tags, usually for one particular
// entity.
// *Genres satisfies the Attributer interface.
type Genres []Genre

func (as *Genres) Len() int { return len(*as) }

// ForEntity fills 'as' with all genre tags correspondings to the entity given.
// Note that genres are sorted alphabetically in ascending order.
func (as *Genres) ForEntity(db csql.Queryer, e Entity) error {
	rows, err := attrs(new(Genre), db, e, "genre", "atom_id",
		"ORDER BY name ASC")
	*as = rows.([]Genre)
	return err
}

// Goof represents a single goof for an entity. There are several types of
// goofs, and each goof is labeled with a single type.
type Goof struct {
	Type  string `imdb_name:"goof_type"`
	Entry string
}

func (g Goof) String() string {
	return sf("(%s) %s", g.Type, g.Entry)
}

// Goofs corresponds to a list of goofs, usually for one particular entity.
// *Goofs satisfies the Attributer interface.
type Goofs []Goof

func (as *Goofs) Len() int { return len(*as) }

// ForEntity fills 'as' with all goofs corresponding to the entity given.
func (as *Goofs) ForEntity(db csql.Queryer, e Entity) error {
	rows, err := attrs(new(Goof), db, e, "goof", "atom_id", "")
	*as = rows.([]Goof)
	return err
}

// Language represents the language for a particular entity. Each language
// label may have miscellaneous attributes.
type Language struct {
	Name  string
	Attrs string
}

func (lang Language) String() string {
	s := lang.Name
	if len(lang.Attrs) > 0 {
		s += " " + lang.Attrs
	}
	return s
}

// Languages corresponds to a list of languages, usually for one particular
// entity.
// *Languages satisfies the Attributer interface.
type Languages []Language

func (as *Languages) Len() int { return len(*as) }

// ForEntity fills 'as' with all language labels corresponding to the entity
// given.
func (as *Languages) ForEntity(db csql.Queryer, e Entity) error {
	rows, err := attrs(new(Language), db, e, "language", "atom_id", "")
	*as = rows.([]Language)
	return err
}

// Literature represents a single written reference to an entity. There are
// different types of references, and each reference is tagged with a single
// type.
type Literature struct {
	Type string `imdb_name:"lit_type"`
	Ref  string
}

func (lit Literature) String() string {
	return sf("(%s) %s", lit.Type, lit.Ref)
}

// Literatures corresponds to a list of literature references, usually for one
// particular entity.
// *Literatures satisfies the Attributer interface.
type Literatures []Literature

func (as *Literatures) Len() int { return len(*as) }

// ForEntity fills 'as' with all literature references corresponding to the
// entity given.
func (as *Literatures) ForEntity(db csql.Queryer, e Entity) error {
	rows, err := attrs(new(Literature), db, e, "literature", "atom_id", "")
	*as = rows.([]Literature)
	return err
}

// Location represents a geographic location for a particular entity, usually
// corresponding to a filming location. Each location may have miscellaneous
// attributes.
type Location struct {
	Place string
	Attrs string
}

func (loc Location) String() string {
	s := loc.Place
	if len(loc.Attrs) > 0 {
		s += " " + loc.Attrs
	}
	return s
}

// Locations corresponds to a list of locations, usually for one particular
// entity.
// *Locations satisfies the Attributer interface.
type Locations []Location

func (as *Locations) Len() int { return len(*as) }

// ForEntity fills 'as' with all locations corresponding to the entity given.
func (as *Locations) ForEntity(db csql.Queryer, e Entity) error {
	rows, err := attrs(new(Location), db, e, "location", "atom_id", "")
	*as = rows.([]Location)
	return err
}

// Link represents a link between two entities of the same type. For example,
// they can describe movie prequels or sequels. Each link has a corresponding
// type (e.g., "followed by", "follows", ...) and the linked entity itself
type Link struct {
	Type   string
	Entity Entity
}

func (lk Link) String() string {
	return sf("%s %d (%s)", lk.Type, lk.Entity, lk.Entity.Type())
}

// Links corresponds to a list of connections between entities, usually
// originating from one particular entity.
// Links satisfies the sort.Interface interface.
// *Links satisfies the Attributer interface.
type Links []Link

func (as Links) Swap(i, j int) { as[i], as[j] = as[j], as[i] }
func (as Links) Less(i, j int) bool {
	iyear, jyear := as[i].Entity.EntityYear(), as[j].Entity.EntityYear()
	// move entity with a 0 year to bottom (usually indicates an entity that
	// is speculated to be released).
	if iyear == 0 {
		return false
	}
	if jyear == 0 {
		return true
	}
	return iyear < jyear
}

func (as *Links) Len() int { return len(*as) }

// ForEntity fills 'as' with all links corresponding to the entity given.
// The links returned are sorted by the year released, in ascending order.
func (as *Links) ForEntity(db csql.Queryer, e Entity) error {
	type link struct {
		Type   string `imdb_name:"link_type"`
		Id     Atom   `imdb_name:"link_atom_id"`
		Entity string
	}
	rows, err := attrs(new(link), db, e, "link", "atom_id", "")
	if err != nil {
		return err
	}

	// Blech, map entity strings to typed entity kinds...
	links := rows.([]link)
	typedLinks := make([]Link, len(links))
	for i := range links {
		kind := entityKindFromString(links[i].Entity)
		ent, err := FromAtom(db, kind, links[i].Id)
		if err != nil {
			return err
		}
		typedLinks[i] = Link{
			Type:   links[i].Type,
			Entity: ent,
		}
	}
	*as = typedLinks
	sort.Sort(as)
	return nil
}

// Plot represents the text of a plot summary---and it's author---for a movie,
// TV show or episode.
type Plot struct {
	Entry string
	By    string
}

func (p Plot) String() string { return sf("(%s) %s", p.By, p.Entry) }

// Plots corresponds to a list of plots, usually for one particular entity.
// *Plots satisfies the Attributer interface.
type Plots []Plot

func (as *Plots) Len() int { return len(*as) }

// ForEntity fills 'as' with all plots corresponding to the entity given.
func (as *Plots) ForEntity(db csql.Queryer, e Entity) error {
	rows, err := attrs(new(Plot), db, e, "plot", "atom_id", "")
	*as = rows.([]Plot)
	return err
}

// Quote represents the text of a quotation from an entity. Quotes are mostly
// freeform text, although the general format seems to be:
//
//	Character 1: Says something.
//		Which may continue to the next line, indented.
//	Character 2: Says something else.
//	...
type Quote struct {
	Entry string
}

func (q Quote) String() string { return q.Entry }

// Quotes corresponds to a list of quotes, usually for one particular entity.
// *Quotes satisfies the Attributer interface.
type Quotes []Quote

func (as *Quotes) Len() int { return len(*as) }

// ForEntity fills 'as' with all quotes corresponding to the entity given.
func (as *Quotes) ForEntity(db csql.Queryer, e Entity) error {
	rows, err := attrs(new(Quote), db, e, "quote", "atom_id", "")
	*as = rows.([]Quote)
	return err
}

// UserRank represents the rank and number votes by users of IMDb for a
// particular entity. If there are no votes, then the entity is considered
// unrated.
// *UserRank satisfies the Attributer interface.
type UserRank struct {
	Votes int
	Rank  int
}

// Unranked returns true if and only if this rank has no votes.
func (r UserRank) Unranked() bool {
	return r.Votes == 0
}

func (r UserRank) String() string {
	return sf("%d/100 (%d votes)", r.Rank, r.Votes)
}

// Len is 0 if there is no rank or if it is unrated. Otherwise, the Len is 1.
func (r *UserRank) Len() int {
	if r == nil || r.Unranked() {
		return 0
	} else {
		return 1
	}
}

// ForEntity fills 'r' with a user rank if it exists. Otherwise, it remains
// nil.
func (r *UserRank) ForEntity(db csql.Queryer, e Entity) error {
	rows, err := attrs(new(UserRank), db, e, "rating", "atom_id", "LIMIT 1")
	rates := rows.([]UserRank)
	if len(rates) > 0 {
		*r = rates[0]
	}
	return err
}

// Credit represents a movie and/or actor credit. It includes optional
// information like the character played and the billing position of the
// actor.
//
// Note that Credit has no corresponding type that satisfies the Attributer
// interface. This may change in the future.
type Credit struct {
	Actor     *Actor
	Media     Entity
	Character string
	Position  int
	Attrs     string
}

// Valid returns true if and only if this credit belong to a valid movie
// and a valid actor.
func (c Credit) Valid() bool {
	return c.Actor != nil && c.Media != nil
}

// String only shows the character/position/attrs of the credit.
func (c Credit) String() string {
	var s string
	if len(c.Character) > 0 {
		s = sf("[%s]", c.Character)
	} else {
		s = "[unknown]"
	}
	if c.Position > 0 {
		s += sf(" <%d>", c.Position)
	}
	if len(c.Attrs) > 0 {
		s += " " + c.Attrs
	}
	return s
}

// Credits corresponds to a list of credits, usually for one particular
// movie/episode or for one particular actor.
// *Credits satisfies the Attributer interface.
type Credits []Credit

func (as *Credits) Len() int     { return len(*as) }
func (as Credits) Swap(i, j int) { as[i], as[j] = as[j], as[i] }

type actorCredits struct {
	*Credits
}

func (asp actorCredits) Less(i, j int) bool {
	as := *asp.Credits
	iyear, jyear := as[i].Media.EntityYear(), as[j].Media.EntityYear()
	if iyear != jyear {
		// Any entity with a year should come before all entities without
		// years.
		switch {
		case iyear > 0 && jyear > 0:
			return iyear > jyear // descending!
		case iyear > 0:
			return true
		case jyear > 0:
			return false
		}
	}
	iname, jname := as[i].Media.Name(), as[j].Media.Name()
	return iname < jname // back to ascending
}

type mediaCredits struct {
	*Credits
}

func (asp mediaCredits) Less(i, j int) bool {
	as := *asp.Credits
	ibill, jbill := as[i].Position, as[j].Position
	if ibill != jbill {
		// Any credit without a billing position should come after all
		// credits with a billing position.
		switch {
		case ibill > 0 && jbill > 0:
			return ibill < jbill
		case ibill > 0:
			return true
		case jbill > 0:
			return false
		}
	}
	return as[i].Actor.FullName < as[j].Actor.FullName
}

// ForEntity fills 'r' with all credits for the given entity. If the entity is
// a movie or episode, then it returns all available cast sorted by
// billing position and then alphabetically by full name, both in ascending
// order. If the entity is a cast member, then it returns all movies
// and episodes that the cast member appeared in, sorted by year of release in
// descending order and then alphabetically in ascending order.
func (r *Credits) ForEntity(db csql.Queryer, e Entity) error {
	type credit struct {
		ActorId   Atom `imdb_name:"actor_atom_id"`
		MediaId   Atom `imdb_name:"media_atom_id"`
		Character string
		Position  int
		Attrs     string
	}

	var idColumn string
	_, isActor := e.(*Actor)
	if isActor {
		idColumn = "actor_atom_id"
	} else {
		idColumn = "media_atom_id"
	}

	rows, err := attrs(new(credit), db, e, "credit", idColumn, "")
	if err != nil {
		return err
	}

	credits := rows.([]credit)
	typedCredits := make([]Credit, len(credits))
	for i, c := range credits {
		if isActor {
			med, err := fromAtomGuess(db, c.MediaId)
			if err != nil {
				return err
			}
			typedCredits[i] = Credit{
				Actor:     e.(*Actor),
				Media:     med,
				Character: c.Character,
				Position:  c.Position,
				Attrs:     c.Attrs,
			}
		} else {
			act, err := FromAtom(db, EntityActor, c.ActorId)
			if err != nil {
				return err
			}
			typedCredits[i] = Credit{
				Actor:     act.(*Actor),
				Media:     e,
				Character: c.Character,
				Position:  c.Position,
				Attrs:     c.Attrs,
			}
		}
	}
	*r = typedCredits
	if isActor {
		sort.Sort(actorCredits{r})
	} else {
		sort.Sort(mediaCredits{r})
	}
	return err
}
