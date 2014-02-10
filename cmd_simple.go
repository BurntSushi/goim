package main

import (
	"flag"

	"github.com/BurntSushi/goim/imdb"
	"github.com/BurntSushi/goim/tpl"
)

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
