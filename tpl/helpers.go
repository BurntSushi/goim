package tpl

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"text/template"

	"github.com/kr/text"

	"github.com/BurntSushi/csql"

	"github.com/BurntSushi/goim/imdb"
)

var (
	sf = fmt.Sprintf
	ef = fmt.Errorf
)

// Functions corresponds to a map of functions that is available in every
// Goim template.
//
// The "lines" function takes a string and returns a list of strings
// corresponding to each line.
//
// The "wrap" function takes a column limit and a string and wraps the string
// to the given number of columns.
//
// The "underlined" function takes a string to repeat and a string to
// underline, and returns the string to underline, followed by a new line,
// followed by the string to repeat N times, where N is the length of the
// string to underline.
//
// The "count_seasons" function takes one parameter that is a TV show and
// returns the number of seasons that have aired.
//
// The "count_episodes" function takes one parameter that is a TV show and
// returns the number of episodes that have aired.
//
// The "tvshow" function takes one parameter that is an episode and returns
// its corresponding TV show.
//
// The list of functions starting with "running_times" retrieve attribute
// values given an entity. All functions accept one argument that must satisfy
// the imdb.Entity interface and return a list of attribute values.
var Functions = template.FuncMap{
	"lines":      lines,
	"wrap":       wrap,
	"underlined": underlined,

	"count_seasons":  countSeasons,
	"count_episodes": countEpisodes,
	"tvshow":         tvshow,

	"running_times":      attrGetter(new(imdb.RunningTimes)),
	"release_dates":      attrGetter(new(imdb.ReleaseDates)),
	"aka_titles":         attrGetter(new(imdb.AkaTitles)),
	"alternate_versions": attrGetter(new(imdb.AlternateVersions)),
	"color_info":         attrGetter(new(imdb.ColorInfos)),
	"mpaa":               attrGetter(new(imdb.RatingReason)),
	"sound_mixes":        attrGetter(new(imdb.SoundMixes)),
	"taglines":           attrGetter(new(imdb.Taglines)),
	"trivia":             attrGetter(new(imdb.Trivias)),
	"goofs":              attrGetter(new(imdb.Goofs)),
	"genres":             attrGetter(new(imdb.Genres)),
	"languages":          attrGetter(new(imdb.Languages)),
	"literature":         attrGetter(new(imdb.Literatures)),
	"locations":          attrGetter(new(imdb.Locations)),
	"links":              attrGetter(new(imdb.Links)),
	"plots":              attrGetter(new(imdb.Plots)),
	"quotes":             attrGetter(new(imdb.Quotes)),
	"rank":               attrGetter(new(imdb.UserRank)),
	"credits":            attrGetter(new(imdb.Credits)),

	"eq": func(a, b interface{}) bool { return a == b },
	"ne": func(a, b interface{}) bool { return a != b },

	"lt": func(a, b int) bool { return a < b },
	"le": func(a, b int) bool { return a <= b },
	"gt": func(a, b int) bool { return a > b },
	"ge": func(a, b int) bool { return a >= b },

	"ltf": func(a, b float64) bool { return a < b },
	"lef": func(a, b float64) bool { return a <= b },
	"gtf": func(a, b float64) bool { return a > b },
	"gef": func(a, b float64) bool { return a >= b },
}

// assert will quit Goim with the specified error if it is not nil.
func assert(err error) {
	if err != nil {
		panic(err)
	}
}

// assertDB makes sure there is a valid DB connection.
func assertDB() {
	if tplDB == nil {
		assert(ef("No database connection found. Please set one with SetDB."))
	}
}

// combine provides a way to compose values during template execution.
// This is particularly useful when executing sub-templates. For example,
// say you've defined two variables `$a` and `$b` that you want to pass to
// a sub-template. But templates can only take a single pipeline. Combine will
// let you bind any number of values. For example:
//
//	{{ template "tpl_name" (Combine "a" $a "b" $b) }}
//
// The template "tpl_name" can then access `$a` and `$b` with `.a` and `.b`.
//
// Note that the first and every other subsequent value must be strings. The
// second and every other subsequent value may be anything. There must be an
// even number of arguments given. If any part of this contract is violated,
// the function panics.
func combine(keyvals ...interface{}) map[string]interface{} {
	if len(keyvals)%2 != 0 {
		panic(sf("Combine must have even number of parameters but %d isn't.",
			len(keyvals)))
	}
	m := make(map[string]interface{})
	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			panic(sf("Parameter %d to Combine must be a string but it is "+
				"a %T.", i, keyvals[i]))
		}
		m[key] = keyvals[i+1]
	}
	return m
}

func wrap(limit int, s interface{}) string {
	return text.Wrap(sf("%s", s), limit)
}

func lines(s interface{}) []string {
	return strings.Split(sf("%s", s), "\n")
}

func underlined(rep string, is interface{}) string {
	s := sf("%s", is)
	return sf("%s\n%s", s, strings.Repeat(rep, len(s)))
}

func sorted(xs sort.Interface) interface{} {
	sort.Sort(xs)
	return xs
}

// attrGetter does some fancy reflection footwork to automatically build
// a function for any attribute retriever satisfying the imdb.Attributer
// interface.
func attrGetter(attrs imdb.Attributer) interface{} {
	// So we can make new attrs values.
	// Note that this is the *underlying* type of the imdb.Attributer.
	tattrs := reflect.TypeOf(attrs).Elem()

	return func(e imdb.Entity) interface{} {
		assertDB()
		vattrs := reflect.New(tattrs).Interface().(imdb.Attributer)
		assert(e.Attrs(tplDB, vattrs))
		return vattrs
	}
}

// countSeasons returns the number of seasons for the TV show given.
func countSeasons(e imdb.Entity) int {
	assertDB()
	q := `
		SELECT COUNT(*)
		FROM (
			SELECT season
			FROM episode
			WHERE tvshow_atom_id = $1 AND season > 0 AND episode_num > 0
			GROUP BY season
		) AS sub
	`
	return csql.Count(tplDB, q, e.Ident())
}

// countEpisodes returns the number of episodes for the TV show given.
func countEpisodes(e imdb.Entity) int {
	assertDB()
	q := `
		SELECT COUNT(*)
		FROM episode
		WHERE tvshow_atom_id = $1 AND season > 0 AND episode_num > 0
	`
	return csql.Count(tplDB, q, e.Ident())
}

// tvshow returns the TV show entity corresponding to the entity given.
func tvshow(e imdb.Entity) *imdb.Tvshow {
	assertDB()
	episode, ok := e.(*imdb.Episode)
	if !ok {
		panic(ef("'%s' is not an episode.", e))
	}
	tv, err := episode.Tvshow(tplDB)
	assert(err)
	return tv
}
