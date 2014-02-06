package main

import (
	"flag"

	"github.com/BurntSushi/goim/tpl"
)

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

	r := c.choose(c.results(db, ""), "")
	if r == nil {
		pef("No results found.")
		return false
	}
	t := c.tpl(sf("info_%s", r.Entity))
	v := tpl.Formatted{tpl.FromSearchResult(db, *r), tpl.Attrs{"Full": true}}
	c.tplExec(t, v)
	return true
}
