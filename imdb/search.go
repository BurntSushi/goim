package imdb

import (
	"strings"

	"github.com/BurntSushi/csql"

	"github.com/BurntSushi/ty/fun"
)

var DefaultSearch = SearchOptions{
	NoCase:   false,
	Limit:    100,
	Order:    []SearchOrder{{"year", "DESC"}},
	Entities: nil,
}

type SearchOptions struct {
	NoCase   bool
	Limit    int
	Order    []SearchOrder
	Entities []Entity
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
	Entity Entity
	Id     Atom
	Title  string
	Year   int

	// Arbitrary additional data specific to an entity.
	// e.g., Whether a movie is straight to video or made for TV.
	// e.g., The season and episode number of a TV episode.
	Attrs string
}

func (opts SearchOptions) Search(db *DB, query string) ([]SearchResult, error) {
	entities := opts.Entities
	if entities == nil {
		less := func(e1, e2 Entity) bool { return int(e1) < int(e2) }
		entities = fun.QuickSort(less, fun.Values(Entities)).([]Entity)
	}
	repeatedQuery := make([]interface{}, len(entities))
	subs, prefix := "", " "
	for i, entity := range entities {
		repeatedQuery[i] = query
		subs += prefix + opts.searchSub(db, query, entity, i+1) + " "
		prefix = " UNION "
	}

	var results []SearchResult
	cols := fun.Map(srColQualified, SearchResultColumns).([]string)
	q := sf(`
		SELECT %s
		FROM (%s) AS s
		%s
		LIMIT %d`, strings.Join(cols, ", "), subs, opts.orderBy(), opts.Limit)
	err := csql.Safe(func() {
		rs := csql.Query(db, q, repeatedQuery...)
		csql.SQLPanic(csql.ForRow(rs, func(s csql.RowScanner) {
			var r SearchResult
			var ent string
			csql.Scan(s, &ent, &r.Id, &r.Title, &r.Year, &r.Attrs)
			r.Entity = Entities[ent]
			results = append(results, r)
		}))
	})
	return results, err
}

func (opts SearchOptions) searchSub(
	db *DB,
	query string,
	entity Entity,
	index int,
) string {
	cmp := opts.cmp(db, query)
	switch entity {
	case EntityMovie:
		return sf(`
			SELECT 
				'%s' AS entity, id, title, year,
				trim(CASE WHEN tv THEN '(TV) ' ELSE '' END
					|| CASE WHEN video THEN '(V)' ELSE '' END)
					AS attrs
			FROM movie
			WHERE title %s $%d`, entity.String(), cmp, index)
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
					AS attrs
			FROM tvshow
			WHERE title %s $%d`, entity.String(), cmp, index)
	case EntityEpisode:
		return sf(`
			SELECT
				'%s' AS entity, episode.id, episode.title, episode.year,
				'(' || tvshow.title
					|| CASE WHEN season > 0 AND episode > 0
							THEN ', #' || cast(season AS text)
								|| '.' || cast(episode AS text)
							ELSE '' END
					|| ')'
					AS attrs
			FROM episode
			LEFT JOIN tvshow ON tvshow.id = episode.tvshow_id
			WHERE episode.title %s $%d`, entity.String(), cmp, index)
	}
	panic(sf("BUG: unrecognized entity %s", entity))
}

func (opts SearchOptions) cmp(db *DB, query string) string {
	cmp := "="
	if opts.NoCase || strings.ContainsAny(query, "%_") {
		if db.Driver == "postgres" && opts.NoCase {
			cmp = "ILIKE"
		} else {
			cmp = "LIKE"
		}
	}
	return cmp
}

func (opts SearchOptions) orderBy() string {
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

var SearchResultColumns = []string{"entity", "id", "title", "year", "attrs"}

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
