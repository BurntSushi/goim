package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/goim/imdb"
	"github.com/BurntSushi/goim/tpl"
)

// results searches the database for the query defined by command line
// arguments.
func (c *command) results(db *imdb.DB) []imdb.SearchResult {
	var query []string

	opts := imdb.DefaultSearch
	opts.Entities = nil
	opts.NoCase = false
	opts.Limit = 20
	opts.Order = []imdb.SearchOrder{{"year", "desc"}}
	opts.Fuzzy = db.IsFuzzyEnabled()

	for _, arg := range c.flags.Args() {
		name, val := argOption(arg)
		if name == "year" || name == "years" {
			opts.YearStart, opts.YearEnd = intRange(
				val, imdb.DefaultSearch.YearStart, imdb.DefaultSearch.YearEnd)
		} else if name == "limit" {
			n, err := strconv.Atoi(val)
			if err != nil {
				fatalf("Not a valid integer '%s' for limit: %s", val, err)
			}
			opts.Limit = int(n)
		} else if ent, ok := imdb.Entities[name]; ok {
			opts.Entities = append(opts.Entities, ent)
		} else {
			query = append(query, arg)
		}
	}

	results, err := opts.Search(db, strings.Join(query, " "))
	if err != nil {
		fatalf("Error searching: %s", err)
	}
	return results
}

// choose searches the database for the query given. If there is more than
// one result, then a list is displayed from which the user can choose.
// The corresponding search result is then returned (or nil if something went
// wrong).
//
// Note that in order to use this effectively, the flags for searching should
// be enabled. e.g., `cmdSearch.addFlags(your_command)`.
func (c *command) choose(results []imdb.SearchResult) *imdb.SearchResult {
	if len(results) == 0 {
		return nil
	} else if len(results) == 1 {
		return &results[0]
	} else if results[0].Similarity > -1 && results[1].Similarity > -1 {
		first, second := results[0].Similarity, results[1].Similarity
		if first-second >= 0.25 {
			return &results[0]
		}
	}
	template := c.tpl("search_result")
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

func argOption(arg string) (name, val string) {
	if len(arg) < 3 {
		return
	}
	if arg[0] != '{' || arg[len(arg)-1] != '}' {
		return
	}
	arg = arg[1 : len(arg)-1]
	sep := strings.Index(arg, ":")
	if sep == -1 {
		name = arg
	} else {
		name, val = arg[0:sep], arg[sep+1:]
	}
	name, val = strings.TrimSpace(name), strings.TrimSpace(val)
	return
}

func argEntity(arg string) (k imdb.EntityKind, ok bool) {
	if len(arg) < 3 {
		return
	}
	if arg[0] != '{' || arg[len(arg)-1] != '}' {
		return
	}
	arg = arg[1 : len(arg)-1]
	k, ok = imdb.Entities[arg]
	return
}
