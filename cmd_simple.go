package main

import (
	"flag"
	"sort"
	"strings"

	"github.com/kr/text"

	"github.com/BurntSushi/ty/fun"

	"github.com/BurntSushi/goim/imdb"
	"github.com/BurntSushi/goim/imdb/search"
	"github.com/BurntSushi/goim/tpl"
)

var cmdClean = &command{
	name:      "clean",
	shortHelp: "empties the database",
	help: `
The clean command will remove stale atom and name records from the database.
This is necessary on occasion since the primary key that Goim uses for each
entity can be changed by IMDb (IMDb does not provide surrogate primary keys).

The database is structured in such a way that stale atom and name records only
have one consequence: they take up space. They won't appear in search results.

Attributes for stale atom and name records are automatically deleted when
the corresponding list is updated.

This operation is idempotent.
`,
	flags: flag.NewFlagSet("clean", flag.ExitOnError),
	run:   cmd_clean,
}

func cmd_clean(c *command) bool {
	db := openDb(c.dbinfo())
	defer closeDb(db)

	pef("Not implemented yet.")
	return false
}

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
test to search the names of entities.

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
episodes with 'bart' in the title) and added an additional sorting criteria.
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
			c.tplExec(template, tpl.Args{result, attrs})
		}
	}
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
	c.tplExec(c.tpl(name), tpl.Args{ent, nil})
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
	c.tplExec(c.tpl(tplName), tpl.Args{ent, nil})
	return true
}
