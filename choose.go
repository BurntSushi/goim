package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/ty/fun"

	"github.com/BurntSushi/goim/imdb"
	"github.com/BurntSushi/goim/tpl"
)

// choose searches the database for the query given. If there is more than
// one result, then a list is displayed from which the user can choose.
// The corresponding search result is then returned (or nil if something went
// wrong).
//
// Note that in order to use this effectively, the flags for searching should
// be enabled. e.g., `cmdSearch.addFlags(your_command)`.
func (c *command) choose(
	db *imdb.DB,
	pickExact bool,
	query string,
) *imdb.SearchResult {
	var entities []imdb.EntityKind
	if len(flagSearchTypes) > 0 {
		entities = fun.Map(imdb.EntityKindFromString,
			strings.Split(flagSearchTypes, ",")).([]imdb.EntityKind)
	}

	opts := imdb.DefaultSearch
	opts.Entities = entities
	opts.NoCase = flagSearchNoCase
	opts.Limit = flagSearchLimit
	opts.Order = []imdb.SearchOrder{{flagSearchSort, flagSearchOrder}}
	opts.Fuzzy = db.IsFuzzyEnabled()

	ystart, yend := intRange(flagSearchYear, opts.YearStart, opts.YearEnd)
	opts.YearStart, opts.YearEnd = ystart, yend

	template := c.tpl("search_result")
	results, err := opts.Search(db, query)
	if err != nil {
		fatalf("Error searching: %s", err)
	}
	if len(results) == 0 {
		return nil
	} else if len(results) == 1 {
		return &results[0]
	} else if opts.Fuzzy && pickExact {
		first, second := results[0].Similarity, results[1].Similarity
		if first-second >= 0.25 {
			return &results[0]
		}
	}
	for i, result := range results {
		c.tplExec(template, tpl.Formatted{result, tpl.Attrs{"Index": i + 1}})
	}

	var choice int
	fmt.Printf("Choice [%d-%d]: ", 1, len(results))
	if _, err := fmt.Fscanln(os.Stdin, &choice); err != nil {
		fatalf("Error reading from stdin: %s", err)
	}
	choice--
	if choice == -1 {
		return nil
	} else if choice < -1 || choice >= len(results) {
		fatalf("Invalid choice %d", choice)
	}
	return &results[choice]
}
