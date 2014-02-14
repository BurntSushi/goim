package main

import (
	"flag"

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

var flagSearchIds = false

var cmdSearch = &command{
	name:            "search",
	positionalUsage: "query",
	shortHelp:       "show information about items in the database",
	help:            "",
	flags:           flag.NewFlagSet("search", flag.ExitOnError),
	run:             cmd_search,
	addFlags: func(c *command) {
		c.flags.BoolVar(&flagSearchIds, "ids", flagSearchIds,
			"When set, only the atom identifiers of each search result "+
				"will be printed.")
	},
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
	if flagSearchIds {
		for _, result := range results {
			pf("%d\n", result.Id)
		}
	} else {
		for i, result := range results {
			attrs := tpl.Attrs{"Index": i + 1}
			c.tplExec(template, tpl.Formatted{result, attrs})
		}
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

var attrCommands = map[string]string{
	"running-times":      "show running times (by region) for media",
	"release-dates":      "show release dates (by region) for media",
	"aka-titles":         "show AKA titles for media",
	"alternate-versions": "show alternate versions for media",
	"color-info":         "show color info for media",
	"mpaa":               "show MPAA rating for media",
	"sound-mix":          "show sound mix information for media",
	"taglines":           "show taglines for media",
	"trivia":             "show trivia for media",
	"genres":             "show genres tags for media",
	"goofs":              "show goofs for media",
	"language":           "show language information for media",
	"literature":         "show literature references for media",
	"locations":          "show geography locations for media",
	"links":              "show links (prequels, sequels, versions) of media",
	"plots":              "show plot summaries for media",
	"quotes":             "show quotes for media",
	"rank":               "show user rank/votes for media",
}

func init() {
	for name, help := range attrCommands {
		commands = append(commands, &command{
			name:            name,
			positionalUsage: "query",
			shortHelp:       help,
			flags:           flag.NewFlagSet(name, flag.ExitOnError),
			run:             cmd_attr(name),
		})
	}
}

func cmd_attr(name string) func(*command) bool {
	return func(c *command) bool {
		c.assertLeastNArg(1)
		db := openDb(c.dbinfo())
		defer closeDb(db)

		ent, ok := c.oneEntity(db)
		if !ok {
			return false
		}

		tpl.SetDB(db)
		c.tplExec(c.tpl(name), tpl.Formatted{ent, nil})
		return true
	}
}
