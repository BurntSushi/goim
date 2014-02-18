package main

import (
	"flag"
	"io"
	path "path/filepath"
	"strings"

	"github.com/kr/text"

	"github.com/BurntSushi/ty/fun"

	"github.com/BurntSushi/goim/imdb"
)

var (
	flagLoadDownload = ""
	flagLoadLists    = "movies"
	flagWarnings     = false
)

// loadLists is the set of all list names that may be passed on the command
// line to be updated. Note that this list also specifies the *order* in
// which lists are updated, which is respected regardless of the order given
// on the command line. (This is important because tables like 'movies' should
// always be updated before their corresponding attribute tables.)
var loadLists = []string{
	"movies", "actors",
	"release-dates", "running-times", "aka-titles",
	"alternate-versions", "color-info", "mpaa-ratings-reasons", "sound-mix",
	"genres", "taglines", "trivia", "goofs", "language", "literature",
	"locations", "movie-links", "quotes", "plot", "ratings",
}

type listHandler func(*imdb.DB, *atomizer, io.ReadCloser) error

var simpleLoaders = map[string]listHandler{
	"release-dates":        listReleaseDates,
	"running-times":        listRunningTimes,
	"aka-titles":           listAkaTitles,
	"alternate-versions":   listAlternateVersions,
	"color-info":           listColorInfo,
	"mpaa-ratings-reasons": listMPAARatings,
	"sound-mix":            listSoundMixes,
	"genres":               listGenres,
	"taglines":             listTaglines,
	"trivia":               listTrivia,
	"goofs":                listGoofs,
	"language":             listLanguages,
	"literature":           listLiterature,
	"locations":            listLocations,
	"movie-links":          listMovieLinks,
	"quotes":               listQuotes,
	"plot":                 listPlots,
	"ratings":              listRatings,
	// Functions for loading movies and actors are excluded from this list
	// since they require some special attention.
}

var cmdLoad = &command{
	name: "load",
	positionalUsage: "[ berlin | digital | funet | uiuc | " +
		"ftp://... | http://... | dir ]",
	shortHelp: "creates/updates database with IMDb data",
	help: `
This command loads the current database with the contents of the IMDb
database given. It may be a named FTP location, an FTP url, an HTTP url or
a directory on the local file system. Regardless of how the location is
specified, it must point to a directory (whether remote or local) containing 
IMDb gzipped list files.

By default, the 'berlin' public FTP site is used and only the 'movies' table
is updated. To update more tables, use the '-lists' flag. It is better to
specify as many lists as possible, since they can be updated in parallel.

This command can create a database from scratch or it can update an existing
one. The update procedure is pretty brutish; in most cases, it truncates the
table it's updating and rebuilds it. The only tables that are immune to this
sort of treatment are 'atom' and 'name'. Therefore, the surrogate primary keys
are preserved across updating *if and only if* the primary keys provided by
IMDb don't change. Unfortunately, IMDb primary keys can change (for example,
by adding a title to an episode). This results in stale rows in the 'atom' and
'name' tables (but will be hidden from search results).
`,
	flags: flag.NewFlagSet("load", flag.ExitOnError),
	run:   cmd_load,
	addFlags: func(c *command) {
		c.flags.StringVar(&flagLoadDownload, "download", flagLoadDownload,
			"When set, the data retrieved will be stored in the directory\n"+
				"specified. Then goim will quit.")
		lists := text.Wrap(strings.Join(loadLists, ", "), 80)
		c.flags.StringVar(&flagLoadLists, "lists", flagLoadLists,
			"Set to a comma separated list of IMDB movie lists to load, with\n"+
				"no whitespace. Only lists named here will be loaded. If not\n"+
				"specified, then only the 'movie' list is loaded.\n"+
				"Use 'all' to load all lists or 'attr' to load all attribute\n"+
				"lists (e.g., quotes, running times, etc.).\n"+
				"Available lists: "+lists)
		c.flags.BoolVar(&flagWarnings, "warn", flagWarnings,
			"When set, warnings messages about the data will be shown.\n"+
				"When enabled, this can produce a lot of output saying that\n"+
				"an could not be found for some entries. This is (likely) a\n"+
				"result of inconsistent data in IMDb's text files.")
	},
}

func cmd_load(c *command) bool {
	driver, dsn := c.dbinfo()
	db := openDb(driver, dsn)
	defer closeDb(db)

	// With SQLite, we can get some performance benefit by disabling
	// synchronous writes.
	// It is still safe from application crashes (e.g., bugs in Goim), but
	// not safe from power failures or operating system crashes.
	// I think we're OK with that, right?
	if db.Driver == "sqlite3" {
		_, err := db.Exec("PRAGMA synchronous = OFF")
		if err != nil {
			pef("Could not disable SQLite synchronous mode: %s", err)
			return false
		}
	}

	// Figure out which lists we're loading and make sure each list name is
	// valid before proceeding.
	var userLoadLists []string
	if flagLoadLists == "all" {
		userLoadLists = loadLists
	} else if flagLoadLists == "attr" {
		for _, name := range loadLists {
			if name == "movies" || name == "actors" {
				continue
			}
			userLoadLists = append(userLoadLists, name)
		}
	} else {
		for _, name := range strings.Split(flagLoadLists, ",") {
			name = strings.ToLower(strings.TrimSpace(name))
			if !fun.In(name, loadLists) {
				pef("%s is not a valid list name. See 'goim help load'.", name)
				return false
			}
			userLoadLists = append(userLoadLists, name)
		}
	}

	// Build the "fetcher" to retrieve lists (whether it be from the file
	// system, HTTP or FTP).
	getFrom := c.flags.Arg(0)
	if len(getFrom) == 0 {
		getFrom = "berlin"
	}

	// If we're downloading, then just do that and quit.
	if len(flagLoadDownload) > 0 {
		// We're just saving to disk, so no need to decompress. Get a plain
		// fetcher.
		fetch := newFetcher(getFrom)
		if fetch == nil {
			return false
		}

		download := func(name string) struct{} {
			if err := downloadList(fetch, name); err != nil {
				pef("%s", err)
			}
			if name == "actors" {
				if err := downloadList(fetch, "actresses"); err != nil {
					pef("%s", err)
				}
			}
			return struct{}{}
		}
		// Limit the download to 3 simultaneous connections.
		fun.ParMapN(download, userLoadLists, 3)
		return true
	}

	// We'll be reading, so get a gzip fetcher.
	fetch := newGzipFetcher(getFrom)
	if fetch == nil {
		return false
	}

	// Figure out which tables we'll be modifying and drop the indices for
	// those tables.
	var tables []string
	for _, name := range userLoadLists {
		tablesForList, ok := listTables[name]
		if !ok {
			pef("BUG: Could not find tables to modify for list %s", name)
			return false
		}
		tables = append(tables, tablesForList...)
	}
	tables = fun.Keys(fun.Set(tables)).([]string)

	logf("Dropping indices for: %s", strings.Join(tables, ", "))
	if err := db.DropIndices(tables...); err != nil {
		pef("Could not drop indices: %s", err)
		return false
	}

	// Before launching into loading---which can be done in parallel---we need
	// to load movies and actors first since they insert data that most of the
	// other lists depend on. Also, they cannot be loaded in parallel since
	// they are the only loaders that *add* atoms to the database.
	if in := loaderIndex("movies", userLoadLists); in > -1 {
		if err := loadMovies(driver, dsn, fetch); err != nil {
			pef("%s", err)
			return false
		}
		userLoadLists = append(userLoadLists[:in], userLoadLists[in+1:]...)
	}
	if in := loaderIndex("actors", userLoadLists); in > -1 {
		if err := loadActors(driver, dsn, fetch); err != nil {
			pef("%s", err)
			return false
		}
		userLoadLists = append(userLoadLists[:in], userLoadLists[in+1:]...)
	}

	// This must be done after movies/actors are loaded so that we get all
	// of their atoms.
	if len(userLoadLists) > 0 {
		logf("Reading atom identifiers from database...")
		atoms, err := newAtomizer(db, nil) // read-only
		if err != nil {
			pef("%s", err)
			return false
		}
		simpleLoad := func(name string) bool {
			loader := simpleLoaders[name]
			if loader == nil {
				// This is a bug since we should have verified all list names.
				logf("BUG: %s does not have a simpler loader.", name)
				return true
			}

			db := openDb(driver, dsn)
			defer closeDb(db)

			list, err := fetch.list(name)
			if err != nil {
				pef("%s", err)
				return false
			}
			defer list.Close()

			if err := loader(db, atoms, list); err != nil {
				pef("Could not store %s list: %s", name, err)
				return false
			}
			return true
		}

		// SQLite doesn't handle concurrent writes very well, so force it
		// to be single-threaded.
		maxConcurrent := flagCpu
		if db.Driver == "sqlite3" {
			maxConcurrent = 1
		}
		fun.ParMapN(simpleLoad, userLoadLists, maxConcurrent)
	}

	logf("Creating indices for: %s", strings.Join(tables, ", "))
	if err := db.CreateIndices(tables...); err != nil {
		pef("Could not create indices: %s", err)
		return false
	}
	return true
}

func downloadList(fetch fetcher, name string) error {
	list, err := fetch.list(name)
	if err != nil {
		return err
	}
	defer list.Close()

	saveto := path.Join(flagLoadDownload, sf("%s.list.gz", name))
	logf("Downloading %s to %s...", name, saveto)
	f := createFile(saveto)
	if _, err := io.Copy(f, list); err != nil {
		return ef("Could not save '%s' to disk: %s", name, err)
	}
	return nil
}

func loadMovies(driver, dsn string, fetch fetcher) error {
	list, err := fetch.list("movies")
	if err != nil {
		return err
	}
	defer list.Close()

	db := openDb(driver, dsn)
	defer closeDb(db)

	if err := listMovies(db, list); err != nil {
		return ef("Could not store movies list: %s", err)
	}
	return nil
}

func loadActors(driver, dsn string, fetch fetcher) error {
	list1, err := fetch.list("actors")
	if err != nil {
		return err
	}
	defer list1.Close()

	list2, err := fetch.list("actresses")
	if err != nil {
		return err
	}
	defer list2.Close()

	db := openDb(driver, dsn)
	defer closeDb(db)

	if err := listActors(db, list1, list2); err != nil {
		return ef("Could not store actors/actresses list: %s", err)
	}
	return nil
}

func loaderIndex(name string, userList []string) int {
	name = strings.ToLower(name)
	for i, load := range userList {
		load = strings.TrimSpace(load)
		if name == strings.ToLower(load) {
			return i
		}
	}
	return -1
}
