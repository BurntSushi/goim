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
			WHERE atom_id = $1
			ORDER BY country ASC
		`, e.Ident())
		csql.Panic(csql.ForRow(rs, func(s csql.RowScanner) {
			var rt RunningTime
			csql.Scan(s, &rt.Country, &rt.Minutes, &rt.Attrs)
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
			WHERE atom_id = $1
			ORDER BY released ASC
		`, e.Ident())
		csql.Panic(csql.ForRow(rs, func(s csql.RowScanner) {
			var d ReleaseDate
			csql.Scan(s, &d.Country, &d.Released, &d.Attrs)
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
			WHERE atom_id = $1
			ORDER BY title ASC
		`, e.Ident())
		csql.Panic(csql.ForRow(rs, func(s csql.RowScanner) {
			var at AkaTitle
			csql.Scan(s, &at.Title, &at.Attrs)
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
			WHERE atom_id = $1
		`, e.Ident())
		csql.Panic(csql.ForRow(rs, func(s csql.RowScanner) {
			var alt string
			csql.Scan(s, &alt)
			alts = append(alts, AlternateVersion(alt))
		}))
	})
	return alts, err
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
	var infos []ColorInfo
	err := csql.Safe(func() {
		rs := csql.Query(db, `
			SELECT color, attrs
			FROM color_info
			WHERE atom_id = $1
		`, e.Ident())
		csql.Panic(csql.ForRow(rs, func(s csql.RowScanner) {
			var info ColorInfo
			csql.Scan(s, &info.Color, &info.Attrs)
			infos = append(infos, info)
		}))
	})
	return infos, err
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
	var rating RatingReason
	err := csql.Safe(func() {
		rs := csql.Query(db, `
			SELECT rating, reason
			FROM mpaa_rating
			WHERE atom_id = $1
			LIMIT 1
		`, e.Ident())
		csql.Panic(csql.ForRow(rs, func(s csql.RowScanner) {
			csql.Scan(s, &rating.Rating, &rating.Reason)
		}))
	})
	return rating, err
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
	var mixes []SoundMix
	err := csql.Safe(func() {
		rs := csql.Query(db, `
			SELECT mix, attrs
			FROM sound_mix
			WHERE atom_id = $1
		`, e.Ident())
		csql.Panic(csql.ForRow(rs, func(s csql.RowScanner) {
			var mix SoundMix
			csql.Scan(s, &mix.Mix, &mix.Attrs)
			mixes = append(mixes, mix)
		}))
	})
	return mixes, err
}
