package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/goim/imdb"
	"github.com/BurntSushi/goim/tpl"
)

func (c *command) results(db *imdb.DB, one bool) ([]imdb.SearchResult, bool) {
	searcher, err := imdb.NewSearcher(db, strings.Join(c.flags.Args(), " "))
	if err != nil {
		pef("%s\n", err)
		return nil, false
	}
	searcher.Chooser(c.chooser)

	results, err := searcher.Results()
	if err != nil {
		pef("%s\n", err)
		return nil, false
	}
	if len(results) == 0 {
		pef("No results found.\n")
		return nil, false
	}
	if one {
		r, err := searcher.Pick(results)
		if err != nil {
			pef("%s\n", err)
			return nil, false
		}
		if r == nil {
			pef("No results to pick from.\n")
			return nil, false
		}
		return []imdb.SearchResult{*r}, true
	}
	return results, true
}

func (c *command) chooser(
	results []imdb.SearchResult,
	what string,
) (*imdb.SearchResult, error) {
	pf("%s is ambiguous. Please choose one:\n", what)
	template := c.tpl("search_result")
	for i, result := range results {
		c.tplExec(template, tpl.Formatted{result, tpl.Attrs{"Index": i + 1}})
	}

	var choice int
	pf("Choice [%d-%d]: ", 1, len(results))
	if _, err := fmt.Fscanln(os.Stdin, &choice); err != nil {
		return nil, ef("Error reading from stdin: %s", err)
	}
	choice--
	if choice == -1 {
		return nil, nil
	} else if choice < -1 || choice >= len(results) {
		return nil, ef("Invalid choice %d", choice)
	}
	return &results[choice], nil
}
