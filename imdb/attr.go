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
// unless the 'imdb_table' struct tag is set, in which case, that name is used.
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
		column := f.Tag.Get("imdb_table")
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
