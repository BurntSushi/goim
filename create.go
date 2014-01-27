package main

import (
	"flag"
	"strings"
)

var (
	flagCreateSaveTo   = ""
	flagCreateDownload = false
	flagCreateLists    = ""
)

var createLists = []string{"movies", "release-dates"}

var namedFtp = map[string]string{
	"berlin":  "ftp://ftp.fu-berlin.de/pub/misc/movies/database",
	"digital": "ftp://gatekeeper.digital.com.au/pub/imdb",
	"funet":   "ftp://ftp.funet.fi/pub/culture/tv+film/database",
	"uiuc":    "ftp://uiarchive.cso.uiuc.edu/pub/info/imdb",
}

var cmdCreate = &command{
	name: "create",
	positionalUsage: "[ berlin | digital | funet | uiuc | " +
		"ftp://... | http://... | dir ]",
	shortHelp: "populates fresh database with IMDB data",
	help: `
This command loads the current database with the contents of the IMDB
database given. It may be a named FTP location, an FTP url, an HTTP url or
a directory on the local file system. Regardless of how the location is
specified, it must point to a directory contain IMDB gzipped list files.

By default, the 'berlin' public FTP site is used.

Note that this *only* creates and populates a fresh database---it does NOT
update an existing one. See 'goim help update' for updating an existing
database.
`,
	flags: flag.NewFlagSet("create", flag.ExitOnError),
	run:   create,
	addFlags: func(c *command) {
		c.flags.StringVar(&flagCreateSaveTo, "saveto", flagCreateSaveTo,
			"When set, all downloaded files will be saved to the directory\n"+
				"given.")
		c.flags.BoolVar(&flagCreateDownload, "download", flagCreateDownload,
			"When set, the data retrieved will be stored in 'saveto' and the\n"+
				"program will quit without adding it to the database.")
		c.flags.StringVar(&flagCreateLists, "lists", flagCreateLists,
			"Set to a comma separated list of IMDB movie lists to load, with\n"+
				"no whitespace. Only lists named here will be loaded. If not\n"+
				"specified, then all lists are loaded.\n"+
				"Available lists: "+strings.Join(createLists, ", "))
	},
}

func create(c *command) {
	if flagCreateDownload && len(flagCreateSaveTo) == 0 {
		fatalf("The 'download' flag must be used with the 'saveto' flag.")
	}

	getFrom := c.flags.Arg(0)
	if len(getFrom) == 0 {
		getFrom = "berlin"
	}
	fetch := saver{newFetcher(getFrom), flagCreateSaveTo}
	loaders := map[string]listHandler{
		"movies": listMovies, "release-dates": listReleases,
	}
	for _, name := range createLists {
		list := fetch.list(name)
		defer list.Close()

		if flagCreateDownload {
			continue
		}
		if ld := loaders[name]; ld != nil && loaderIn(name, flagCreateLists) {
			func() {
				db := c.db()
				defer closeDb(db)

				if err := listLoad(db, list, ld); err != nil {
					fatalf("Could not store %s list: %s", name, err)
				}
			}()
		}
	}
}

func loaderIn(name, commaSep string) bool {
	commaSep = strings.TrimSpace(commaSep)
	if len(commaSep) == 0 {
		return true
	}
	return strings.Contains(commaSep, name)
}
