package search

import (
	"strconv"
	"strings"
)

func init() {
	// Add synonyms of commands to the map of commands.
	for _, cmd := range commands {
		for _, synonym := range cmd.synonyms {
			// Don't add to the map while iterating.
			defer func(s string, c *command) { commands[s] = c }(synonym, cmd)
		}
	}
}

// A command is a directive included in a string representation of a search.
// They are of the form '{name:value}', where 'value' is interpreted specially
// depending upon the command.
//
// A command may also have synonyms. For example '{season:1-5}' can also be
// expressed more tersely as '{s:1-5}'.
type command struct {
	desc     string
	synonyms []string
	add      func(s *Searcher, value string) error
}

func addRange(v string, max int, add func(mn, mx int) *Searcher) error {
	if mn, mx, err := intRange(v, 0, max); err != nil {
		return err
	} else {
		add(mn, mx)
		return nil
	}
}

var commands = map[string]*command{
	"debug": {
		"When enabled, the SQL queries used in the search will be logged " +
			"to stderr.",
		nil,
		func(s *Searcher, v string) error {
			s.debug = true
			return nil
		},
	},
	"years": {
		"Only show search results for the year or years specified. " +
			"e.g., {1990-1999} only shows movies in the 90s.",
		[]string{"year"},
		func(s *Searcher, v string) error {
			return addRange(v, maxYear, s.Years)
		},
	},
	"rank": {
		"Only show search results with the rank or ranks specified. " +
			"e.g., {70-} only shows entities with a rank of 70 or better. " +
			"Ranks are on a scale of 0 to 100, where 100 is the best.",
		nil,
		func(s *Searcher, v string) error {
			return addRange(v, maxRank, s.Ranks)
		},
	},
	"votes": {
		"Only show search results with ranks that have the vote count " +
			"specified. e.g., {10000-} only shows entities with a rank that " +
			"has 10,000 or more votes.",
		nil,
		func(s *Searcher, v string) error {
			return addRange(v, maxVotes, s.Votes)
		},
	},
	"billed": {
		"Only show search results with credits with the billing position " +
			"specified. e.g., {1-5} only shows movies where the actor was " +
			"in the top 5 billing order (or only shows actors of a movie " +
			"in the top 5 billing positions).",
		[]string{"billing"},
		func(s *Searcher, v string) error {
			return addRange(v, maxBilled, s.Billed)
		},
	},
	"seasons": {
		"Only show search results for the season or seasons specified. " +
			"e.g., {seasons:1} only shows episodes from the first season " +
			"of a TV show. Note that this only filters episodes---movies and " +
			"TV shows are still returned otherwise.",
		[]string{"s"},
		func(s *Searcher, v string) error {
			return addRange(v, maxSeason, s.Seasons)
		},
	},
	"episodes": {
		"Only show search results for the season or seasons specified. " +
			"e.g., {episodes:1-5} only shows the first five episodes of a " +
			"of a season. Note that this only filters episodes---movies and " +
			"TV shows are still returned otherwise.",
		[]string{"e"},
		func(s *Searcher, v string) error {
			return addRange(v, maxEpisode, s.Episodes)
		},
	},
	"notv": {
		"Removes 'made for TV' movies from the search results.",
		nil,
		func(s *Searcher, v string) error {
			s.NoTvMovies()
			return nil
		},
	},
	"novideo": {
		"Removes 'made for video' movies from the search results.",
		nil,
		func(s *Searcher, v string) error {
			s.NoVideoMovies()
			return nil
		},
	},
	"limit": {
		"Specifies a limit on the total number of search results returned.",
		nil,
		func(s *Searcher, v string) error {
			n, err := strconv.Atoi(v)
			if err != nil {
				return ef("Invalid integer '%s' for limit: %s", v, err)
			}
			s.Limit(int(n))
			return nil
		},
	},
	"sort": {
		"Sorts the search results according to the field given. It may be " +
			"specified multiple times for more specific sorting. Note that " +
			"this doesn't really work with fuzzy searching, since results " +
			"are always sorted by their similarity with the query in a fuzzy " +
			"search. e.g., {sort:episode desc} sorts episode in descending " +
			"(biggest to smallest) order.",
		nil,
		func(s *Searcher, v string) error {
			fields := strings.Fields(v)
			if len(fields) == 0 || len(fields) > 2 {
				return ef("Invalid sort format: '%s'", v)
			}

			var order string
			if len(fields) > 1 {
				order = fields[1]
			} else {
				order = defaultOrders[fields[0]]
				if len(order) == 0 {
					order = "asc"
				}
			}
			s.Sort(fields[0], order)
			return nil
		},
	},
}
