package main

import (
	"flag"
	"strings"
)

var cmdInfo = &command{
	name:            "info",
	positionalUsage: "(movie | tvshow | episode) query",
	shortHelp:       "show information about items in the database",
	help:            "",
	flags:           flag.NewFlagSet("info", flag.ExitOnError),
	run:             info,
}

func info(c *command) {
	c.assertLeastNArg(2)

	db := openDb(c.dbinfo())
	defer closeDb(db)

	entity := entityFromString(c.flags.Arg(0))
	query := strings.Join(c.flags.Args()[1:], " ")

	pf("%s : %s\n", entity, query)
}
