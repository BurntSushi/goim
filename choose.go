package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/goim/imdb"
	"github.com/BurntSushi/goim/tpl"
)

var defaultOrders = map[string]string{
	"year": "desc", "rating": "desc", "similarity": "desc",
	"title": "asc", "entity": "asc",
	"season": "asc", "episode_num": "asc",
}

// results searches the database for the query defined by command line
// arguments.
func (c *command) results(db *imdb.DB, query string) []imdb.SearchResult {
	return nil
	// var titleQuery []string
	// var orders []imdb.SearchOrder
	//
	// if len(query) == 0 {
	// query = strings.Join(c.flags.Args(), " ")
	// }
	// opts := imdb.DefaultSearch
	// opts.Entities = nil
	// opts.NoCase = false
	// opts.Limit = 20
	// opts.Fuzzy = db.IsFuzzyEnabled()
	//
	// for _, arg := range queryTokens(query) {
	// name, val := argOption(arg)
	// if ent, ok := imdb.Entities[name]; len(val) == 0 && ok {
	// if opts.TvshowId == 0 {
	// // If the tvshow id is set, then we always restrict searching
	// // to episodes.
	// opts.Entities = append(opts.Entities, ent)
	// }
	// } else if name == "year" || name == "years" {
	// opts.YearMin, opts.YearMax = intRange(
	// val, imdb.DefaultSearch.YearMin, imdb.DefaultSearch.YearMax)
	// } else if name == "s" || name == "season" || name == "seasons" {
	// opts.SeasonMin, opts.SeasonMax = intRange(
	// val, imdb.DefaultSearch.SeasonMin, imdb.DefaultSearch.SeasonMax)
	// } else if name == "e" || name == "episode" || name == "episodes" {
	// opts.EpisodeNumMin, opts.EpisodeNumMax = intRange(
	// val,
	// imdb.DefaultSearch.EpisodeNumMin,
	// imdb.DefaultSearch.EpisodeNumMax)
	// } else if name == "tv" || name == "tvshow" {
	// if len(val) == 0 {
	// fatalf("No query found for 'tvshow'.")
	// }
	// r := c.choose(c.results(db, val+" {tvshow}"),
	// "TV show name is ambiguous. Please choose one:")
	// if r == nil {
	// fatalf("No results for TV show query '%s'.", val)
	// }
	// opts.TvshowId = r.Id
	// opts.Entities = []imdb.EntityKind{imdb.EntityEpisode}
	// } else if name == "limit" {
	// n, err := strconv.Atoi(val)
	// if err != nil {
	// fatalf("Not a valid integer '%s' for limit: %s", val, err)
	// }
	// opts.Limit = int(n)
	// } else if name == "sort" {
	// fields := strings.Fields(val)
	// if len(fields) == 0 || len(fields) > 2 {
	// fatalf("Too little or too much in sort option: '%s'", val)
	// } else {
	// var order string
	// if len(fields) > 1 {
	// order = fields[1]
	// } else {
	// order = defaultOrders[fields[0]]
	// if len(order) == 0 {
	// order = "asc"
	// }
	// }
	// orders = append(orders, imdb.SearchOrder{fields[0], order})
	// }
	// } else {
	// titleQuery = append(titleQuery, arg)
	// }
	// }
	// if orders != nil {
	// opts.Order = orders
	// }
	//
	// results, err := opts.Search(db, strings.Join(titleQuery, " "))
	// if err != nil {
	// fatalf("Error searching: %s", err)
	// }
	// return results
}

// choose searches the database for the query given. If there is more than
// one result, then a list is displayed from which the user can choose.
// The corresponding search result is then returned (or nil if something went
// wrong).
//
// Note that in order to use this effectively, the flags for searching should
// be enabled. e.g., `cmdSearch.addFlags(your_command)`.
func (c *command) choose(
	results []imdb.SearchResult,
	desc string,
) *imdb.SearchResult {
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
	if len(desc) > 0 {
		pf("%s\n", desc)
	}
	template := c.tpl("search_result")
	for i, result := range results {
		c.tplExec(template, tpl.Formatted{result, tpl.Attrs{"Index": i + 1}})
	}

	var choice int
	pf("Choice [%d-%d]: ", 1, len(results))
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

func queryTokens(query string) []string {
	var tokens []string
	var buf []rune
	curlyDepth := 0
	for _, r := range query {
		switch r {
		case ' ':
			if curlyDepth == 0 {
				if len(buf) > 0 {
					tokens = append(tokens, string(buf))
				}
				buf = nil
			} else {
				buf = append(buf, r)
			}
		case '{':
			curlyDepth++
			buf = append(buf, r)
		case '}':
			curlyDepth--
			buf = append(buf, r)
			if curlyDepth == 0 {
				tokens = append(tokens, string(buf))
				buf = nil
			}
		default:
			buf = append(buf, r)
		}
	}
	if len(buf) > 0 {
		tokens = append(tokens, string(buf))
	}
	return tokens
}
