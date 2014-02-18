package main

import (
	"bytes"
	"flag"
	"os"
	path "path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/csql"

	"github.com/BurntSushi/goim/imdb/search"
	"github.com/BurntSushi/goim/tpl"
)

var (
	flagRenameGuessMovie   = false
	flagRenameGuessEpisode = false
)

var cmdRename = &command{
	name:            "rename",
	positionalUsage: "file [ file ... ] query",
	shortHelp:       "renames files to match search results",
	help: `
This command is currently experimental. I'm confident that it works, but I
think there is room for it to be smarter. I just haven't figured out how yet.
`,
	flags: flag.NewFlagSet("rename", flag.ExitOnError),
	run:   cmd_rename,
	addFlags: func(c *command) {
		c.flags.BoolVar(&flagRenameGuessMovie, "guess-movie",
			flagRenameGuessMovie,
			"When set, the search query can be omitted and Goim will try "+
				"to guess the movie from the file name.")
		c.flags.BoolVar(&flagRenameGuessEpisode, "guess-episode",
			flagRenameGuessEpisode,
			"When set, the search query can be omitted and Goim will try "+
				"to guess the episode from the file name.")
	},
}

func cmd_rename(c *command) bool {
	if flagRenameGuessMovie {
		return cmd_rename_guess_movie(c)
	}
	if flagRenameGuessEpisode {
		return cmd_rename_guess_episode(c)
	}

	c.assertLeastNArg(2)
	db := openDb(c.dbinfo())
	defer closeDb(db)

	query := c.flags.Arg(0)
	files := c.flags.Args()[1:]
	searcher, err := search.New(db, query)
	if err != nil {
		pef("%s", err)
		return false
	}
	searcher.Chooser(c.chooser)
	searcher.Limit(len(files))

	results, err := searcher.Results()
	if err != nil {
		pef("%s", err)
		return false
	}
	if len(results) == 0 {
		pef("No search results.")
		return false
	}
	if len(results) < len(files) {
		pef("Omitting last %d file(s) since there are only %d search results.",
			len(files)-len(results), len(results))
		files = files[0:len(results)]
	}
	names := renames(c, db, files, results)
	if len(names) == 0 {
		return false
	}
	if len(names) != len(files) {
		pef("BUG: Have %d names but have %d files.", len(names), len(files))
		return false
	}

	for i := range names {
		oldName, newName := files[i], names[i]
		pf("Rename '%s' to '%s'\n", oldName, newName)
	}
	if !areYouSure("Are you sure you want to rename these files?") {
		return true
	}
	for i := range names {
		if err := os.Rename(files[i], names[i]); err != nil {
			pef("Error renaming '%s' to '%s': %s", files[i], names[i], err)
			return false
		}
	}
	return true
}

func renames(
	c *command,
	db csql.Queryer,
	files []string,
	rs []search.Result,
) []string {
	var names []string
	for i := range files {
		file, r := files[i], rs[i]
		t := c.tpl(sf("rename_%s", r.Entity))
		ent, err := r.GetEntity(db)
		if err != nil {
			pef("%s", err)
			return nil
		}

		// If this file is a directory, don't both with extensions.
		stat, err := os.Stat(file)
		if err != nil {
			pef("%s", err)
			return nil
		}
		ext := path.Ext(file)
		if stat.IsDir() {
			ext = ""
		}

		buf := new(bytes.Buffer)
		attrs := tpl.Attrs{"Ext": ext}
		if err := t.Execute(buf, tpl.Args{ent, attrs}); err != nil {
			pef("%s", err)
			return nil
		}
		names = append(names, strings.TrimSpace(buf.String()))
	}
	return names
}

var (
	matchYear = regexp.MustCompile("\b[0-9]{4}\b")
)

func cmd_rename_guess_movie(c *command) bool {
	return true
}

func cmd_rename_guess_episode(c *command) bool {
	return true
}
