package main

import (
	"flag"
	"strings"

	"github.com/kr/text"

	"github.com/BurntSushi/goim/imdb/search"
	"github.com/BurntSushi/goim/tpl"
)

var flagSearchIds = false

var cmdSearch = &command{
	name:            "search",
	positionalUsage: "query",
	shortHelp:       "search IMDb for movies, TV shows, episodes and actors",
	help:            "", // added below in init
	flags:           flag.NewFlagSet("search", flag.ExitOnError),
	run:             cmd_search,
	addFlags: func(c *command) {
		c.flags.BoolVar(&flagSearchIds, "ids", flagSearchIds,
			"When set, only the atom identifiers of each search result "+
				"will be printed.")
	},
}

func init() {
	var directives []string
	for _, cmd := range search.Commands {
		s := cmd.Name
		if len(cmd.Synonyms) > 0 {
			s += sf(" (synonyms: %s)", strings.Join(cmd.Synonyms, ", "))
		}
		s += "\n"
		s += text.Indent(text.Wrap(cmd.Description, 78), "  ")
		directives = append(directives, s)
	}
	cmdDoc := strings.Join(directives, "\n\n")

	cmdSearch.help = sf(`
The search command exposes a flexible interface for quickly searching IMDb
for entities, where entities includes movies, TV shows, episodes and actors.

A search query has two different components: text to search the names of 
entities in the database and directives to do additional filtering on 
attributes of entities (like year released, episode number, cast/credits, 
etc.). Included in those directives are options to sort the results or specify 
a limit on the number of results returned.

The search query is composed of whitespace delimited tokens. Each token that 
starts and ends with a '{' and '}' is a directive. All other tokens are used as 
text to search the names of entities.

If you're using PostgreSQL with the 'pg_trgm' extension enabled, then text 
searching is fuzzy. Otherwise, text may contain the wildcard '%%' which matches 
any sequence of characters or the wildcard '_' which matches any single 
character. Whenever a wildcard character is used, fuzzy search is disabled (and 
the search will be case insensitive).

Directives have the form '{NAME[:ARGUMENT]}', where NAME is the name of the 
directive and ARGUMENT is an argument for the directive. Each directive either 
requires no argument or requires a single argument.

Examples
--------
The following are some example query strings. They can be used in 'goim search'
as is. Note that examples without wildcards assume that a PostgreSQL database 
is used with the 'pg_trgm' extension enabled. Some also assume that your 
database has certain data (for example, the 'actors' list must be loaded to use 
the 'cast' and 'credits' directives).

Find all entities with names beginning with 'The Matrix' (case insensitive):

  'the matrix%%'

Now restrict those results to only movies:

  'the matrix%%' {movie}

Or restrict them further by only listing movies where Keanu Reeves is a 
credited cast member:

  'the matrix%%' {movie} {cast:keanu reeves}

Finally, sort the list of movies by IMDb rank and restrict the results to only
movies with 10,000 votes or more:

  'the matrix%%' {movie} {cast:keanu reeves} {sort:rank desc} {votes:10000-}

We could also search in the other direction, for example, by finding the top
5 credits in the movie The Matrix:

  {credits:the matrix} {billing:1-5} {sort:billing asc}

If you try this with 'goim search', then you'll get a prompt that 'credits is
ambiguous' with a list of entities to choose. This can be rather inconvenient
to see every time. Luckily, directives like 'credits' and 'cast' are actually
entire sub-searches that support directives themselves. For example, we can 
specify that the matrix is a movie, which should be enough to make an 
umabiguous selection:

  {credits:the matrix {movie}} {billing:1-5} {sort:billing asc}

Let's switch gears and look at searching episodes for television shows. For 
example, we can list the episode names for the first season of The Simpsons:

  {show:simpsons} {seasons:1} {sort:episode asc}

Note here that there is no text to search here. But we could add some if we 
wanted to, for example, to see all episodes in the entire series with 'bart' in 
the title:

  {show:simpsons} {sort:season asc} {sort:episode asc} '%%bart%%' {limit:1000}

Note the changes here: we removed the restriction on the first season, added
a limit of 1000 (since the default limit is 30, but there may be more than 30 
episodes with 'bart' in the title) and added an additional sorting criterion.
In this case, we want to sort by season first and then by episode. (The order
in which they appear in the query matters.)

We can view this data in a lot of different ways, for example, by finding the
top 10 best ranked Simpsons episodes with more than 500 votes:

  {show:simpsons} {sort:rank desc} {limit:10} {votes:500-}

All search directives
---------------------
%s
`, cmdDoc)
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
			c.tplExec(template, tpl.Args{E: result, A: attrs})
		}
	}
	return true
}
