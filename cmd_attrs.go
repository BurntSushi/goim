package main

import (
	"flag"
	"sort"

	"github.com/BurntSushi/ty/fun"

	"github.com/BurntSushi/goim/imdb"
	"github.com/BurntSushi/goim/tpl"
)

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
	"languages":          "show language information for media",
	"literature":         "show literature references for media",
	"locations":          "show geography locations for media",
	"links":              "show links (prequels, sequels, versions) of media",
	"plots":              "show plot summaries for media",
	"quotes":             "show quotes for media",
	"rank":               "show user rank/votes for media",
	"credits":            "show actor/media credits",
}

func init() {
	for name, help := range attrCommands {
		commands = append(commands, &command{
			name:            name,
			other:           true,
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
		return c.showAttr(db, ent, name)
	}
}

func (c *command) showAttr(db *imdb.DB, ent imdb.Entity, name string) bool {
	tpl.SetDB(db)
	c.tplExec(c.tpl(name), tpl.Args{E: ent, A: nil})
	return true
}

var cmdFull = &command{
	name:            "full",
	other:           true,
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

	attrs := fun.Keys(attrCommands).([]string)
	sort.Sort(sort.StringSlice(attrs))

	ent, ok := c.oneEntity(db)
	if !ok {
		return false
	}

	for _, attr := range attrs {
		if !c.showAttr(db, ent, attr) {
			return false
		}
	}
	return true
}

var cmdShort = &command{
	name:            "short",
	other:           true,
	positionalUsage: "query",
	shortHelp:       "show selected information about an entity",
	help:            "",
	flags:           flag.NewFlagSet("short", flag.ExitOnError),
	run:             cmd_short,
}

func cmd_short(c *command) bool {
	c.assertLeastNArg(1)
	db := openDb(c.dbinfo())
	defer closeDb(db)

	attrs := fun.Keys(attrCommands).([]string)
	sort.Sort(sort.StringSlice(attrs))

	ent, ok := c.oneEntity(db)
	if !ok {
		return false
	}

	tplName := sf("short_%s", ent.Type().String())
	tpl.SetDB(db)
	c.tplExec(c.tpl(tplName), tpl.Args{E: ent, A: nil})
	return true
}
