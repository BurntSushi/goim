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
					atom_id INTEGER NOT NULL,
					title TEXT NOT NULL,
					year INTEGER NOT NULL,
					sequence TEXT,
					tv INTEGER NOT NULL,
					video INTEGER NOT NULL,
					PRIMARY KEY (atom_id)
				);
				CREATE TABLE tvshow (
					atom_id INTEGER NOT NULL,
					title TEXT NOT NULL,
					year INTEGER NOT NULL,
					sequence TEXT,
					year_start INTEGER,
					year_end INTEGER,
					PRIMARY KEY (atom_id)
				);
				CREATE TABLE episode (
					atom_id INTEGER NOT NULL,
					tvshow_atom_id INTEGER NOT NULL,
					title TEXT NOT NULL,
					year INTEGER NOT NULL,
					season INTEGER NOT NULL,
					episode_num INTEGER NOT NULL,
					PRIMARY KEY (atom_id)
				);
				CREATE TABLE release_date (
					atom_id INTEGER,
					country TEXT,
					released DATE,
					attrs TEXT
				);
				CREATE TABLE running_time (
					atom_id INTEGER,
					country TEXT,
					minutes INTEGER,
					attrs TEXT
				);
				CREATE TABLE aka_title (
					atom_id INTEGER,
					title TEXT NOT NULL,
					attrs TEXT
				);
				CREATE TABLE alternate_version (
					atom_id INTEGER,
					about TEXT
				);
				CREATE TABLE color_info (
					atom_id INTEGER,
					color INTEGER NOT NULL,
					attrs TEXT
				);
				CREATE TABLE mpaa_rating (
					atom_id INTEGER,
					rating TEXT
						CHECK (rating = "G"
						       OR rating = "PG"
							   OR rating = "PG-13"
							   OR rating = "R"
							   OR rating = "NC-17"
							  ),
					reason TEXT
				);
				CREATE TABLE sound_mix (
					atom_id INTEGER,
					mix TEXT,
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
				CREATE TYPE mpaa AS ENUM ('G', 'PG', 'PG-13', 'R', 'NC-17');

				CREATE TABLE atom (
					id INTEGER,
					hash BYTEA NOT NULL,
					PRIMARY KEY (id)
				);
				CREATE TABLE movie (
					atom_id INTEGER,
					title TEXT NOT NULL,
					year SMALLINT NOT NULL,
					sequence TEXT,
					tv BOOLEAN NOT NULL,
					video BOOLEAN NOT NULL,
					PRIMARY KEY (atom_id)
				);
				CREATE TABLE tvshow (
					atom_id INTEGER,
					title TEXT NOT NULL,
					year SMALLINT NOT NULL,
					sequence TEXT,
					year_start SMALLINT,
					year_end SMALLINT,
					PRIMARY KEY (atom_id)
				);
				CREATE TABLE episode (
					atom_id INTEGER,
					tvshow_atom_id INTEGER NOT NULL,
					title TEXT NOT NULL,
					year SMALLINT NOT NULL,
					season SMALLINT NOT NULL,
					episode_num INTEGER NOT NULL,
					PRIMARY KEY (atom_id)
				);
				CREATE TABLE release_date (
					atom_id INTEGER,
					country TEXT,
					released DATE,
					attrs TEXT
				);
				CREATE TABLE running_time (
					atom_id INTEGER,
					country TEXT,
					minutes SMALLINT,
					attrs TEXT
				);
				CREATE TABLE aka_title (
					atom_id INTEGER,
					title TEXT NOT NULL,
					attrs TEXT
				);
				CREATE TABLE alternate_version (
					atom_id INTEGER,
					about TEXT
				);
				CREATE TABLE color_info (
					atom_id INTEGER,
					color BOOLEAN NOT NULL,
					attrs TEXT
				);
				CREATE TABLE mpaa_rating (
					atom_id INTEGER,
					rating mpaa,
					reason TEXT
				);
				CREATE TABLE sound_mix (
					atom_id INTEGER,
					mix TEXT,
					attrs TEXT
				);
				CREATE TABLE tagline (
					atom_id INTEGER,
					tag TEXT
				);
				CREATE TABLE trivia (
					atom_id INTEGER,
					entry TEXT
				);
				CREATE TABLE genre (
					atom_id INTEGER,
					name TEXT
				);
				CREATE TABLE goof (
					atom_id INTEGER,
					goof_type TEXT,
					entry TEXT
				);
				CREATE TABLE language (
					atom_id INTEGER,
					name TEXT,
					attrs TEXT
				);
				CREATE TABLE literature (
					atom_id INTEGER,
					lit_type TEXT,
					ref TEXT
				);
				CREATE TABLE location (
					atom_id INTEGER,
					place TEXT,
					attrs TEXT
				);
				CREATE TABLE link (
					atom_id INTEGER,
					link_type TEXT,
					link_atom_id INTEGER
				);
				CREATE TABLE plot (
					atom_id INTEGER,
					entry TEXT,
					by TEXT
				);
				CREATE TABLE quote (
					atom_id INTEGER,
					entry TEXT
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
		"tvshow_atom_id", "title", "season", "episode_num"},
	},
	{false, "episode", "tv", "", []string{"tvshow_atom_id"}},
	{false, "episode", "tvseason", "", []string{"tvshow_atom_id", "season"}},

	{false, "release_date", "", "", []string{"atom_id"}},
	{false, "running_time", "", "", []string{"atom_id"}},
	{false, "aka_title", "", "", []string{"atom_id"}},
	{false, "alternate_version", "", "", []string{"atom_id"}},
	{false, "color_info", "", "", []string{"atom_id"}},
	{false, "mpaa_rating", "", "", []string{"atom_id"}},
	{false, "sound_mix", "", "", []string{"atom_id"}},

	{false, "movie", "trgm_title", "gin", []string{"title"}},
	{false, "tvshow", "trgm_title", "gin", []string{"title"}},
	{false, "episode", "trgm_title", "gin", []string{"title"}},
	{false, "aka_title", "trgm_title", "gin", []string{"title"}},
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
