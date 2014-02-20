package main

import (
	"flag"
)

var cmdClean = &command{
	name:      "clean",
	shortHelp: "empties the database",
	help: `
The clean command will remove stale atom and name records from the database.
This is necessary on occasion since the primary key that Goim uses for each
entity can be changed by IMDb (IMDb does not provide surrogate primary keys).

The database is structured in such a way that stale atom and name records only
have one consequence: they take up space. They won't appear in search results.

Attributes for stale atom and name records are automatically deleted when
the corresponding list is updated.

This operation is idempotent.
`,
	flags: flag.NewFlagSet("clean", flag.ExitOnError),
	run:   cmd_clean,
}

func cmd_clean(c *command) bool {
	db := openDb(c.dbinfo())
	defer closeDb(db)

	pef("Not implemented yet.")
	return false
}
