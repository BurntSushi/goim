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
