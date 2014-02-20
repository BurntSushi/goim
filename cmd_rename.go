package main

// This command seems to work well, but the code below is a mess. It was
// written stream-of-consciousness, but I think there are more opportunities
// for abstraction. In particular, I think it should be easier to customize
// the behavior of the "smart" modes of this command.
//
// TODO: Rewrite (or major refactor) the 'rename' command when we get a better
// idea of what its behavior should be.

import (
	"bytes"
	"flag"
	"os"
	path "path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/BurntSushi/ty/fun"

	"github.com/BurntSushi/goim/imdb"
	"github.com/BurntSushi/goim/imdb/search"
	"github.com/BurntSushi/goim/tpl"
)

var (
	flagRenameTvshow       = ""
	flagRenameRegexEpisode = `\b[Ss]([0-9]+)[Ee]([0-9]+)\b`
	flagRenameRegexYear    = `\b([0-9]{4})\b`
	flagRenameTvshowName   = false
)

var cmdRename = &command{
	name:            "rename",
	positionalUsage: "[ query ] file [ file ... ]",
	shortHelp:       "renames files to match search results",
	help: `
NOTICE: Note that this command is in BETA. It should work OK, but I'm not
sold completely on its behavior. This means that its interface could change.

This command will ALWAYS prompt you before renaming your files.

The rename command renames files to names found in IMDb's database. The naming
scheme is specified in templates with the "rename_" prefix found in your
command.tpl file. 

The most general operation is to provide an arbitrary search query as the first
argument (so it must be in quotes, unlike with 'goim search') and a set of
files. This command will assume that the search results are in correspondence
with your files and attempt to rename them as such. This is the most cumbersome
interface but it is also the most precise. In this mode, Goim does not try to
guess anything---the results of your search are taken for gospel.

This command can also be smart by looking for key pieces of information that
are frequently in similar formats (like years or season/episode numbers). Note
that this "smart" mode assumes that fuzzy searching is available.

If the first argument is a file name (i.e., the query is omitted), then Goim
will try to be smart and guess what the file corresponds to based on any name
or year information. This only works with movies or episodes. For movies, a
big piece of distinguishing information is the year, which is extracted with
the regular expression in the 'match-year' flag.

If you're renaming multiple episodes, use the '-tv' flag to specify the TV show 
and omit the query. This will also be significantly faster, since only one
search will be performed instead of a search for each file. Goim will try to 
extract episode numbers from the file names. This extraction can be controlled 
with the 'match-episode' flag.
`,
	flags: flag.NewFlagSet("rename", flag.ExitOnError),
	run:   cmd_rename,
	addFlags: func(c *command) {
		c.flags.StringVar(&flagRenameTvshow, "tv", flagRenameTvshow,
			"Set this to a search query, and the set of files will be\n"+
				"interpreted as a list of episodes for the matching TV show.\n"+
				"When this is set, episode names are guessed. No additional\n"+
				"search query should be provided.")
		c.flags.BoolVar(&flagRenameTvshowName, "tvname",
			flagRenameTvshowName,
			"When set, the name of the TV show is included as a prefix\n"+
				"when renaming an episode.")
		c.flags.StringVar(&flagRenameRegexEpisode, "match-episode",
			flagRenameRegexEpisode,
			"An RE2 regular expression for matching the season and episode\n"+
				"numbers in an episode file name. The regex MUST contain two\n"+
				"capturing groups, where the first is the season number and\n"+
				"the second is the episode number.")
		c.flags.StringVar(&flagRenameRegexYear, "match-year",
			flagRenameRegexYear,
			"An RE2 regular expression for matching the year in a file name.\n"+
				"The regex MUST contain one capturing group for the year\n"+
				"as an integer.")
	},
}

func cmd_rename(c *command) bool {
	if len(flagRenameTvshow) > 0 {
		return cmd_rename_tvshow(c, flagRenameTvshow)
	}
	firstArg := c.flags.Arg(0)
	if _, err := os.Stat(firstArg); err == nil {
		return cmd_rename_smart(c)
	}

	c.assertLeastNArg(2)
	db := openDb(c.dbinfo())
	defer closeDb(db)

	query := c.flags.Arg(0)
	files := fun.Map(path.Clean, c.flags.Args()[1:]).([]string)
	searcher, err := search.Query(db, query)
	if err != nil {
		pef("%s", err)
		return false
	}
	searcher.Chooser(c.chooser)
	searcher.Limit(len(files))

	results, err := searcher.Results()
	if err != nil {
		pef("%s", err)
		return false
	}
	if len(results) == 0 {
		pef("No search results.")
		return false
	}
	if len(results) < len(files) {
		pef("Omitting last %d file(s) since there are only %d search results.",
			len(files)-len(results), len(results))
		files = files[0:len(results)]
	}

	entities := make([]imdb.Entity, len(results))
	for i, r := range results {
		e, err := r.GetEntity(db)
		if err != nil {
			pef("Could not get entity for '%s': %s", r, err)
			return false
		}
		entities[i] = e
	}
	return doRename(c, db, files, entities)
}

func cmd_rename_smart(c *command) bool {
	c.assertLeastNArg(1)
	db := openDb(c.dbinfo())
	defer closeDb(db)

	files := fun.Map(path.Clean, c.flags.Args()).([]string)

	var oldNames []string
	var newNames []imdb.Entity
	for _, file := range files {
		e, err := guessEntity(c, db, file)
		if err != nil {
			pef("Could not guess entity for '%s': %s", file, err)
			continue
		}
		oldNames = append(oldNames, file)
		newNames = append(newNames, e)
	}
	return doRename(c, db, oldNames, newNames)
}

func cmd_rename_tvshow(c *command, tvQuery string) bool {
	c.assertLeastNArg(1)
	db := openDb(c.dbinfo())
	defer closeDb(db)

	files := fun.Map(path.Clean, c.flags.Args()).([]string)
	tv, err := searchTvshow(c, db, tvQuery)
	if err != nil {
		pef("%s", err)
		return false
	}
	episodes, err := tvEpisodes(db, tv)
	if err != nil {
		pef("%s", err)
		return false
	}

	var oldNames []string
	var newNames []imdb.Entity
	for _, file := range files {
		baseFile := path.Base(file)
		s, e, _, _, err := episodeNumbers(baseFile, flagRenameRegexEpisode)
		if err != nil {
			pef("Could not find episode numbers in '%s': %s", file, err)
			continue
		}
		if s == 0 || e == 0 {
			pef("Found numbers, but they look wrong: (s: %d, e: %d)", s, e)
			continue
		}
		ep, ok := episodes[episodeKey{s, e}]
		if !ok {
			pef("Could not find episode (%s, S%02dE%02d) for '%s'.",
				tv, s, e, file)
			continue
		}
		oldNames = append(oldNames, file)
		newNames = append(newNames, ep)
	}
	return doRename(c, db, oldNames, newNames)
}

func doRename(
	c *command,
	db *imdb.DB,
	files []string,
	entities []imdb.Entity,
) bool {
	names := renames(c, db, files, entities)
	if len(names) == 0 {
		return false
	}
	if len(names) != len(files) {
		pef("BUG: Have %d names but have %d files.", len(names), len(files))
		return false
	}

	for i := range names {
		oldName, newName := files[i], names[i]
		pf("Rename '%s' to '%s'\n", oldName, newName)
	}
	if !areYouSure("Are you sure you want to rename these files?") {
		return true
	}
	for i := range names {
		if err := os.Rename(files[i], names[i]); err != nil {
			pef("Error renaming '%s' to '%s': %s", files[i], names[i], err)
			return false
		}
	}
	return true
}

func renames(
	c *command,
	db *imdb.DB,
	files []string,
	entities []imdb.Entity,
) []string {
	var names []string
	tpl.SetDB(db)
	for i := range files {
		file, ent := files[i], entities[i]
		t := c.tpl(sf("rename_%s", ent.Type()))

		// If this file is a directory, don't both with extensions.
		stat, err := os.Stat(file)
		if err != nil {
			pef("%s", err)
			return nil
		}
		ext := path.Ext(file)
		if stat.IsDir() {
			ext = ""
		}

		buf := new(bytes.Buffer)
		attrs := tpl.Attrs{"Ext": ext, "ShowTv": flagRenameTvshowName}
		if err := t.Execute(buf, tpl.Args{E: ent, A: attrs}); err != nil {
			pef("%s", err)
			return nil
		}
		name := strings.TrimSpace(buf.String())
		name = path.Join(path.Dir(file), name)
		names = append(names, name)
	}
	return names
}

func searchTvshow(c *command, db *imdb.DB, query string) (*imdb.Tvshow, error) {
	tvsearch, err := search.Query(db, query)
	if err != nil {
		return nil, err
	}
	tvsearch.Chooser(c.chooser)
	tvsearch.Entity(imdb.EntityTvshow)

	results, err := tvsearch.Results()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, ef("Could not find TV show.")
	}
	tv, err := tvsearch.Pick(results)
	if err != nil {
		return nil, err
	}
	if tv == nil {
		return nil, ef("No TV show results to pick from.")
	}
	if tv.Entity != imdb.EntityTvshow {
		return nil, ef("Expected TV show but got %s", tv.Entity)
	}
	ent, err := tv.GetEntity(db)
	if err != nil {
		return nil, err
	}
	return ent.(*imdb.Tvshow), nil
}

type episodeKey struct {
	s, e int
}

type episodeMap map[episodeKey]*imdb.Episode

func tvEpisodes(db *imdb.DB, tv *imdb.Tvshow) (episodeMap, error) {
	episodes := make(episodeMap, 30)
	epsearch := search.New(db)
	epsearch.Entity(imdb.EntityEpisode)
	epsearch.Tvshow(search.New(db).Atom(tv.Id))
	epsearch.Seasons(1, -1).Episodes(1, -1)
	epsearch.Limit(-1)

	results, err := epsearch.Results()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, ef("Could not find any episodes for %s", tv)
	}
	for _, r := range results {
		ent, err := r.GetEntity(db)
		if err != nil {
			return nil, err
		}
		ep := ent.(*imdb.Episode)
		episodes[episodeKey{ep.Season, ep.EpisodeNum}] = ep
	}
	return episodes, nil
}

func guessEntity(c *command, db *imdb.DB, fname string) (imdb.Entity, error) {
	// Look for episode numbers. If we can find them, then this is an
	// episode.
	fname = path.Base(fname)
	_, _, _, _, err := episodeNumbers(fname, flagRenameRegexEpisode)
	if err == nil {
		return guessEpisode(c, db, fname)
	} else {
		return guessMovie(c, db, fname)
	}
}

func guessEpisode(
	c *command,
	db *imdb.DB,
	fname string,
) (*imdb.Episode, error) {
	fname = path.Base(fname)
	s, e, start, _, err := episodeNumbers(fname, flagRenameRegexEpisode)
	if err != nil {
		return nil, ef("Could not find episode numbers: %s", err)
	}

	// A guess at where the TV show name is in the file name.
	title := strings.TrimSpace(fname[0:start])

	tvsub, err := search.Query(db, title)
	if err != nil {
		return nil, err
	}
	tvsub.Entity(imdb.EntityTvshow)

	esearch := search.New(db)
	esearch.Tvshow(tvsub)
	esearch.Entity(imdb.EntityEpisode)
	esearch.Seasons(s, s).Episodes(e, e)
	esearch.Chooser(c.chooser)

	results, err := esearch.Results()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, ef("Could not find episode.")
	}
	m, err := esearch.Pick(results)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ef("No episode results to pick from.")
	}
	if m.Entity != imdb.EntityEpisode {
		return nil, ef("Expected episode but got %s", m.Entity)
	}
	ent, err := m.GetEntity(db)
	if err != nil {
		return nil, err
	}
	return ent.(*imdb.Episode), nil
}

func guessMovie(c *command, db *imdb.DB, fname string) (*imdb.Movie, error) {
	fname = path.Base(fname)
	year, ystart, _, err := fileNameYear(fname, flagRenameRegexYear)
	if err != nil {
		return nil, ef("Could not find year for movie: %s", err)
	}

	// A guess at where the title is in the file name.
	title := strings.TrimSpace(fname[0:ystart])

	msearch, err := search.Query(db, title)
	if err != nil {
		return nil, err
	}
	msearch.Entity(imdb.EntityMovie)
	msearch.Years(year-1, year+1)
	msearch.Chooser(c.chooser)

	results, err := msearch.Results()
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, ef("Could not find movie.")
	}
	m, err := msearch.Pick(results)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ef("No movie results to pick from.")
	}
	if m.Entity != imdb.EntityMovie {
		return nil, ef("Expected movie but got %s", m.Entity)
	}
	ent, err := m.GetEntity(db)
	if err != nil {
		return nil, err
	}
	return ent.(*imdb.Movie), nil
}

// fileNameYear attempts to extract the year of release for an entity from a
// file name using the RE2 regular expression provided. If the match is
// unsuccessful, 0, -1, -1 is returned. The regex must have exactly 1 capturing
// group, where the group corresponds to the year which must be an integer.
// The triple of integers returned correspond to the year and the start and end
// indices of the year in 'fname'.
//
// fname should be a base name of a file path.
func fileNameYear(fname, regex string) (year, start, end int, err error) {
	reg, err := regexp.Compile(regex)
	if err != nil {
		return 0, -1, -1, ef("Could not compile regex '%s': %s", regex, err)
	}
	groups := reg.FindStringSubmatchIndex(fname)
	if len(groups) != 4 || groups[2] == -1 || groups[3] == -1 {
		return 0, -1, -1, ef("Unsuccessful match.")
	}
	start, end = groups[2], groups[3]
	year64, err := strconv.Atoi(fname[groups[2]:groups[3]])
	if err != nil {
		return 0, -1, -1, ef("Could not parse '%s' as int: %s", groups[1], err)
	}
	year = int(year64)
	return
}

// episodeNumbers attempts to extract the season and episode numbers from
// a file name. If the match is unsuccessfull, zeroes are returned. The regex
// must have exactly two capturing groups, where the first is the season and
// the second is the episode. Both of these capturing groups must correspond
// to integers.
//
// fname should be a base name of a file path.
func episodeNumbers(
	fname,
	regex string,
) (season, episode, start, end int, err error) {
	fname = path.Base(fname)
	reg, err := regexp.Compile(regex)
	if err != nil {
		return 0, 0, -1, -1, ef("Could not compile regex '%s': %s", regex, err)
	}
	groups := reg.FindStringSubmatchIndex(fname)
	if len(groups) != 6 || groups[2] == -1 || groups[5] == -1 {
		return 0, 0, -1, -1, ef("Unsuccessful match.")
	}

	sstart, send := groups[2], groups[3]
	estart, eend := groups[4], groups[5]
	start, end = sstart, eend
	nseason, err := strconv.Atoi(fname[sstart:send])
	if err != nil {
		return 0, 0, -1, -1,
			ef("Could not parse '%s' as an int: %s", groups[1], err)
	}
	nepisode, err := strconv.Atoi(fname[estart:eend])
	if err != nil {
		return 0, 0, -1, -1,
			ef("Could not parse '%s' as an int: %s", groups[2], err)
	}
	return int(nseason), int(nepisode), start, end, nil
}
