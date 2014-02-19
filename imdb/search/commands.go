package search

import (
	"strconv"
	"strings"

	"github.com/BurntSushi/ty/fun"

	"github.com/BurntSushi/goim/imdb"
)

// Commands represents all available search directives that are available
// in a search query string.
var Commands []Command

// Command represents a single search directive available in a search query
// string. Each command has a canonical name, a list of possibly empty
// synonyms and a brief description describing what the directive does.
type Command struct {
	Name        string
	Synonyms    []string
	Description string
}

// A command is a directive included in a string representation of a search.
// They are of the form '{name:value}', where 'value' is interpreted specially
// depending upon the command.
//
// A command may also have synonyms. For example '{season:1-5}' can also be
// expressed more tersely as '{s:1-5}'.
type command struct {
	name        string
	synonyms    []string
	hasArg      bool
	description string
	add         func(s *Searcher, value string) error
}

func addRange(v string, add func(mn, mx int) *Searcher) error {
	if mn, mx, err := intRange(v); err != nil {
		return err
	} else {
		var min, max int = -1, -1
		if mn != nil {
			min = *mn
		}
		if mx != nil {
			max = *mx
		}
		add(min, max)
		return nil
	}
}

func addSub(s *Searcher, name, v string, add func(*Searcher) *Searcher) error {
	sub, err := s.subSearcher(name, v)
	if err != nil {
		return err
	}
	add(sub)
	return nil
}

var (
	// commands corresponds to the single point of truth about all possible
	// search commands. There is exactly one 'command' value for each
	// logical command directive.
	commands []command

	// allCommands represents the same information in commands, except it's
	// represented as a map where keys are command names. (Synonyms are
	// included in the keys.)
	allCommands = map[string]command{}
)

func init() {
	less := func(f1, f2 string) bool { return f1 < f2 }
	fields := fun.QuickSort(less, fun.Keys(qualifiedColumns)).([]string)
	sortFields := strings.Join(fields, ", ")

	commands = []command{
		{
			"movie", nil, false,
			"Restricts results to only include movies. Note that this may " +
				"be combined with other entity types to form a disjunction.",
			func(s *Searcher, v string) error {
				s.Entity(imdb.EntityMovie)
				return nil
			},
		},
		{
			"tvshow", nil, false,
			"Restricts results to only include TV shows. Note that this may " +
				"be combined with other entity types to form a disjunction.",
			func(s *Searcher, v string) error {
				s.Entity(imdb.EntityTvshow)
				return nil
			},
		},
		{
			"episode", nil, false,
			"Restricts results to only include episodes. Note that this may " +
				"be combined with other entity types to form a disjunction.",
			func(s *Searcher, v string) error {
				s.Entity(imdb.EntityEpisode)
				return nil
			},
		},
		{
			"actor", nil, false,
			"Restricts results to only include actors. Note that this may " +
				"be combined with other entity types to form a disjunction.",
			func(s *Searcher, v string) error {
				s.Entity(imdb.EntityActor)
				return nil
			},
		},
		{
			"credits", nil, true,
			"A sub-search for media entities that restricts results to " +
				"only actors media item returned from this sub-search.",
			func(s *Searcher, v string) error {
				return addSub(s, "credits", v, s.Credits)
			},
		},
		{
			"cast", nil, true,
			"A sub-search for cast entities that restricts results to " +
				"only media entities in which the cast member appeared.",
			func(s *Searcher, v string) error {
				return addSub(s, "cast", v, s.Cast)
			},
		},
		{
			"show", nil, true,
			"A sub-search for TV shows that restricts results to " +
				"only episodes in the TV show.",
			func(s *Searcher, v string) error {
				return addSub(s, "show", v, s.Tvshow)
			},
		},
		{
			"debug", nil, false,
			"When enabled, the SQL queries used in the search will be logged " +
				"to stderr.",
			func(s *Searcher, v string) error {
				s.debug = true
				return nil
			},
		},
		{
			"id", []string{"atom"}, true,
			"Precisely selects a single identity with the atom identifier " +
				"given. e.g., {id:123} returns the entity with id 123." +
				"Note that one SHOULD NOT rely on any specific atom " +
				"identifier to always correspond to a specific entity, since " +
				"identifiers can (sadly) change when updating your database.",
			func(s *Searcher, v string) error {
				n, err := strconv.Atoi(v)
				if err != nil {
					return ef("Invalid integer '%s' for atom id: %s", v, err)
				}
				s.Atom(imdb.Atom(n))
				return nil
			},
		},
		{
			"years", []string{"year"}, true,
			"Only show search results for the year or years specified. " +
				"e.g., {1990-1999} only shows movies in the 90s.",
			func(s *Searcher, v string) error {
				return addRange(v, s.Years)
			},
		},
		{
			"rank", nil, true,
			"Only show search results with the rank or ranks specified. " +
				"e.g., {70-} only shows entities with a rank of 70 or " +
				"better. Ranks are on a scale of 0 to 100, where 100 is the " +
				"best.",
			func(s *Searcher, v string) error {
				return addRange(v, s.Ranks)
			},
		},
		{
			"votes", nil, true,
			"Only show search results with ranks that have the vote count " +
				"specified. e.g., {10000-} only shows entities with a rank " +
				"that has 10,000 or more votes.",
			func(s *Searcher, v string) error {
				return addRange(v, s.Votes)
			},
		},
		{
			"billing", []string{"billed"}, true,
			"Only show search results with credits with the billing position " +
				"specified. e.g., {1-5} only shows movies where the actor " +
				"was in the top 5 billing order (or only shows actors of a " +
				"movie in the top 5 billing positions).",
			func(s *Searcher, v string) error {
				return addRange(v, s.Billed)
			},
		},
		{
			"seasons", []string{"s"}, true,
			"Only show search results for the season or seasons specified. " +
				"e.g., {seasons:1} only shows episodes from the first season " +
				"of a TV show. Note that this only filters episodes---movies " +
				"and TV shows are still returned otherwise.",
			func(s *Searcher, v string) error {
				return addRange(v, s.Seasons)
			},
		},
		{
			"episodes", []string{"e"}, true,
			"Only show search results for the season or seasons specified. " +
				"e.g., {episodes:1-5} only shows the first five episodes of " +
				"a of a season. Note that this only filters " +
				"episodes---movies and TV shows are still returned otherwise.",
			func(s *Searcher, v string) error {
				return addRange(v, s.Episodes)
			},
		},
		{
			"notv", nil, false,
			"Removes 'made for TV' movies from the search results.",
			func(s *Searcher, v string) error {
				s.NoTvMovies()
				return nil
			},
		},
		{
			"novideo", nil, false,
			"Removes 'made for video' movies from the search results.",
			func(s *Searcher, v string) error {
				s.NoVideoMovies()
				return nil
			},
		},
		{
			"similar", nil, true,
			"Sets the threshold at which to return results from a fuzzy text " +
				"search. Results scoring below this threshold are omitted. " +
				"Note that setting this value too low can dramatically " +
				"increase the search time.",
			func(s *Searcher, v string) error {
				n, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return ef("Invalid float '%s' for similar: %s", v, err)
				}
				s.SimilarThreshold(n)
				return nil
			},
		},
		{
			"limit", nil, true,
			"Specifies a limit on the total number of search results returned.",
			func(s *Searcher, v string) error {
				n, err := strconv.Atoi(v)
				if err != nil {
					return ef("Invalid integer '%s' for limit: %s", v, err)
				}
				s.Limit(int(n))
				return nil
			},
		},
		{
			"sort", nil, true,
			"Sorts the search results according to the field given. It may " +
				"be specified multiple times for more specific sorting. Note " +
				"that this doesn't really work with fuzzy searching, since " +
				"results are always sorted by their similarity with the " +
				"query in a fuzzy search. e.g., {sort:episode desc} sorts " +
				"episode in descending (biggest to smallest) order. " +
				"Valid sort fields: " + sortFields + ".",
			func(s *Searcher, v string) error {
				fields := strings.Fields(v)
				if len(fields) != 2 {
					return ef("Invalid sort format "+
						"(must have field and order): '%s'", v)
				}
				s.Sort(fields[0], fields[1])
				return nil
			},
		},
	}

	// Add synonyms of commands to the map of commands.
	for _, cmd := range commands {
		allCommands[cmd.name] = cmd
		for _, synonym := range cmd.synonyms {
			allCommands[synonym] = cmd
		}
		Commands = append(Commands, Command{
			Name:        cmd.name,
			Synonyms:    cmd.synonyms,
			Description: cmd.description,
		})
	}
	fun.Sort(func(c1, c2 Command) bool { return c1.Name < c2.Name }, Commands)
}

// intRange parses a range of integers of the form "x-y" and returns x and y
// as integers. If given only "x", then intRange returns x and x. If given
// "x-", then intRange returns x and nil. If given "-x", then intRange returns
// nil and x.
func intRange(s string) (*int, *int, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return nil, nil, nil
	}
	if !strings.Contains(s, "-") {
		n, err := strconv.Atoi(s)
		if err != nil {
			return nil, nil, ef("Could not parse '%s' as integer: %s", s, err)
		}
		return &n, &n, nil
	}

	var pcs []string
	for _, p := range strings.SplitN(s, "-", 2) {
		pcs = append(pcs, strings.TrimSpace(p))
	}

	var start, end *int
	if len(pcs[0]) > 0 {
		nstart, err := strconv.Atoi(pcs[0])
		if err != nil {
			return nil, nil, ef("Could not parse '%s' as int: %s", pcs[0], err)
		}
		start = &nstart
	}
	if len(pcs[1]) > 0 {
		nend, err := strconv.Atoi(pcs[1])
		if err != nil {
			return nil, nil, ef("Could not parse '%s' as int: %s", pcs[1], err)
		}
		end = &nend
	}
	return start, end, nil
}
