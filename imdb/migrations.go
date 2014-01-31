package imdb

import (
	"strings"

	"github.com/BurntSushi/ty/fun"

	"github.com/BurntSushi/csql"

	"github.com/BurntSushi/migration"
)

var migrations = map[string][]migration.Migrator{
	"sqlite3": {
		func(tx migration.LimitedTx) error {
			var err error
			_, err = tx.Exec(`
				CREATE TABLE atom (
					id integer,
					hash BLOB NOT NULL,
					PRIMARY KEY (id)
				);
				CREATE TABLE movie (
					id INTEGER NOT NULL,
					title TEXT NOT NULL,
					year INTEGER NOT NULL,
					sequence TEXT,
					tv INTEGER NOT NULL,
					video INTEGER NOT NULL,
					PRIMARY KEY (id)
				);
				CREATE TABLE tvshow (
					id INTEGER NOT NULL,
					title TEXT NOT NULL,
					year INTEGER NOT NULL,
					sequence TEXT,
					year_start INTEGER,
					year_end INTEGER,
					PRIMARY KEY (id)
				);
				CREATE TABLE episode (
					id INTEGER NOT NULL,
					tvshow_id INTEGER NOT NULL,
					title TEXT NOT NULL,
					year INTEGER NOT NULL,
					season INTEGER NOT NULL,
					episode_num INTEGER NOT NULL,
					PRIMARY KEY (id)
				);
				CREATE TABLE release (
					atom_id INTEGER,
					outlet TEXT
						CHECK (outlet = "movie"
							   OR outlet = "tvshow"
							   OR outlet = "episode"
							  ),
					country TEXT,
					released DATE,
					attrs TEXT
				);
				CREATE TABLE running_time (
					atom_id INTEGER,
					outlet TEXT
						CHECK (outlet = "movie"
							   OR outlet = "tvshow"
							   OR outlet = "episode"
							  ),
					country TEXT,
					minutes INTEGER,
					attrs TEXT
				);
				`)
			return err
		},
	},
	"postgres": {
		func(tx migration.LimitedTx) error {
			var err error
			_, err = tx.Exec(`
				CREATE TYPE medium AS ENUM ('movie', 'tvshow', 'episode');

				CREATE TABLE atom (
					id integer,
					hash bytea NOT NULL,
					PRIMARY KEY (id)
				);
				CREATE TABLE movie (
					id integer,
					title TEXT NOT NULL,
					year INTEGER NOT NULL,
					sequence TEXT,
					tv boolean NOT NULL,
					video boolean NOT NULL,
					PRIMARY KEY (id)
				);
				CREATE TABLE tvshow (
					id integer,
					title TEXT NOT NULL,
					year INTEGER NOT NULL,
					sequence TEXT,
					year_start INTEGER,
					year_end INTEGER,
					PRIMARY KEY (id)
				);
				CREATE TABLE episode (
					id integer,
					tvshow_id integer NOT NULL,
					title TEXT NOT NULL,
					year INTEGER NOT NULL,
					season INTEGER NOT NULL,
					episode_num INTEGER NOT NULL,
					PRIMARY KEY (id)
				);
				CREATE TABLE release (
					atom_id integer,
					outlet medium,
					country TEXT,
					released date,
					attrs TEXT
				);
				`)
			return err
		},
		func(tx migration.LimitedTx) error {
			_, err := tx.Exec(`
				CREATE TABLE running_time (
					atom_id INTEGER,
					outlet medium,
					country TEXT,
					minutes smallint,
					attrs TEXT
				);
			`)
			return err
		},
	},
}

type index struct {
	unique   bool
	table    string
	name     string
	fulltext string // empty, "gin" or "gist"
	columns  []string
}

var indices = []index{
	{true, "atom", "", "", []string{"hash"}},
	{true, "movie", "imdbpk", "", []string{
		"title", "year", "sequence", "tv", "video"},
	},
	{true, "tvshow", "imdbpk", "", []string{"title", "year", "sequence"}},
	{true, "episode", "imdbpk", "", []string{
		"tvshow_id", "title", "season", "episode_num"},
	},
	{false, "episode", "tv", "", []string{"tvshow_id"}},
	{false, "episode", "tvseason", "", []string{"tvshow_id", "season"}},
	{false, "release", "entity", "", []string{"atom_id", "outlet"}},
	{false, "running_time", "entity", "", []string{"atom_id", "outlet"}},

	{false, "movie", "trgm_title", "gin", []string{"title"}},
	{false, "tvshow", "trgm_title", "gin", []string{"title"}},
	{false, "episode", "trgm_title", "gin", []string{"title"}},
}

func (in index) sqlName() string {
	name := in.name
	if len(in.columns) == 0 {
		panic("indices must have at least one column")
	}
	if len(name) == 0 {
		if len(in.columns) > 1 {
			panic("must specify index name for multi-column indices")
		}
		name = in.columns[0]
	}
	return sf("idx_%s_%s", in.table, name)
}

func (in index) sqlCreate(db *DB) string {
	uni := ""
	if in.unique {
		uni = " UNIQUE "
	}
	using, class := "", ""
	if in.isFulltext() {
		using = sf(" USING %s ", in.fulltext)
		switch in.fulltext {
		case "gin":
			class = " gin_trgm_ops"
		case "gist":
			class = " gist_trgm_ops"
		default:
			panic(sf("unrecognized fulltext index type: %s", in.fulltext))
		}
	}
	return sf("CREATE %s INDEX %s ON %s %s (%s%s)",
		uni, in.sqlName(), in.table, using,
		strings.Join(in.columns, ", "), class)
}

func (in index) isFulltext() bool {
	return len(in.fulltext) > 0
}

func (in index) sqlDrop(db *DB) string {
	return sf("DROP INDEX IF EXISTS %s", in.sqlName())
}

func doIndices(db *DB, getSql func(index, *DB) string, tables ...string) error {
	trgmEnabled := db.IsFuzzyEnabled()
	return csql.Safe(func() {
		var q string
		for _, idx := range indices {
			if idx.isFulltext() && !trgmEnabled {
				// Only show the error message if we're on PostgreSQL.
				if db.Driver == "postgres" {
					logf("Skipping fulltext index '%s' since "+
						"the pg_trgm extension is not enabled.", idx.sqlName())
				}
				continue
			}
			if len(tables) == 0 || fun.In(idx.table, tables) {
				q += getSql(idx, db) + "; "
			}
		}
		csql.Exec(db, q)
	})
}

func CreateIndices(db *DB, tables ...string) error {
	return doIndices(db, index.sqlCreate, tables...)
}

func DropIndices(db *DB, tables ...string) error {
	return doIndices(db, index.sqlDrop, tables...)
}
