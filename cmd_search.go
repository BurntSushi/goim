package main

import (
	"flag"
	"strings"

	"github.com/BurntSushi/ty/fun"

	"github.com/BurntSushi/goim/imdb"
	"github.com/BurntSushi/goim/tpl"
)

var (
	sortedEnts = fun.QuickSort(func(e1, e2 string) bool { return e1 < e2 },
		fun.Keys(imdb.Entities)).([]string)
)

var (
	flagFormatFull = false
)

var (
	flagSearchTypes  = ""
	flagSearchNoCase = false
	flagSearchLimit  = 20
	flagSearchSort   = "year"
	flagSearchOrder  = "desc"
	flagSearchYear   = ""
)

var cmdSearch = &command{
	name:            "search",
	positionalUsage: "query",
	shortHelp:       "show information about items in the database",
	help:            "",
	flags:           flag.NewFlagSet("search", flag.ExitOnError),
	run:             search,
	addFlags: func(c *command) {
		c.flags.StringVar(&flagSearchTypes, "types", flagSearchTypes,
			"A comma separated list of entity names that filters search\n"+
				"results to only entities in this list. There should be no\n"+
				"whitespace. By default, all entities are searched.\n"+
				"Valid entities: "+strings.Join(sortedEnts, ", "))
		c.flags.BoolVar(&flagSearchNoCase, "i", flagSearchNoCase,
			"Always search case insensitively.")
		c.flags.IntVar(&flagSearchLimit, "limit", flagSearchLimit,
			"Restricts the number of search results to the number given.")
		c.flags.StringVar(&flagSearchSort, "sort", flagSearchSort,
			"Sort by one of "+strings.Join(imdb.SearchResultColumns, ", "))
		c.flags.StringVar(&flagSearchOrder, "order", flagSearchOrder,
			"Order results by 'desc' (descending) or 'asc' (ascending).")
		c.flags.StringVar(&flagSearchYear, "year", flagSearchYear,
			"Specify a year or an inclusive range of years to filter the\n"+
				"search. For example '1999' only returns results that were\n"+
				"released/born in 1999. Or, for a range, '1990-1999' will\n"+
				"only return results from the 1990s.")

		c.flags.BoolVar(&flagFormatFull, "full", flagFormatFull,
			"When set, as much information will be shown as possible.")
	},
}

func search(c *command) bool {
	c.assertLeastNArg(1)
	db := openDb(c.dbinfo())
	defer closeDb(db)

	res := c.choose(db, false, strings.Join(c.flags.Args(), " "))
	if res == nil {
		pef("No choices found or selected.")
		return true
	}

	fmtd := tpl.Formatted{
		O: tpl.FromSearchResult(db, *res),
		A: tpl.Attrs{"Full": flagFormatFull},
	}
	c.tplExec(c.tpl(sf("info_%s", res.Entity)), fmtd)
	return true
}
