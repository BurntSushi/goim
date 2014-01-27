package main

import "flag"

var cmdClean = &command{
	name: "clean",
	shortHelp: "empties the database such that 'create' can run",
	help: "",
	flags: flag.NewFlagSet("clean", flag.ExitOnError),
	run:   clean,
}

func clean(c *command) {
	db := c.db()
	defer closeDb(db)

	if err := db.Clean(); err != nil {
		fatalf("Error cleaning database: %s", err)
	}
}
