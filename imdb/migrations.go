package imdb

import (
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
					PRIMARY KEY (id),
					UNIQUE (hash)
				);
				CREATE TABLE movie (
					id INTEGER NOT NULL,
					title TEXT NOT NULL,
					year INTEGER NOT NULL,
					sequence TEXT,
					tv INTEGER NOT NULL,
					video INTEGER NOT NULL,
					PRIMARY KEY (id),
					UNIQUE (title, year, sequence, tv, video)
				);
				CREATE TABLE tvshow (
					id INTEGER NOT NULL,
					title TEXT NOT NULL,
					year INTEGER NOT NULL,
					sequence TEXT,
					year_start INTEGER,
					year_end INTEGER,
					PRIMARY KEY (id),
					UNIQUE (title, year, sequence)
				);
				CREATE TABLE episode (
					id INTEGER NOT NULL,
					tvshow_id INTEGER NOT NULL,
					title TEXT NOT NULL,
					year INTEGER NOT NULL,
					season INTEGER NOT NULL,
					episode INTEGER NOT NULL,
					PRIMARY KEY (id),
					UNIQUE (tvshow_id, title, season, episode)
				);
				CREATE TABLE release (
					atom_id INTEGER,
					outlet TEXT
						CHECK (outlet = "movie"
							   OR outlet = "tvshow"
							   OR outlet = "episode"
							  ),
					country TEXT,
					released TEXT,
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
					PRIMARY KEY (id),
					UNIQUE (hash)
				);
				CREATE TABLE movie (
					id integer,
					title TEXT NOT NULL,
					year INTEGER NOT NULL,
					sequence TEXT,
					tv boolean NOT NULL,
					video boolean NOT NULL,
					PRIMARY KEY (id),
					UNIQUE (title, year, sequence, tv, video)
				);
				CREATE TABLE tvshow (
					id integer,
					title TEXT NOT NULL,
					year INTEGER NOT NULL,
					sequence TEXT,
					year_start INTEGER,
					year_end INTEGER,
					PRIMARY KEY (id),
					UNIQUE (title, year, sequence)
				);
				CREATE TABLE episode (
					id integer,
					tvshow_id integer NOT NULL,
					title TEXT NOT NULL,
					year INTEGER NOT NULL,
					season INTEGER NOT NULL,
					episode INTEGER NOT NULL,
					PRIMARY KEY (id),
					UNIQUE (tvshow_id, title, season, episode)
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
	},
}
