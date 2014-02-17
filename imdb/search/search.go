package search

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/csql"

	"github.com/BurntSushi/goim/imdb"
)

const (
	maxYear    = 3000
	maxRank    = 100
	maxVotes   = 1000000000
	maxBilled  = 1000000
	maxSeason  = 1000000
	maxEpisode = 1000000
)

var (
	sf  = fmt.Sprintf
	ef  = fmt.Errorf
	pef = func(f string, v ...interface{}) {
		fmt.Fprintf(os.Stderr, f, v...)
	}
)

// Result represents the data returned for each result of a search.
type Result struct {
	Entity imdb.EntityKind
	Id     imdb.Atom
	Name   string
	Year   int

	// Arbitrary additional data specific to an entity.
	// e.g., Whether a movie is straight to video or made for TV.
	// e.g., The season and episode number of a TV episode.
	Attrs string

	// Similarity corresponds to the amount of similarity between the name
	// given in the query and the name returned in this result.
	// This is set to -1 when fuzzy searching is not available (e.g., for
	// SQLite or Postgres when the 'pg_trgm' extension isn't enabled).
	Similarity float64

	// If a rating exists for a search result, it will be stored here.
	Rank imdb.UserRank

	// If the search accesses credit information, then it will be stored here.
	Credit Credit
}

// Credit represents the credit information available in a search result.
// This is distinct from the normal imdb.Credit type since it stores atom
// identifiers instead of the entities themselves.
type Credit struct {
	ActorId   imdb.Atom
	MediaId   imdb.Atom
	Character string
	Position  int
	Attrs     string
}

// Valid returns true if and only if this credit belong to a valid movie
// and a valid actor.
func (c Credit) Valid() bool {
	return c.ActorId > 0 && c.MediaId > 0
}

func (sr Result) GetEntity(db csql.Queryer) (imdb.Entity, error) {
	return imdb.FromAtom(db, sr.Entity, sr.Id)
}

// Searcher represents the parameters of a search.
type Searcher struct {
	db            *imdb.DB
	fuzzy         bool
	name          string
	what          string
	debug         bool
	atom          imdb.Atom
	entities      []imdb.EntityKind
	order         []searchOrder
	limit         int
	goodThreshold float64
	chooser       Chooser

	subTvshow, subCredits, subCast                *subsearch
	year, rating, votes, season, episode, billing *irange

	noTvMovie, noVideoMovie bool
}

// Chooser corresponds to a function called by the searcher in this
// package to resolve ambiguous query parameters. For example, if a TV show
// is specified with '{tvshow:supernatural}' and there is more than one good
// hit, then the chooser function will be called.
//
// If the search result returned is nil and the error is nil, then the
// search will return no results without error.
//
// If an error is returned, then the search stops and the error is passed to
// the caller of Searcher.Results.
//
// If no chooser function is supplied, then the first search result is always
// picked. If there are no results, then the query stops and returns no
// results.
//
// The string provided to the chooser function is a short noun phrase that
// represents the thing being searched. (e.g., "TV show".)
type Chooser func([]Result, string) (*Result, error)

type searchOrder struct {
	column, order string
}

type irange struct {
	min, max int
}

type subsearch struct {
	*Searcher
	id imdb.Atom
}

func New(db *imdb.DB, query string) (*Searcher, error) {
	s := &Searcher{
		db:            db,
		fuzzy:         db.IsFuzzyEnabled(),
		limit:         30,
		goodThreshold: 0.25,
		what:          "entity",
	}

	var qname []string
	var err error
	for _, arg := range queryTokens(query) {
		qname, err = s.addToken(qname, arg)
		if err != nil {
			return nil, err
		}
	}
	s.name = strings.Join(qname, " ")

	// Disable similarity scores if a wildcard is used.
	if strings.ContainsAny(s.name, "%_") {
		s.fuzzy = false
	}
	return s, nil
}

func (s *Searcher) addToken(queryName []string, arg string) ([]string, error) {
	name, val := argOption(arg)
	if cmd, ok := allCommands[name]; ok {
		if cmd.hasArg && len(val) == 0 {
			return nil, ef("The %s command requires an argument.", name)
		} else if !cmd.hasArg && len(val) > 0 {
			return nil, ef("The %s command does not have an argument.", name)
		}
		return queryName, cmd.add(s, val)
	} else {
		if len(name) > 0 {
			return nil, ef("Unrecognized search option: %s", name)
		}
		return append(queryName, arg), nil
	}
}

func (s *Searcher) subSearcher(name, query string) (*Searcher, error) {
	if len(query) == 0 {
		return nil, ef("No query found for '%s'.", name)
	}
	sub, err := New(s.db, query)
	if err != nil {
		return nil, ef("Error with sub-search for %s: %s", name, err)
	}
	return sub, nil
}

// Results executes the parameters of the search and returns the results.
func (s *Searcher) Results() (rs []Result, err error) {
	defer csql.Safe(&err)

	if s.subTvshow != nil {
		if err := s.subTvshow.choose(s, s.chooser); err != nil {
			return nil, err
		}
	}
	if s.subCredits != nil {
		if err := s.subCredits.choose(s, s.chooser); err != nil {
			return nil, err
		}
	}
	if s.subCast != nil {
		if err := s.subCast.choose(s, s.chooser); err != nil {
			return nil, err
		}
	}

	var rows *sql.Rows
	if len(s.name) == 0 {
		rows = csql.Query(s.db, s.sql())
	} else {
		rows = csql.Query(s.db, s.sql(), s.name)
	}
	csql.Panic(csql.ForRow(rows, func(scanner csql.RowScanner) {
		var r Result
		var ent string
		csql.Scan(scanner, &ent, &r.Id, &r.Name, &r.Year,
			&r.Similarity, &r.Attrs,
			&r.Rank.Votes, &r.Rank.Rank,
			&r.Credit.ActorId, &r.Credit.MediaId, &r.Credit.Character,
			&r.Credit.Position, &r.Credit.Attrs)
		r.Entity = imdb.Entities[ent]
		rs = append(rs, r)
	}))
	return
}

func (s *Searcher) Pick(rs []Result) (*Result, error) {
	if len(rs) == 0 {
		return nil, nil
	} else if len(rs) == 1 {
		return &rs[0], nil
	} else if ft, sd := rs[0].Similarity, rs[1].Similarity; ft > -1 && sd > -1 {
		if ft-sd >= s.goodThreshold {
			return &rs[0], nil
		}
	}
	if s.chooser == nil {
		return &rs[0], nil
	}
	r, err := s.chooser(rs, s.what)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return r, nil
}

func (sub *subsearch) choose(parent *Searcher, chooser Chooser) error {
	sub.goodThreshold = parent.goodThreshold
	sub.chooser = parent.chooser
	sub.debug = parent.debug

	rs, err := sub.Results()
	if err != nil {
		return ef("Error with %s sub-search: %s", sub.what, err)
	}
	r, err := sub.Pick(rs)
	if err != nil {
		return ef("Error picking %s result: %s", sub.what, err)
	}
	if r == nil {
		sub.id = -1 // force search to fail.
		return nil
	}
	sub.id = r.Id
	return nil
}

func (sub *subsearch) empty() bool {
	return sub == nil || sub.id == 0
}

// GoodThreshold sets the threshold at which a result is considered "good"
// relative to other results returned. This is used to automatically pick a
// good hit from sub-searches (like for a TV show). Namely, if the difference
// in similarity between the first and second hits is greater than or equal to
// the threshold given, then the first hit is automatically returned.
//
// By default, the threshold is set to 0.25.
//
// Set the threshold to 1.0 to disable this behavior.
func (s *Searcher) GoodThreshold(diff float64) *Searcher {
	s.goodThreshold = diff
	return s
}

// Entity adds the given entity to the search. Results only belonging to the
// entities in the search will be returned. This function may called more than
// once to specify additional entities to allow.
func (s *Searcher) Entity(e imdb.EntityKind) *Searcher {
	s.entities = append(s.entities, e)
	return s
}

// Atom specifies that the result returned must have the atom identifier
// given. Note that this guarantees that the number of results will either
// be 0 or 1.
func (s *Searcher) Atom(id imdb.Atom) *Searcher {
	s.atom = id
	return s
}

// Years specifies that the results must be in the range of years given.
// The range is inclusive.
func (s *Searcher) Years(min, max int) *Searcher {
	s.year = &irange{min, max}
	return s
}

// Seasons specifies that the results must be in the range of seasons given.
// The range is inclusive.
func (s *Searcher) Seasons(min, max int) *Searcher {
	s.season = &irange{min, max}
	return s
}

// Episodes specifies that the results must be in the range of episodes given.
// The range is inclusive.
func (s *Searcher) Episodes(min, max int) *Searcher {
	s.episode = &irange{min, max}
	return s
}

// NoTvMovies filters out "made for TV" movies from a search.
func (s *Searcher) NoTvMovies() *Searcher {
	s.noTvMovie = true
	return s
}

// NoVideoMovies filters out "made for video" movies from a search.
func (s *Searcher) NoVideoMovies() *Searcher {
	s.noVideoMovie = true
	return s
}

// Ranks specifies that the results must be in the range of ranks given.
// The range is inclusive.
// Note that the minimum rank is 0 and the maximum is 100.
func (s *Searcher) Ranks(min, max int) *Searcher {
	s.rating = &irange{min, max}
	return s
}

// Votes specifies that the results must be in the range of votes given.
// The range is inclusive.
func (s *Searcher) Votes(min, max int) *Searcher {
	s.votes = &irange{min, max}
	return s
}

// Billed specifies that the results---when they correspond to credits---must
// be in the billed range provided. For example, when showing credits for an
// actor, this will restrict the results to movies where the actor has a billed
// position in this range. Similarly for showing credits for a movie.
// The range is inclusive.
func (s *Searcher) Billed(min, max int) *Searcher {
	s.billing = &irange{min, max}
	return s
}

// Tvshow specifies a sub-search that will be performed when Results is called.
// The TV show returned by this sub-search will be used to filter the results
// of its parent search. If no TV show is found, then the search quits and
// returns no results. If more than one good matching TV show is found, then
// the searcher's "chooser" is called. (See the documentation for the
// Chooser type.)
func (s *Searcher) Tvshow(tvs *Searcher) *Searcher {
	tvs.Entity(imdb.EntityTvshow)
	tvs.what = "TV show"
	s.subTvshow = &subsearch{tvs, 0}
	return s
}

// Episode specifies a sub-search that will be performed when Results is called.
// The entity returned restrict the results of the parent search to only
// include credits for the entity. (Note that TV shows generally don't have
// credits associated with them.)
// If no entity is found, then the parent search quits and returns no results.
func (s *Searcher) Credits(credits *Searcher) *Searcher {
	credits.what = "credits"
	s.subCredits = &subsearch{credits, 0}
	return s
}

// Cast specifies a sub-search that will be performed when Results is called.
// The cast member returned restricts the results of the parent search to only
// include credits for the cast member.
// If no cast member is found, then the parent search quits and returns no
// results.
func (s *Searcher) Cast(cast *Searcher) *Searcher {
	cast.what = "actor"
	s.subCast = &subsearch{cast, 0}
	return s
}

// Limit restricts the number of results to the limit given. If Limit is never
// specified, then the search defaults to a limit of 30.
func (s *Searcher) Limit(n int) *Searcher {
	s.limit = n
	return s
}

// Sort specifies the order in which to return the results.
// Note that Sort can be called multiple times. Each call adds the column and
// order to the current sort criteria.
func (s *Searcher) Sort(column, order string) *Searcher {
	s.order = append(s.order, searchOrder{column, order})
	return s
}

// Chooser specifies the function to call when a sub-search returns 2 or more
// good hits. See the documentation for the Chooser type for details.
func (s *Searcher) Chooser(chooser Chooser) *Searcher {
	s.chooser = chooser
	return s
}

// queryTokens breaks a search query into tokens. Namely, a token is whitespace
// delimited, except when curly braces ('{' and '}') are presents. For example,
// in the string "a b {x y z} c", there are exactly four tokens: "a", "b",
// "{x y z}" and "c".
func queryTokens(query string) []string {
	var tokens []string
	var buf []rune
	curlyDepth := 0
	for _, r := range query {
		switch r {
		case ' ':
			if curlyDepth == 0 {
				if len(buf) > 0 {
					tokens = append(tokens, string(buf))
				}
				buf = nil
			} else {
				buf = append(buf, r)
			}
		case '{':
			curlyDepth++
			buf = append(buf, r)
		case '}':
			curlyDepth--
			buf = append(buf, r)
			if curlyDepth == 0 {
				tokens = append(tokens, string(buf))
				buf = nil
			}
		default:
			buf = append(buf, r)
		}
	}
	if len(buf) > 0 {
		tokens = append(tokens, string(buf))
	}
	return tokens
}

// argOption returns the name and optional value corresponding to a search
// parameter in a query string. Query params are of the form '{name[:val]}'.
func argOption(arg string) (name, val string) {
	if len(arg) < 3 {
		return
	}
	if arg[0] != '{' || arg[len(arg)-1] != '}' {
		return
	}
	arg = arg[1 : len(arg)-1]
	sep := strings.Index(arg, ":")
	if sep == -1 {
		name = arg
	} else {
		name, val = arg[0:sep], arg[sep+1:]
	}
	name, val = strings.TrimSpace(name), strings.TrimSpace(val)
	return
}

// intRange parses a range of integers of the form "x-y" and returns x and y
// as integers. If given only "x", then intRange returns x and x. If given
// "x-", then intRange returns x and max. If given "-x", then intRange returns
// min and x.
func intRange(s string, min, max int) (int, int, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return min, max, nil
	}
	if !strings.Contains(s, "-") {
		n, err := strconv.Atoi(s)
		if err != nil {
			return 0, 0, ef("Could not parse '%s' as integer: %s", s, err)
		}
		return n, n, nil
	}

	var pcs []string
	for _, p := range strings.SplitN(s, "-", 2) {
		pcs = append(pcs, strings.TrimSpace(p))
	}

	start, end := min, max
	var err error
	if len(pcs[0]) > 0 {
		start, err = strconv.Atoi(pcs[0])
		if err != nil {
			return 0, 0, ef("Could not parse '%s' as integer: %s", pcs[0], err)
		}
	}
	if len(pcs[1]) > 0 {
		end, err = strconv.Atoi(pcs[1])
		if err != nil {
			return 0, 0, ef("Could not parse '%s' as integer: %s", pcs[1], err)
		}
	}
	return start, end, nil
}

func (s *Searcher) sql() string {
	q := sf(`
		SELECT
			%s AS entity,
			COALESCE(m.atom_id, t.atom_id, e.atom_id, a.atom_id) AS atom_id,
			name.name AS name,
			COALESCE(m.year, t.year, e.year, 0) AS year,
			%s,
			CASE
				WHEN m.atom_id IS NOT NULL THEN
					trim(
						CASE WHEN m.tv THEN '(TV) ' ELSE '' END
						||
						CASE WHEN m.video THEN '(V)' ELSE '' END
					)
				WHEN t.atom_id IS NOT NULL THEN
					CASE
						WHEN t.year_start > 0 THEN cast(t.year_start AS text)
						ELSE '????'
					END
					|| '-' ||
					CASE
						WHEN t.year_end > 0 THEN cast(t.year_end AS text)
						ELSE '????'
					END
				WHEN e.atom_id IS NOT NULL THEN
					'(TV show: ' || et.name
					||
					CASE
						WHEN e.season > 0 AND e.episode_num > 0 THEN
							', #' || cast(e.season AS text)
							||
							'.' || cast(e.episode_num AS text)
						ELSE ''
					END
					|| ')'
				WHEN a.atom_id IS NOT NULL THEN ''
				ELSE ''
			END
			AS attrs,
			COALESCE(rating.votes, 0) AS votes,
			COALESCE(rating.rank, 0) AS rank,
			%s
		FROM name
		LEFT JOIN movie AS m ON name.atom_id = m.atom_id
		LEFT JOIN tvshow AS t ON name.atom_id = t.atom_id
		LEFT JOIN episode AS e ON name.atom_id = e.atom_id
		LEFT JOIN name AS et ON e.tvshow_atom_id = et.atom_id
		LEFT JOIN actor AS a ON name.atom_id = a.atom_id
		LEFT JOIN rating ON name.atom_id = rating.atom_id
		%s
		WHERE
			COALESCE(m.atom_id, t.atom_id, e.atom_id, a.atom_id) IS NOT NULL
			AND
			%s
		%s
		LIMIT %d
		`,
		s.entityColumn(), s.similarColumn("name.name"), s.creditAttrs(),
		s.creditJoin(), s.where(), s.orderby(), s.limit)
	if s.debug {
		pef("%s\n", q)
	}
	return q
}

func (s *Searcher) creditJoin() string {
	var joins string
	if !s.subCast.empty() {
		joins += sf(`
		LEFT JOIN credit AS c_actor ON
			name.atom_id = c_actor.media_atom_id
			AND c_actor.actor_atom_id = %d
		`, s.subCast.id)
	}
	if !s.subCredits.empty() {
		joins += sf(`
		LEFT JOIN credit AS c_media ON
			a.atom_id = c_media.actor_atom_id
			AND c_media.media_atom_id = %d
		`, s.subCredits.id)
	}
	return joins
}

func (s *Searcher) creditAttrs() string {
	act, med := !s.subCast.empty(), !s.subCredits.empty()
	switch {
	case !act && !med:
		return `
		0 AS c_actor_id,
		0 AS c_media_id,
		'' AS c_character,
		0 AS c_position,
		'' AS c_attrs
		`
	case !act && med:
		return `
		COALESCE(c_media.actor_atom_id, 0) AS c_actor_id,
		COALESCE(c_media.media_atom_id, 0) AS c_media_id,
		COALESCE(c_media.character, '') AS c_character,
		COALESCE(c_media.position, 0) AS c_position,
		COALESCE(c_media.attrs, '') AS c_attrs
		`
	case act && !med:
		return `
		COALESCE(c_actor.actor_atom_id, 0) AS c_actor_id,
		COALESCE(c_actor.media_atom_id, 0) AS c_media_id,
		COALESCE(c_actor.character, '') AS c_character,
		COALESCE(c_actor.position, 0) AS c_position,
		COALESCE(c_actor.attrs, '') AS c_attrs
		`
	case act && med:
		return `
		COALESCE(c_actor.actor_atom_id, c_media.actor_atom_id) AS c_actor_id,
		COALESCE(c_actor.media_atom_id, c_media.media_atom_id) AS c_media_id,
		COALESCE(c_actor.character, c_media.character) AS c_character,
		COALESCE(c_actor.position, c_media.position) AS c_position,
		COALESCE(c_actor.attrs, c_media.attrs) AS c_attrs
		`
	}
	panic("unreachable")
}

func (s *Searcher) where() string {
	var conj []string
	conj = append(conj, s.whereCredits()...)
	if len(s.entities) > 0 {
		var entsIn []string
		for _, e := range s.entities {
			entsIn = append(entsIn, sf("'%s'", e.String()))
		}
		in := sf("%s IN(%s)", s.entityColumn(), strings.Join(entsIn, ", "))
		conj = append(conj, in)
	}
	if !s.subTvshow.empty() {
		conj = append(conj, sf("e.tvshow_atom_id = %d", s.subTvshow.id))
	}
	if s.atom > 0 {
		conj = append(conj, sf("name.atom_id = %d", s.atom))
	}
	if s.year != nil {
		conj = append(conj, s.year.cond("COALESCE(m.year, t.year, e.year, 0)"))
	}
	if s.rating != nil {
		conj = append(conj, s.rating.cond("rating.rank"))
	}
	if s.votes != nil {
		conj = append(conj, s.votes.cond("rating.votes"))
	}
	if s.season != nil {
		cond := sf("(e.atom_id IS NULL OR %s)", s.season.cond("e.season"))
		conj = append(conj, cond)
	}
	if s.episode != nil {
		cond := sf("(e.atom_id IS NULL OR %s)", s.episode.cond("e.episode_num"))
		conj = append(conj, cond)
	}
	if s.noTvMovie {
		conj = append(conj, "(m.atom_id IS NULL OR m.tv = cast(0 as boolean))")
	}
	if s.noVideoMovie {
		conj = append(conj,
			"(m.atom_id IS NULL OR m.video = cast(0 as boolean))")
	}
	if len(s.name) > 0 {
		if s.fuzzy {
			conj = append(conj, "name.name % $1")
		} else {
			if strings.ContainsAny(s.name, "%_") {
				if s.db.Driver == "postgres" {
					conj = append(conj, sf("name.name ILIKE $1"))
				} else {
					conj = append(conj, sf("name.name LIKE $1"))
				}
			} else {
				conj = append(conj, sf("name.name = $1"))
			}
		}
	}
	return strings.Join(conj, " AND ")
}

func (s *Searcher) whereCredits() []string {
	var conj []string
	var joined string
	if !s.subCredits.empty() {
		conj = append(conj, sf("c_media.actor_atom_id IS NOT NULL"))
		joined = "c_media"
	}
	if !s.subCast.empty() {
		conj = append(conj, sf("c_actor.media_atom_id IS NOT NULL"))
		joined = "c_actor"
	}
	if len(joined) > 0 && s.billing != nil {
		conj = append(conj, s.billing.cond(sf("%s.position", joined)))
	}
	return conj
}

func (s *Searcher) orderby() string {
	q, prefix := "", ""
	for _, ord := range s.order {
		qualed := orderColumnQualified(ord.column)
		if len(qualed) == 0 {
			continue
		}
		q += sf("%s%s %s NULLS LAST", prefix, qualed, ord.order)
		prefix = ", "
	}
	if s.fuzzy && len(s.name) > 0 {
		return sf("ORDER BY similarity DESC NULLS LAST %s %s", prefix, q)
	}
	if len(q) == 0 {
		return ""
	}
	return sf("ORDER BY %s", q)
}

func (s *Searcher) entityColumn() string {
	return `
			CASE
				WHEN m.atom_id IS NOT NULL THEN 'movie'
				WHEN t.atom_id IS NOT NULL THEN 'tvshow'
				WHEN e.atom_id IS NOT NULL THEN 'episode'
				WHEN a.atom_id IS NOT NULL THEN 'actor'
				ELSE ''
			END`
}

func (s *Searcher) similarColumn(col string) string {
	if len(s.name) > 0 && s.fuzzy {
		return sf("similarity(%s, $1) AS similarity", col)
	} else {
		return "-1 AS similarity"
	}
}

func (ir *irange) cond(col string) string {
	return sf("%s >= %d AND %s <= %d", col, ir.min, col, ir.max)
}

var ResultColumns = map[string][]string{
	"all":     {"entity", "atom_id", "title", "year", "attrs", "similarity"},
	"episode": {"season", "episode_num"},
}

var qualifiedColumns = map[string]string{
	"entity":     "entity",
	"atom_id":    "atom_id",
	"name":       "name",
	"year":       "year",
	"attrs":      "attrs",
	"similarity": "similarity",

	"season":  "e.season",
	"episode": "e.episode_num",

	"rank":  "rating.rank",
	"votes": "rating.votes",

	"billing": "c_media.position",
}

func orderColumnQualified(column string) string {
	return qualifiedColumns[column]
}

func isValidColumn(ent imdb.EntityKind, column string) bool {
	for _, c := range validColumns(ent) {
		if c == column {
			return true
		}
	}
	return false
}

func validColumns(ent imdb.EntityKind) []string {
	if ent != imdb.EntityNone {
		var cols []string
		for _, col := range ResultColumns["all"] {
			cols = append(cols, col)
		}
		for _, col := range ResultColumns[ent.String()] {
			cols = append(cols, col)
		}
		return cols
	} else {
		return ResultColumns["all"]
	}
}
