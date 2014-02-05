package imdb

import (
	"reflect"
	"strings"
	"time"

	"github.com/BurntSushi/csql"
)

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
	extra string,
) (interface{}, error) {
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

	err := csql.Safe(func() {
		q := sf("SELECT %s FROM %s WHERE atom_id = $1 %s",
			strings.Join(columns, ", "), tableName, extra)
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
	})
	return vattrs.Interface(), err
}

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

func RunningTimes(db csql.Queryer, e Entity) ([]RunningTime, error) {
	rows, err := attrs(new(RunningTime), db, e,
		"running_time", "ORDER BY country ASC")
	return rows.([]RunningTime), err
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

func ReleaseDates(db csql.Queryer, e Entity) ([]ReleaseDate, error) {
	rows, err := attrs(new(ReleaseDate), db, e, "release_date",
		"ORDER BY released")
	return rows.([]ReleaseDate), err
}

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

func AkaTitles(db csql.Queryer, e Entity) ([]AkaTitle, error) {
	rows, err := attrs(new(AkaTitle), db, e, "aka_title", "ORDER BY title")
	return rows.([]AkaTitle), err
}

type AlternateVersion struct {
	About string
}

func (av AlternateVersion) String() string {
	return av.About
}

func AlternateVersions(db csql.Queryer, e Entity) ([]AlternateVersion, error) {
	rows, err := attrs(new(AlternateVersion), db, e, "alternate_version", "")
	return rows.([]AlternateVersion), err
}

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

func ColorInfos(db csql.Queryer, e Entity) ([]ColorInfo, error) {
	rows, err := attrs(new(ColorInfo), db, e, "color_info", "")
	return rows.([]ColorInfo), err
}

type RatingReason struct {
	Rating string
	Reason string
}

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

func MPAARating(db csql.Queryer, e Entity) (RatingReason, error) {
	rows, err := attrs(new(RatingReason), db, e, "mpaa_rating", "LIMIT 1")
	reasons := rows.([]RatingReason)
	if len(reasons) == 0 {
		return RatingReason{}, err
	}
	return reasons[0], err
}

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

func SoundMixes(db csql.Queryer, e Entity) ([]SoundMix, error) {
	rows, err := attrs(new(SoundMix), db, e, "sound_mix", "")
	return rows.([]SoundMix), err
}

type Tagline struct {
	Tag string
}

func (t Tagline) String() string {
	return t.Tag
}

func Taglines(db csql.Queryer, e Entity) ([]Tagline, error) {
	rows, err := attrs(new(Tagline), db, e, "tagline", "")
	return rows.([]Tagline), err
}

type Trivia struct {
	Entry string
}

func (t Trivia) String() string {
	return t.Entry
}

func Trivias(db csql.Queryer, e Entity) ([]Trivia, error) {
	rows, err := attrs(new(Trivia), db, e, "trivia", "")
	return rows.([]Trivia), err
}

type Genre struct {
	Name string
}

func (g Genre) String() string {
	return g.Name
}

func Genres(db csql.Queryer, e Entity) ([]Genre, error) {
	rows, err := attrs(new(Genre), db, e, "genre", "ORDER BY name ASC")
	return rows.([]Genre), err
}

type Goof struct {
	Type  string `imdb_name:"goof_type"`
	Entry string
}

func (g Goof) String() string {
	return sf("(%s) %s", g.Type, g.Entry)
}

func Goofs(db csql.Queryer, e Entity) ([]Goof, error) {
	rows, err := attrs(new(Goof), db, e, "goof", "")
	return rows.([]Goof), err
}

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

func Languages(db csql.Queryer, e Entity) ([]Language, error) {
	rows, err := attrs(new(Language), db, e, "language", "")
	return rows.([]Language), err
}

type Literature struct {
	Type string `imdb_name:"lit_type"`
	Ref  string
}

func (lit Literature) String() string {
	return sf("(%s) %s", lit.Type, lit.Ref)
}

func Literatures(db csql.Queryer, e Entity) ([]Literature, error) {
	rows, err := attrs(new(Literature), db, e, "literature", "")
	return rows.([]Literature), err
}

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

func Locations(db csql.Queryer, e Entity) ([]Location, error) {
	rows, err := attrs(new(Location), db, e, "location", "")
	return rows.([]Location), err
}

type Link struct {
	Type   string `imdb_name:"link_type"`
	Id     Atom   `imdb_name:"link_atom_id"`
	Entity string
}

func (lk Link) String() string {
	return sf("%s %d (%s)", lk.Type, lk.Id, lk.Entity)
}

func Links(db csql.Queryer, e Entity) ([]Link, error) {
	rows, err := attrs(new(Link), db, e, "link", "")
	return rows.([]Link), err
}

type Plot struct {
	Entry string
	By    string
}

func (p Plot) String() string {
	return sf("(%s) %s", p.By, p.Entry)
}

func Plots(db csql.Queryer, e Entity) ([]Plot, error) {
	rows, err := attrs(new(Plot), db, e, "plot", "")
	return rows.([]Plot), err
}

type Quote struct {
	Entry string
}

func (q Quote) String() string {
	return q.Entry
}

func Quotes(db csql.Queryer, e Entity) ([]Quote, error) {
	rows, err := attrs(new(Quote), db, e, "quote", "")
	return rows.([]Quote), err
}

type UserRating struct {
	Votes int
	Rank  int
}

func (r UserRating) Unrated() bool {
	return r.Votes == 0
}

func (r UserRating) String() string {
	return sf("%d/100 (%d votes)", r.Rank, r.Votes)
}

func Rating(db csql.Queryer, e Entity) (UserRating, error) {
	rows, err := attrs(new(UserRating), db, e, "rating", "LIMIT 1")
	rates := rows.([]UserRating)
	if len(rates) == 0 {
		return UserRating{}, err
	}
	return rates[0], err
}
