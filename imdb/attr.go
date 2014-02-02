package imdb

import (
	"time"

	"github.com/BurntSushi/csql"
)

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
	var times []RunningTime
	err := csql.Safe(func() {
		// IMDb claims that the "default" running time is one with a blank
		// country. This is nowhere near consistent, but we try anyway.
		// So we sort by country---the blank one will come first.
		// See: http://www.imdb.com/updates/guide/running_times
		rs := csql.Query(db, `
			SELECT country, minutes, attrs
			FROM running_time
			WHERE atom_id = $1 AND outlet = $2
			ORDER BY country ASC
		`, e.Ident(), e.Type().String())
		csql.Panic(csql.ForRow(rs, func(s csql.RowScanner) {
			var rt RunningTime
			csql.Panic(s.Scan(&rt.Country, &rt.Minutes, &rt.Attrs))
			times = append(times, rt)
		}))
	})
	return times, err
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
	var dates []ReleaseDate
	err := csql.Safe(func() {
		rs := csql.Query(db, `
			SELECT country, released, attrs
			FROM release_date
			WHERE atom_id = $1 AND outlet = $2
			ORDER BY released ASC
		`, e.Ident(), e.Type().String())
		csql.Panic(csql.ForRow(rs, func(s csql.RowScanner) {
			var d ReleaseDate
			csql.Panic(s.Scan(&d.Country, &d.Released, &d.Attrs))
			dates = append(dates, d)
		}))
	})
	return dates, err
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
	var titles []AkaTitle
	err := csql.Safe(func() {
		rs := csql.Query(db, `
			SELECT title, attrs
			FROM aka_title
			WHERE atom_id = $1 AND outlet = $2
			ORDER BY title ASC
		`, e.Ident(), e.Type().String())
		csql.Panic(csql.ForRow(rs, func(s csql.RowScanner) {
			var at AkaTitle
			csql.Panic(s.Scan(&at.Title, &at.Attrs))
			titles = append(titles, at)
		}))
	})
	return titles, err
}

type AlternateVersion string

func AlternateVersions(db csql.Queryer, e Entity) ([]AlternateVersion, error) {
	var alts []AlternateVersion
	err := csql.Safe(func() {
		rs := csql.Query(db, `
			SELECT about
			FROM alternate_version
			WHERE atom_id = $1 AND outlet = $2
		`, e.Ident(), e.Type().String())
		csql.Panic(csql.ForRow(rs, func(s csql.RowScanner) {
			var alt string
			csql.Panic(s.Scan(&alt))
			alts = append(alts, AlternateVersion(alt))
		}))
	})
	return alts, err
}
