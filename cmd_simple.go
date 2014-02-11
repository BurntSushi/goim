package main

import (
	"flag"

	"github.com/BurntSushi/goim/imdb"
	"github.com/BurntSushi/goim/tpl"
)

var cmdClean = &command{
	name:      "clean",
	shortHelp: "empties the database such that 'create' can run",
	help:      "",
	flags:     flag.NewFlagSet("clean", flag.ExitOnError),
	run:       cmd_clean,
}

func cmd_clean(c *command) bool {
	db := openDb(c.dbinfo())
	defer closeDb(db)

	if err := db.Clean(); err != nil {
		pef("Error cleaning database: %s", err)
		return false
	}
	return true
}

var cmdSearch = &command{
	name:            "search",
	positionalUsage: "query",
	shortHelp:       "show information about items in the database",
	help:            "",
	flags:           flag.NewFlagSet("search", flag.ExitOnError),
	run:             cmd_search,
}

func cmd_search(c *command) bool {
	c.assertLeastNArg(1)
	db := openDb(c.dbinfo())
	defer closeDb(db)

	template := c.tpl("search_result")
	results, ok := c.results(db, false)
	if !ok {
		return false
	}
	for i, result := range results {
		c.tplExec(template, tpl.Formatted{result, tpl.Attrs{"Index": i + 1}})
	}
	return true
}

var cmdFull = &command{
	name:            "full",
	positionalUsage: "query",
	shortHelp:       "show exhaustive information about an entity",
	help:            "",
	flags:           flag.NewFlagSet("full", flag.ExitOnError),
	run:             cmd_full,
}

func cmd_full(c *command) bool {
	c.assertLeastNArg(1)
	db := openDb(c.dbinfo())
	defer closeDb(db)

	rs, ok := c.results(db, true)
	if !ok {
		return false
	}
	r := rs[0]
	t := c.tpl(sf("info_%s", r.Entity))
	v := tpl.Formatted{tpl.FromSearchResult(db, r), tpl.Attrs{"Full": true}}
	c.tplExec(t, v)
	return true
}

var cmdPlot = &command{
	name:            "plot",
	positionalUsage: "query",
	shortHelp:       "show plot summaries for media",
	help:            "",
	flags:           flag.NewFlagSet("plot", flag.ExitOnError),
	run:             cmd_plot,
}

func cmd_plot(c *command) bool {
	c.assertLeastNArg(1)
	db := openDb(c.dbinfo())
	defer closeDb(db)

	r, ok := c.oneResult(db)
	if !ok {
		return false
	}
	ent, err := r.GetEntity(db)
	if err != nil {
		pef("%s\n", err)
		return false
	}
	plots, err := imdb.Plots(db, ent)
	if err != nil {
		pef("%s\n", err)
		return false
	}
	if len(plots) == 0 {
		pef("No plots found.\n")
		return false
	}
	before := ""
	for _, plot := range plots {
		pf("%s%s\n", before, tpl.HelpWrap(80, plot.String()))
		before = "\n"
	}
	return true
}
