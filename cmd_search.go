package main

import (
	"flag"

	"github.com/BurntSushi/ty/fun"

	"github.com/BurntSushi/goim/imdb"
	"github.com/BurntSushi/goim/tpl"
)

var (
	sortedEnts = fun.QuickSort(func(e1, e2 string) bool { return e1 < e2 },
		fun.Keys(imdb.Entities)).([]string)
)

var cmdSearch = &command{
	name:            "search",
	positionalUsage: "query",
	shortHelp:       "show information about items in the database",
	help:            "",
	flags:           flag.NewFlagSet("search", flag.ExitOnError),
	run:             search,
}

func search(c *command) bool {
	c.assertLeastNArg(1)
	db := openDb(c.dbinfo())
	defer closeDb(db)

	template := c.tpl("search_result")
	results := c.results(db, "")
	for i, result := range results {
		c.tplExec(template, tpl.Formatted{result, tpl.Attrs{"Index": i + 1}})
	}
	return true
}
