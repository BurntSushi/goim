package imdb

import (
	"database/sql"
	"strings"

	"github.com/BurntSushi/csql"

	"github.com/BurntSushi/ty/fun"
)

var DefaultSearch = SearchOptions{
	NoCase:        false,
	Fuzzy:         false,
	Limit:         20,
	Order:         []SearchOrder{{"year", "DESC"}},
	Entities:      nil,
	TvshowId:      0,
	YearMin:       0,
	YearMax:       3000,
	RateMin:       0,
	RateMax:       100,
	SeasonMin:     0,
	SeasonMax:     1000000,
	EpisodeNumMin: 0,
	EpisodeNumMax: 1000000,
}

type SearchOptions struct {
	NoCase                       bool
	Fuzzy                        bool
	Limit                        int
	Order                        []SearchOrder
	Entities                     []EntityKind
	TvshowId                     Atom
	YearMin, YearMax             int
	RateMin, RateMax             int
	SeasonMin, SeasonMax         int
	EpisodeNumMin, EpisodeNumMax int
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

type SearchResultOld struct {
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

type searcher struct {
	db    *DB
	query string
	opts  SearchOptions
}

func (opts SearchOptions) Search(
	db *DB,
	query string,
) ([]SearchResultOld, error) {
	if opts.Entities == nil {
		less := func(e1, e2 EntityKind) bool { return int(e1) < int(e2) }
		opts.Entities = fun.QuickSort(less, fun.Values(Entities)).([]EntityKind)
	}
	return searcher{db, query, opts}.search()
}

func (s searcher) search() ([]SearchResultOld, error) {
	var results []SearchResultOld

	err := csql.Safe(func() {
		subs, prefix := "(", " "
		for _, entity := range s.opts.Entities {
			subs += prefix + s.searchSub(entity) + " "
			prefix = ") UNION ("
		}
		subs += ")"

		var rs *sql.Rows
		if len(s.query) == 0 {
			rs = csql.Query(s.db, s.parentSelect(subs))
		} else {
			rs = csql.Query(s.db, s.parentSelect(subs), s.query)
		}
		csql.Panic(csql.ForRow(rs, func(s csql.RowScanner) {
			var r SearchResultOld
			var ent string
			csql.Scan(s, &ent, &r.Id, &r.Name, &r.Year, &r.Attrs, &r.Similarity)
			r.Entity = Entities[ent]
			results = append(results, r)
		}))
	})
	return results, err
}

func (s searcher) searchSub(entity EntityKind) string {
	switch entity {
	case EntityMovie:
		return s.sqlsubMovie()
	case EntityTvshow:
		return s.sqlsubTvshow()
	case EntityEpisode:
		return s.sqlsubEpisode()
	}
	panic(sf("BUG: unrecognized entity %s", entity))
}

func (s searcher) sqlsubMovie() string {
	return sf(`
		SELECT 
			'%s' AS entity, atom_id, title, year,
			trim(CASE WHEN tv THEN '(TV) ' ELSE '' END
				|| CASE WHEN video THEN '(V)' ELSE '' END)
				AS attrs,
			%s
		FROM movie
		WHERE %s AND %s
		%s
		LIMIT %d
		`,
		EntityMovie.String(),
		s.similarColumn("title"),
		s.years("year"),
		s.cmp("title"),
		s.orderBy(EntityMovie, ""),
		s.opts.Limit*len(s.opts.Entities),
	)
}

func (s searcher) sqlsubTvshow() string {
	return sf(`
		SELECT
			'%s' AS entity, atom_id, title, year,
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
		WHERE %s AND %s
		%s
		LIMIT %d
		`,
		EntityTvshow.String(),
		s.similarColumn("title"),
		s.years("year"),
		s.cmp("title"),
		s.orderBy(EntityTvshow, ""),
		s.opts.Limit*len(s.opts.Entities),
	)
}

func (s searcher) sqlsubEpisode() string {
	return sf(`
		SELECT
			'%s' AS entity, episode.atom_id, episode.title, episode.year,
			'(TV show: ' || tvshow.title
				|| CASE WHEN season > 0 AND episode_num > 0
						THEN ', #' || cast(season AS text)
							|| '.' || cast(episode_num AS text)
						ELSE '' END
				|| ')'
				AS attrs,
			%s
		FROM episode
		LEFT JOIN tvshow ON tvshow.atom_id = episode.tvshow_atom_id
		WHERE %s AND %s AND %s AND %s AND %s
		%s
		LIMIT %d
		`,
		EntityEpisode.String(),
		s.similarColumn("episode.title"),
		s.years("episode.year"),
		s.tvshow("episode.tvshow_atom_id"),
		s.seasons("episode.season"),
		s.episodes("episode.episode_num"),
		s.cmp("episode.title"),
		s.orderBy(EntityEpisode, ""),
		s.opts.Limit*len(s.opts.Entities),
	)
}

func (s searcher) years(column string) string {
	return sf("%s >= %d AND %s <= %d",
		column, s.opts.YearMin, column, s.opts.YearMax)
}

func (s searcher) tvshow(column string) string {
	if s.opts.TvshowId == 0 {
		return s.noop()
	} else {
		return sf("%s = %d", column, s.opts.TvshowId)
	}
}

func (s searcher) noop() string {
	return "1 = 1"
}

func (s searcher) seasons(column string) string {
	return sf("%s >= %d AND %s <= %d",
		column, s.opts.SeasonMin, column, s.opts.SeasonMax)
}

func (s searcher) episodes(column string) string {
	return sf("%s >= %d AND %s <= %d",
		column, s.opts.EpisodeNumMin, column, s.opts.EpisodeNumMax)
}

func (s searcher) similarColumn(column string) string {
	if s.opts.Fuzzy && len(s.query) > 0 {
		return sf("%s AS similarity", s.similarity(column))
	} else {
		return "-1 AS similarity"
	}
}

func (s searcher) cmp(column string) string {
	if len(s.query) == 0 {
		return s.noop()
	}
	if s.opts.Fuzzy {
		return sf("%s %% $1", column)
	} else {
		cmp := "="
		if s.opts.NoCase || strings.ContainsAny(s.query, "%_") {
			if s.db.Driver == "postgres" && s.opts.NoCase {
				cmp = "ILIKE"
			} else {
				cmp = "LIKE"
			}
		}
		return sf("%s %s $1", column, cmp)
	}
}

func (s searcher) similarity(column string) string {
	return sf("similarity(%s, $1)", column)
}

func (s searcher) orderBy(entity EntityKind, colPrefix string) string {
	if s.opts.Fuzzy && len(s.query) > 0 {
		s.opts.Order = append(
			[]SearchOrder{{"similarity", "DESC"}}, s.opts.Order...)
	}
	if len(s.opts.Order) == 0 {
		return ""
	}
	q, prefix := "", ""
	for _, ord := range s.opts.Order {
		if !isValidColumnOld(entity, ord.Column) {
			continue
		}
		q += sf("%s%s%s %s", prefix, colPrefix, ord.Column, srOrder(ord.Order))
		prefix = ", "
	}
	if len(q) == 0 {
		return ""
	}
	return sf("ORDER BY %s", q)
}

func (s searcher) parentSelect(subQueries string) string {
	var cols []string
	for _, col := range SearchResultOldColumns["all"] {
		cols = append(cols, sf("s.%s", col))
	}
	order := s.orderBy(EntityNone, "s.")
	q := sf(`
		SELECT %s
		FROM (%s) AS s
		%s
		LIMIT %d`,
		strings.Join(cols, ", "), subQueries, order, s.opts.Limit)
	// logf("%s", q)
	return q
}

var SearchResultOldColumns = map[string][]string{
	"all":     {"entity", "atom_id", "title", "year", "attrs", "similarity"},
	"episode": {"season", "episode_num"},
}

func isValidColumnOld(ent EntityKind, column string) bool {
	return fun.In(column, validColumnsOld(ent))
}

func validColumnsOld(ent EntityKind) []string {
	if ent != EntityNone {
		var cols []string
		for _, col := range SearchResultOldColumns["all"] {
			cols = append(cols, col)
		}
		for _, col := range SearchResultOldColumns[ent.String()] {
			cols = append(cols, col)
		}
		return cols
	} else {
		return SearchResultOldColumns["all"]
	}
}

func srOrder(o string) string {
	uo := strings.ToUpper(o)
	if uo != "ASC" && uo != "DESC" {
		csql.Panic(ef("Not a valid order: %s (must be one of ASC or DESC)", o))
	}
	return uo
}
