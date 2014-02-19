package imdb

import (
	"log"
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
					id INTEGER NOT NULL,
					hash BLOB NOT NULL,
					PRIMARY KEY (id)
				);
				CREATE TABLE name (
					atom_id INTEGER NOT NULL,
					name TEXT NOT NULL
				);
				CREATE TABLE actor (
					atom_id INTEGER NOT NULL,
					sequence TEXT NOT NULL,
					PRIMARY KEY (atom_id)
				);
				CREATE TABLE credit (
					actor_atom_id INTEGER NOT NULL,
					media_atom_id INTEGER NOT NULL,
					character TEXT NOT NULL,
					position INTEGER NOT NULL,
					attrs TEXT NOT NULL
				);
				CREATE TABLE movie (
					atom_id INTEGER NOT NULL,
					year SMALLINT NOT NULL,
					sequence TEXT NOT NULL,
					tv BOOLEAN NOT NULL,
					video BOOLEAN NOT NULL,
					PRIMARY KEY (atom_id)
				);
				CREATE TABLE tvshow (
					atom_id INTEGER NOT NULL,
					year SMALLINT NOT NULL,
					sequence TEXT NOT NULL,
					year_start SMALLINT NOT NULL,
					year_end SMALLINT NOT NULL,
					PRIMARY KEY (atom_id)
				);
				CREATE TABLE episode (
					atom_id INTEGER NOT NULL,
					tvshow_atom_id INTEGER NOT NULL,
					year SMALLINT NOT NULL,
					season SMALLINT NOT NULL,
					episode_num INTEGER NOT NULL,
					PRIMARY KEY (atom_id)
				);
				CREATE TABLE release_date (
					atom_id INTEGER NOT NULL,
					country TEXT NOT NULL,
					released DATE NOT NULL,
					attrs TEXT
				);
				CREATE TABLE running_time (
					atom_id INTEGER NOT NULL,
					country TEXT NOT NULL,
					minutes SMALLINT NOT NULL,
					attrs TEXT
				);
				CREATE TABLE aka_title (
					atom_id INTEGER NOT NULL,
					title TEXT NOT NULL,
					attrs TEXT NOT NULL
				);
				CREATE TABLE alternate_version (
					atom_id INTEGER NOT NULL,
					about TEXT NOT NULL
				);
				CREATE TABLE color_info (
					atom_id INTEGER NOT NULL,
					color BOOLEAN NOT NULL,
					attrs TEXT NOT NULL
				);
				CREATE TABLE sound_mix (
					atom_id INTEGER NOT NULL,
					mix TEXT NOT NULL,
					attrs TEXT NOT NULL
				);
				CREATE TABLE tagline (
					atom_id INTEGER NOT NULL,
					tag TEXT NOT NULL
				);
				CREATE TABLE trivia (
					atom_id INTEGER NOT NULL,
					entry TEXT NOT NULL
				);
				CREATE TABLE genre (
					atom_id INTEGER NOT NULL,
					name TEXT NOT NULL
				);
				CREATE TABLE goof (
					atom_id INTEGER NOT NULL,
					goof_type TEXT NOT NULL,
					entry TEXT NOT NULL
				);
				CREATE TABLE language (
					atom_id INTEGER NOT NULL,
					name TEXT NOT NULL,
					attrs TEXT NOT NULL
				);
				CREATE TABLE literature (
					atom_id INTEGER NOT NULL,
					lit_type TEXT NOT NULL,
					ref TEXT NOT NULL
				);
				CREATE TABLE location (
					atom_id INTEGER NOT NULL,
					place TEXT NOT NULL,
					attrs TEXT NOT NULL
				);
				CREATE TABLE link (
					atom_id INTEGER NOT NULL,
					link_type TEXT NOT NULL,
					link_atom_id INTEGER NOT NULL,
					entity TEXT NOT NULL
						CHECK (entity = 'movie'
						       OR entity = 'tvshow'
							   OR entity = 'episode')
				);
				CREATE TABLE plot (
					atom_id INTEGER NOT NULL,
					entry TEXT NOT NULL,
					by TEXT NOT NULL
				);
				CREATE TABLE quote (
					atom_id INTEGER NOT NULL,
					entry TEXT NOT NULL
				);
				CREATE TABLE rating (
					atom_id INTEGER NOT NULL,
					votes INTEGER NOT NULL,
					rank INTEGER NOT NULL
				);
				CREATE TABLE mpaa_rating (
					atom_id INTEGER NOT NULL,
					rating TEXT NOT NULL
						CHECK (rating = "G"
						       OR rating = "PG"
							   OR rating = "PG-13"
							   OR rating = "R"
							   OR rating = "NC-17"
							  ),
					reason TEXT NOT NULL
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
					id INTEGER NOT NULL,
					hash BYTEA NOT NULL,
					PRIMARY KEY (id)
				);
				CREATE TABLE name (
					atom_id INTEGER NOT NULL,
					name TEXT NOT NULL
				);
				CREATE TABLE actor (
					atom_id INTEGER NOT NULL,
					sequence TEXT NOT NULL,
					PRIMARY KEY (atom_id)
				);
				CREATE TABLE credit (
					actor_atom_id INTEGER NOT NULL,
					media_atom_id INTEGER NOT NULL,
					character TEXT NOT NULL,
					position INTEGER NOT NULL,
					attrs TEXT NOT NULL
				);
				CREATE TABLE movie (
					atom_id INTEGER NOT NULL,
					year SMALLINT NOT NULL,
					sequence TEXT NOT NULL,
					tv BOOLEAN NOT NULL,
					video BOOLEAN NOT NULL,
					PRIMARY KEY (atom_id)
				);
				CREATE TABLE tvshow (
					atom_id INTEGER NOT NULL,
					year SMALLINT NOT NULL,
					sequence TEXT NOT NULL,
					year_start SMALLINT NOT NULL,
					year_end SMALLINT NOT NULL,
					PRIMARY KEY (atom_id)
				);
				CREATE TABLE episode (
					atom_id INTEGER NOT NULL,
					tvshow_atom_id INTEGER NOT NULL,
					year SMALLINT NOT NULL,
					season SMALLINT NOT NULL,
					episode_num INTEGER NOT NULL,
					PRIMARY KEY (atom_id)
				);
				CREATE TABLE release_date (
					atom_id INTEGER NOT NULL,
					country TEXT NOT NULL,
					released DATE NOT NULL,
					attrs TEXT
				);
				CREATE TABLE running_time (
					atom_id INTEGER NOT NULL,
					country TEXT NOT NULL,
					minutes SMALLINT NOT NULL,
					attrs TEXT
				);
				CREATE TABLE aka_title (
					atom_id INTEGER NOT NULL,
					title TEXT NOT NULL,
					attrs TEXT NOT NULL
				);
				CREATE TABLE alternate_version (
					atom_id INTEGER NOT NULL,
					about TEXT NOT NULL
				);
				CREATE TABLE color_info (
					atom_id INTEGER NOT NULL,
					color BOOLEAN NOT NULL,
					attrs TEXT NOT NULL
				);
				CREATE TABLE sound_mix (
					atom_id INTEGER NOT NULL,
					mix TEXT NOT NULL,
					attrs TEXT NOT NULL
				);
				CREATE TABLE tagline (
					atom_id INTEGER NOT NULL,
					tag TEXT NOT NULL
				);
				CREATE TABLE trivia (
					atom_id INTEGER NOT NULL,
					entry TEXT NOT NULL
				);
				CREATE TABLE genre (
					atom_id INTEGER NOT NULL,
					name TEXT NOT NULL
				);
				CREATE TABLE goof (
					atom_id INTEGER NOT NULL,
					goof_type TEXT NOT NULL,
					entry TEXT NOT NULL
				);
				CREATE TABLE language (
					atom_id INTEGER NOT NULL,
					name TEXT NOT NULL,
					attrs TEXT NOT NULL
				);
				CREATE TABLE literature (
					atom_id INTEGER NOT NULL,
					lit_type TEXT NOT NULL,
					ref TEXT NOT NULL
				);
				CREATE TABLE location (
					atom_id INTEGER NOT NULL,
					place TEXT NOT NULL,
					attrs TEXT NOT NULL
				);
				CREATE TABLE link (
					atom_id INTEGER NOT NULL,
					link_type TEXT NOT NULL,
					link_atom_id INTEGER NOT NULL,
					entity TEXT NOT NULL
						CHECK (entity = 'movie'
						       OR entity = 'tvshow'
							   OR entity = 'episode')
				);
				CREATE TABLE plot (
					atom_id INTEGER NOT NULL,
					entry TEXT NOT NULL,
					by TEXT NOT NULL
				);
				CREATE TABLE quote (
					atom_id INTEGER NOT NULL,
					entry TEXT NOT NULL
				);
				CREATE TABLE rating (
					atom_id INTEGER NOT NULL,
					votes INTEGER NOT NULL,
					rank INTEGER NOT NULL
				);
				CREATE TABLE mpaa_rating (
					atom_id INTEGER NOT NULL,
					rating mpaa NOT NULL,
					reason TEXT NOT NULL
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
	{false, "episode", "tv", "", []string{"tvshow_atom_id"}},
	{false, "episode", "tvseason", "", []string{"tvshow_atom_id", "season"}},

	{false, "release_date", "", "", []string{"atom_id"}},
	{false, "running_time", "", "", []string{"atom_id"}},
	{false, "aka_title", "", "", []string{"atom_id"}},
	{false, "alternate_version", "", "", []string{"atom_id"}},
	{false, "color_info", "", "", []string{"atom_id"}},
	{false, "mpaa_rating", "", "", []string{"atom_id"}},
	{false, "sound_mix", "", "", []string{"atom_id"}},
	{false, "genre", "", "", []string{"atom_id"}},
	{false, "tagline", "", "", []string{"atom_id"}},
	{false, "trivia", "", "", []string{"atom_id"}},
	{false, "goof", "", "", []string{"atom_id"}},
	{false, "language", "", "", []string{"atom_id"}},
	{false, "literature", "", "", []string{"atom_id"}},
	{false, "location", "", "", []string{"atom_id"}},
	{false, "link", "", "", []string{"atom_id"}},
	{false, "plot", "", "", []string{"atom_id"}},
	{false, "quote", "", "", []string{"atom_id"}},
	{false, "rating", "", "", []string{"atom_id"}},
	{false, "name", "", "", []string{"atom_id"}},
	{false, "credit", "", "", []string{"actor_atom_id"}},
	{false, "credit", "", "", []string{"media_atom_id"}},

	{false, "name", "trgm_name", "gist", []string{"name"}},
	{false, "aka_title", "trgm_title", "gist", []string{"title"}},
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

func doIndices(
	db *DB,
	getSql func(index, *DB) string,
	tables ...string,
) (err error) {
	defer csql.Safe(&err)

	trgmEnabled := db.IsFuzzyEnabled()
	var q string
	var ok bool
	for _, idx := range indices {
		if idx.isFulltext() && !trgmEnabled {
			// Only show the error message if we're on PostgreSQL.
			if db.Driver == "postgres" {
				log.Printf("Skipping fulltext index '%s' since "+
					"the pg_trgm extension is not enabled.", idx.sqlName())
			}
			continue
		}
		if len(tables) == 0 || fun.In(idx.table, tables) {
			q += getSql(idx, db) + "; "
			ok = true
		}
	}
	if ok {
		csql.Exec(db, q)
	}
	return
}

// CreateIndices creates indices for each of the tables specified. This is
// automatically done for you if you're using 'goim load'.
func (db *DB) CreateIndices(tables ...string) error {
	return doIndices(db, index.sqlCreate, tables...)
}

// DropIndices drops indices for each of the tables specified. It is safe to
// call this with tables that may or may not have indices already created.
// Dropping indices is useful when performing large updates on tables.
// This is automatically done for you if you're using 'goim load'.
func (db *DB) DropIndices(tables ...string) error {
	return doIndices(db, index.sqlDrop, tables...)
}
