package main

import (
	"flag"
	"strings"
)

var (
	flagLoadDownload = ""
	flagLoadLists    = "movies"
)

var loadLists = []string{
	"movies", "release-dates", "running-times", "aka-titles",
}

var namedFtp = map[string]string{
	"berlin":  "ftp://ftp.fu-berlin.de/pub/misc/movies/database",
	"digital": "ftp://gatekeeper.digital.com.au/pub/imdb",
	"funet":   "ftp://ftp.funet.fi/pub/culture/tv+film/database",
	"uiuc":    "ftp://uiarchive.cso.uiuc.edu/pub/info/imdb",
}

var cmdLoad = &command{
	name: "load",
	positionalUsage: "[ berlin | digital | funet | uiuc | " +
		"ftp://... | http://... | dir ]",
	shortHelp: "populates fresh database with IMDB data",
	help: `
This command loads the current database with the contents of the IMDB
database given. It may be a named FTP location, an FTP url, an HTTP url or
a directory on the local file system. Regardless of how the location is
specified, it must point to a directory (whether remote or local) containing 
IMDB gzipped list files.

By default, the 'berlin' public FTP site is used.

This command can create a database from scratch or it can update an existing
one. The update procedure is currently not that sophisticated, and some
portions of it are actually done by wiping existing data and reloading it
from scratch. (e.g., Release dates.) Other portions are append only (movies,
TV shows, episodes), which means that errant data persists.

Because of that, it's generally recommended to rebuild the database by using
the 'clean' command and then running 'load'.
`,
	flags: flag.NewFlagSet("load", flag.ExitOnError),
	run:   load,
	addFlags: func(c *command) {
		c.flags.StringVar(&flagLoadDownload, "download", flagLoadDownload,
			"When set, the data retrieved will be stored in the directory\n"+
				"specified. Then goim will quit.")
		c.flags.StringVar(&flagLoadLists, "lists", flagLoadLists,
			"Set to a comma separated list of IMDB movie lists to load, with\n"+
				"no whitespace. Only lists named here will be loaded. If not\n"+
				"specified, then only the 'movie' list is load.\n"+
				"Use 'all' to load all lists.\n"+
				"Available lists: "+strings.Join(loadLists, ", "))
	},
}

func load(c *command) bool {
	driver, dsn := c.dbinfo()

	getFrom := c.flags.Arg(0)
	if len(getFrom) == 0 {
		getFrom = "berlin"
	}
	fetch := saver{newFetcher(getFrom), flagLoadDownload}
	if fetch.fetcher == nil {
		return false
	}
	loaders := map[string]listHandler{
		"movies":        listMovies,
		"release-dates": listReleaseDates,
		"running-times": listRunningTimes,
		"aka-titles":    listAkaTitles,
	}
	for _, name := range loadLists {
		if !loaderIn(name, flagLoadLists) {
			continue
		}
		list := fetch.list(name)
		defer list.Close()

		if len(flagLoadDownload) > 0 {
			logf("Downloading %s...", name)
			continue
		}
		if ld := loaders[name]; ld != nil {
			ok := func() bool {
				db := openDb(driver, dsn)
				defer closeDb(db)

				if err := listLoad(db, list, ld); err != nil {
					pef("Could not store %s list: %s", name, err)
					return false
				}
				return true
			}()
			if !ok {
				return false
			}
		}
	}
	return true
}

func loaderIn(name, commaSep string) bool {
	commaSep = strings.TrimSpace(commaSep)
	if len(commaSep) == 0 || commaSep == "all" {
		return true
	}
	return strings.Contains(commaSep, name)
}
