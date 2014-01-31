package imdb

import (
	"strings"

	"github.com/BurntSushi/csql"

	"github.com/BurntSushi/ty/fun"
)

var DefaultSearch = SearchOptions{
	NoCase:    false,
	Fuzzy:     false,
	Limit:     20,
	Order:     []SearchOrder{{"year", "DESC"}},
	Entities:  nil,
	YearStart: 0,
	YearEnd:   3000,
}

type SearchOptions struct {
	NoCase             bool
	Fuzzy              bool
	Limit              int
	Order              []SearchOrder
	Entities           []EntityKind
	YearStart, YearEnd int
}

type SearchOrder struct {
	// Must be one of 'entity', 'id', 'title', 'year' or 'attrs'.
	// Behavior is undefined if is any other value.
	// Note that this string MUST be SQL safe. It is not escaped.
	Column string

	// Must be one of 'ASC' or 'DESC'.
	// Behavior is undefined if is any other value.
	// Note that this string MUST be SQL safe. It is not escaped.
	Order string
}

type SearchResult struct {
	Entity EntityKind
	Id     Atom
	Name   string
	Year   int

	// Arbitrary additional data specific to an entity.
	// e.g., Whether a movie is straight to video or made for TV.
	// e.g., The season and episode number of a TV episode.
	Attrs      string
	Similarity float64
}

func (opts SearchOptions) Search(db *DB, query string) ([]SearchResult, error) {
	entities := opts.Entities
	if entities == nil {
		less := func(e1, e2 EntityKind) bool { return int(e1) < int(e2) }
		entities = fun.QuickSort(less, fun.Values(Entities)).([]EntityKind)
	}
	subs, prefix := "", " "
	for i, entity := range entities {
		subs += prefix + opts.searchSub(db, query, entity, i+1) + " "
		prefix = " UNION "
	}

	var results []SearchResult
	repeatedQuery := opts.repeatedSearch(query, len(entities))
	err := csql.Safe(func() {
		rs := csql.Query(db, opts.parentSelect(subs), repeatedQuery...)
		csql.Panic(csql.ForRow(rs, func(s csql.RowScanner) {
			var r SearchResult
			var ent string
			csql.Scan(s, &ent, &r.Id, &r.Name, &r.Year, &r.Attrs, &r.Similarity)
			r.Entity = Entities[ent]
			results = append(results, r)
		}))
	})
	return results, err
}

func (opts SearchOptions) searchSub(
	db *DB,
	query string,
	entity EntityKind,
	index int,
) string {
	switch entity {
	case EntityMovie:
		return sf(`
			SELECT 
				'%s' AS entity, id, title, year,
				trim(CASE WHEN tv THEN '(TV) ' ELSE '' END
					|| CASE WHEN video THEN '(V)' ELSE '' END)
					AS attrs,
				%s
			FROM movie
			WHERE %s AND %s`,
			entity.String(),
			opts.similarColumn("title", index),
			opts.years("year"),
			opts.cmp(db, "title", query, index),
		)
	case EntityTvshow:
		return sf(`
			SELECT
				'%s' AS entity, id, title, year,
				CASE WHEN year_start > 0
					THEN cast(year_start AS text)
					ELSE '????' END
					|| '-'
					|| CASE WHEN year_end > 0
						THEN cast(year_end AS text)
						ELSE '????' END
					AS attrs,
				%s
			FROM tvshow
			WHERE %s AND %s`,
			entity.String(),
			opts.similarColumn("title", index),
			opts.years("year"),
			opts.cmp(db, "title", query, index),
		)
	case EntityEpisode:
		return sf(`
			SELECT
				'%s' AS entity, episode.id, episode.title, episode.year,
				'(TV show: ' || tvshow.title
					|| CASE WHEN season > 0 AND episode_num > 0
							THEN ', #' || cast(season AS text)
								|| '.' || cast(episode_num AS text)
							ELSE '' END
					|| ')'
					AS attrs,
				%s
			FROM episode
			LEFT JOIN tvshow ON tvshow.id = episode.tvshow_id
			WHERE %s AND %s`,
			entity.String(),
			opts.similarColumn("episode.title", index),
			opts.years("episode.year"),
			opts.cmp(db, "episode.title", query, index),
		)
	}
	panic(sf("BUG: unrecognized entity %s", entity))
}

func (opts SearchOptions) years(column string) string {
	return sf("%s >= %d AND %s <= %d",
		column, opts.YearStart, column, opts.YearEnd)
}

func (opts SearchOptions) similarColumn(column string, index int) string {
	if opts.Fuzzy {
		return sf("%s AS similarity", opts.similarity(column, index))
	} else {
		return "-1 AS similarity"
	}
}

func (opts SearchOptions) cmp(db *DB, column, query string, index int) string {
	if opts.Fuzzy {
		return sf("%s %% $%d", column, index)
	} else {
		cmp := "="
		if opts.NoCase || strings.ContainsAny(query, "%_") {
			if db.Driver == "postgres" && opts.NoCase {
				cmp = "ILIKE"
			} else {
				cmp = "LIKE"
			}
		}
		return sf("%s %s $%d", column, cmp, index)
	}
}

func (opts SearchOptions) similarity(column string, index int) string {
	return sf("similarity(%s, $%d)", column, index)
}

func (opts SearchOptions) orderBy() string {
	if opts.Fuzzy {
		opts.Order = append(
			[]SearchOrder{{"similarity", "DESC"}}, opts.Order...)
	}
	if len(opts.Order) == 0 {
		return ""
	}
	q, prefix := "ORDER BY ", ""
	for _, ord := range opts.Order {
		q += sf("%s%s %s", prefix,
			srColQualified(ord.Column), srOrder(ord.Order))
		prefix = ", "
	}
	return q
}

func (opts SearchOptions) parentSelect(subQueries string) string {
	cols := fun.Map(srColQualified, SearchResultColumns).([]string)
	q := sf(`
		SELECT %s
		FROM (%s) AS s
		%s
		LIMIT %d`,
		strings.Join(cols, ", "), subQueries, opts.orderBy(), opts.Limit)
	return q
}

func (opts SearchOptions) repeatedSearch(q string, nents int) []interface{} {
	repeated := make([]interface{}, nents)
	for i := 0; i < nents; i++ {
		repeated[i] = q
	}
	return repeated
}

var SearchResultColumns = []string{
	"entity", "id", "title", "year", "attrs", "similarity",
}

func srColQualified(name string) string {
	lname := strings.ToLower(name)
	found := false
	for _, n := range SearchResultColumns {
		if n == lname {
			found = true
			break
		}
	}
	if !found {
		fatalf("Not a valid search result column: %s (must be one of %s)",
			name, strings.Join(SearchResultColumns, ", "))
	}
	return sf("s.%s", lname)
}

func srOrder(o string) string {
	uo := strings.ToUpper(o)
	if uo != "ASC" && uo != "DESC" {
		fatalf("Not a valid order: %s (must be one of ASC or DESC)", o)
	}
	return uo
}
