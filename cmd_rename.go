package main

import (
	"flag"
)

var cmdRename = &command{
	name:            "rename",
	positionalUsage: "file [ file ... ] query",
	shortHelp:       "renames files to match search results",
	help:            "",
	flags:           flag.NewFlagSet("rename", flag.ExitOnError),
	run:             cmd_rename,
}

func cmd_rename(c *command) bool {
	c.assertLeastNArg(2)
	db := openDb(c.dbinfo())
	defer closeDb(db)

	return true
}
