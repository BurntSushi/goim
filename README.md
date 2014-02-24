Goim is a command line utility for maintaining and querying the [Internet Movie 
Database (IMDb)](http://www.imdb.com). Goim automatically downloads IMDb's data 
[in plain text format](http://www.imdb.com/interfaces) and loads it into a 
relational database. Goim can then interact with the data in the database in 
various ways: fuzzy (with trigrams) searching, simple renaming of media files 
(including TV episodes), view information like plots, credits, goofs, quotes, 
IMDb rankings, trivia, release dates, film locations, prequels/sequels, etc.

Goim currently supports both SQLite and PostgreSQL. By default, Goim uses 
SQLite---which is more of a convenience for users that don't want to run a 
database server. Using PostgreSQL should be faster, and more importantly, will 
give you insanely fast fuzzy searching.

For Go programmers, the 
[`imdb`](http://godoc.org/github.com/BurntSushi/goim/imdb)
sub-package contains types and functions for handling data in the database. The
[`imdb/search`](http://godoc.org/github.com/BurntSushi/goim/imdb/search)
sub-package exposes the full power and flexibility of Goim's searching via an 
API.

Goim is relased under the [UNLICENSE](http://unlicense.org).


### Installation

Goim depends on Go and is go-gettable. Assuming you have Go installed and your 
[GOPATH](http://golang.org/doc/code.html#GOPATH) is set, then the following 
will install Goim into `$GOPATH/bin`:

    go get github.com/BurntSushi/goim

By default, this will attempt to install SQLite. If you don't want SQLite or 
can't install it easily, then install Goim with CGO disabled:

    CGO_ENABLED=0 go get github.com/BurntSushi/goim

When CGO is disabled, Goim will only work with PostgreSQL.


### Quickstart with SQLite

If you want to give Goim a quick spin, it's easy to create a SQLite database 
with a subset of IMDb's data:

    goim load -db goim.sqlite

This command downloads a list of all movies, TV shows and episodes and creates 
a new SQLite database in goim.sqlite. Depending on your system and internet 
connection, this might take anywhere from 1 minute to 5 minutes (including 
building indices).

Now you can find all episodes of The Simpsons that have "maggie" in the title:

    # goim search -db goim.sqlite '%maggie%' {show:the simpsons}
      1. episode  And Maggie Makes Three (1995) (TV show: The Simpsons, #6.13)
      2. episode  Gone Maggie Gone (2009) (TV show: The Simpsons, #20.13)

If you add IMDb user rankings (should take less than a minute):

    goim load -db goim.sqlite -lists ratings

Then you can find the top ten ranked Simpsons episodes with at least 500 votes:

    # time goim search -db goim.sqlite {show:the simpsons} {votes:500-} {sort:rank desc} {limit:10}
      1. episode  Homer the Smithers (1996) (TV show: The Simpsons, #7.17) (rank: 90/100, votes: 840)
      2. episode  Homer's Enemy (1997) (TV show: The Simpsons, #8.23) (rank: 89/100, votes: 1217)
      3. episode  The City of New York vs. Homer Simpson (1997) (TV show: The Simpsons, #9.1) (rank: 89/100, votes: 1160)
      4. episode  Boy Scoutz 'n the Hood (1993) (TV show: The Simpsons, #5.8) (rank: 88/100, votes: 874)
      5. episode  Homer Badman (1994) (TV show: The Simpsons, #6.9) (rank: 88/100, votes: 960)
      6. episode  Homer the Heretic (1992) (TV show: The Simpsons, #4.3) (rank: 88/100, votes: 1090)
      7. episode  Homer's Phobia (1997) (TV show: The Simpsons, #8.15) (rank: 88/100, votes: 1031)
      8. episode  Homer's Triple Bypass (1992) (TV show: The Simpsons, #4.11) (rank: 88/100, votes: 895)
      9. episode  Hurricane Neddy (1996) (TV show: The Simpsons, #8.8) (rank: 88/100, votes: 855)
     10. episode  King Size Homer (1995) (TV show: The Simpsons, #7.7) (rank: 88/100, votes: 997)

Dig deeper by adding plot information to your database (takes minutes):

    goim load -db goim.sqlite -lists plot

And check out the plot for King Size Homer:

    # goim plots -db goim.sqlite king size homer
    
    Plot summaries for King Size Homer (1995)
    =========================================
    Mr. Burns institutes a new calisthenics program at work. Most employees enjoy
    the morning workout, except Homer, who is too lazy. He finds out that if he
    goes in disability, he will be exempt from the exercises. He finds hyper-obesity
    among the list of disability, so he gorges himself on food to balloon up to 300
    pounds.
    -- Anonymous

You can read more examples and see a complete list of search options by running
`goim help search`. For example, if you load the `actors` list, you can search 
the credits of movies and episodes.

Also, see `goim help` for a list of all commands, which includes a command for
each type of information available.


### Upping the ante with PostgreSQL

You will need to install a PostgreSQL server and have it running on your 
machine. This can be done for Windows, Mac or Linux. [Start 
here](https://wiki.postgresql.org/wiki/Detailed_installation_guides).

Once you're all set up, create a database and enable the `pg_trgm` extension 
(which is what provides fuzzy searching):

    createdb imdb
    psql -U postgres imdb -c 'CREATE EXTENSION pg_trgm;'

Note that enabling an extension can only be done by a PostgreSQL superuser, 
which is what the '-U postgres' is for (you may use any user here that has 
superuser privileges).

Technically, you can use Goim with PostgreSQL without enabling the `pg_trgm` 
extension, but it isn't recommended (and Goim will yell at you).

Now all you need to do is fill in your connection information. You can use the 
`-db` flag, but typing in all your connection details every time is painful. 
Instead, tell Goim to write a default config file:

    goim write config

Now edit and fill in your details (the comments in the config file should 
help):

    $EDITOR ~/.config/goim/config.toml

Note that the config file can specify a SQLite database too.

With all of that out of the way, you can now follow the steps above for loading 
and searching with SQLite. (Leave out the `-db ...` flag.) Also, with fuzzy 
searching, you don't need to use the '%' wildcard any more (although you can).
For example, you can use `goim search maggie {show:simpsons}` to find all 
episodes of The Simpsons with "maggie" in the title.


### Renaming media files

I just copied the first season of The Simpsons off my DVD box set, but I have a 
problem. All of my files look like this:

    S01E01.mkv  S01E04.mkv  S01E07.mkv  S01E10.mkv  S01E13.mkv
    S01E02.mkv  S01E05.mkv  S01E08.mkv  S01E11.mkv
    S01E03.mkv  S01E06.mkv  S01E09.mkv  S01E12.mkv

No problem. Goim can rename these easily with the `rename` command:

    # goim rename -tv 'the simpsons' *.mkv
    Rename 'S01E01.mkv' to 'S01E01 - Simpsons Roasting on an Open F
    Rename 'S01E02.mkv' to 'S01E02 - Bart the Genius.mkv'
    Rename 'S01E03.mkv' to 'S01E03 - Homer's Odyssey.mkv'
    Rename 'S01E04.mkv' to 'S01E04 - There's No Disgrace Like Home.
    Rename 'S01E05.mkv' to 'S01E05 - Bart the General.mkv'
    Rename 'S01E06.mkv' to 'S01E06 - Moaning Lisa.mkv'
    Rename 'S01E07.mkv' to 'S01E07 - The Call of the Simpsons.mkv'
    Rename 'S01E08.mkv' to 'S01E08 - The Telltale Head.mkv'
    Rename 'S01E09.mkv' to 'S01E09 - Life on the Fast Lane.mkv'
    Rename 'S01E10.mkv' to 'S01E10 - Homer's Night Out.mkv'
    Rename 'S01E11.mkv' to 'S01E11 - The Crepes of Wrath.mkv'
    Rename 'S01E12.mkv' to 'S01E12 - Krusty Gets Busted.mkv'
    Rename 'S01E13.mkv' to 'S01E13 - Some Enchanted Evening.mkv'
    Are you sure you want to rename these files? [y/n]: y

And now my files look like this:

    S01E01 - Simpsons Roasting on an Open Fire.mkv                 
    S01E02 - Bart the Genius.mkv                                   
    S01E03 - Homer's Odyssey.mkv
    S01E04 - There's No Disgrace Like Home.mkv
    S01E05 - Bart the General.mkv
    S01E06 - Moaning Lisa.mkv
    S01E07 - The Call of the Simpsons.mkv
    S01E08 - The Telltale Head.mkv
    S01E09 - Life on the Fast Lane.mkv
    S01E10 - Homer's Night Out.mkv
    S01E11 - The Crepes of Wrath.mkv
    S01E12 - Krusty Gets Busted.mkv
    S01E13 - Some Enchanted Evening.mkv

The above command executes in less than a second on my machine. The exact same 
command could be used to *rename an entire series at once*.

The rename command is very flexible, and it can also rename movies and work 
with different file name formats. Read more about it with `goim help rename`.


### Updating the database

Whether you're loading data for the first time or updating an existing 
database, you'll want to use Goim's `load` command. By default, data is 
downloaded from one of IMDb's FTP mirrors, but it also supports HTTP 
downloading or reading from the local file system.

The `load` command lets you pick and choose which lists you want. By default, 
it *only* loads the `movies` list. But let's say you also want plots and 
quotes:

    goim load -lists plot,quotes

Since plots and quotes are completely independent, this load will be done in 
parallel if you're using PostgreSQL.

If you want to add all attribute information (i.e., plots, quotes, trivia,
goofs, etc.), then you can use the special `attr` list (make sure `movies` has
already been loaded):

    goim load -lists attr

Or you can load all information available with the `all` list. (Warning: 
loading actors can take a while!)

I haven't been clever enough to come up with a good way for updating the 
database in place, so every update will truncate the corresponding table and 
rebuild it from scratch. (This is done inside a transaction, so if something 
bad happens, your old data should be preserved.) The **only exceptions** to 
the truncating scheme are the `atom` and `name` table. The short story here is 
that this will allow primary (surrogate) keys to persist across updates. Under 
this scheme, you should never have to worry about stale data cluterring search 
results.

Typically, IMDb updates its plain text data sets some time between Friday and 
Saturday morning, so there's no need to have Goim update your database more 
frequently than once a week.


### Entity-Relationship diagram

The schema of the database is very simple, but I've made an
[ER diagram](http://burntsushi.net/stuff/goim/goim.pdf).
It was automatically generated with [erd](https://github.com/BurntSushi/erd)
and
[goim-write-erd](https://github.com/BurntSushi/goim/blob/master/scripts/goim-write-erd).


### Database loading time and size

The following benchmarks were measured with data downloaded from IMDb on 
February 3, 2014 (872MB compressed). The specs of my machine: Intel i7 3930K 
(12 logical CPUs) with 32GB of DDR3 1600MHz RAM. Both PostgreSQL and SQLite 
databases were stored on a Crucial M4 128GB solid state drive (CT128M4SSD2).

A complete database (with indices) for SQLite uses approximately 3GB 
of space on disk. A complete load (with all IMDb data downloaded first) took 
about 12 minutes. Note that since this is SQLite, this did not use any 
concurrent updating. After completion, a search query of `%Matrix%` takes 
approximately 0.5 seconds.

A complete database (with indices) for PostgreSQL 9.3 (using a default 
configuration) uses approximately 5.5GB of space on disk. A complete load (with 
all IMDb data downloaded first) took about 7.5 minutes. There is a significant 
speed boost from parallel table updates, although about half the time is spent 
building indices (the trigram indices take especially long). After completion, 
a search query of `%Matrix%` takes approximately 0.18 seconds. A search query 
of `matrix` (using the trigram indices) takes approximately 1 second. (Searches 
were done only when the Postgres autovacuum appeared to be idling. On my 
system, it tends to run for a few minutes after a full load of the database.)

Goim is smart about updating and will avoid rebuilding indices where 
appropriate. For example, while the initial load took 7.5 minutes, updating the 
database with new data took only 5.5 minutes on my machine.

A PostgreSQL database with just movies/TV shows/episodes takes about 1.5 
minutes to load completely, including indices.

Loading all attribute lists (excludes only `movies` and `actors`/`actresses`)
into a PostgreSQL database takes about 2 minutes to load completely, including
indices.

The point of these benchmarks is not to be rigorous, but to give you a general 
ballpark of the sorts of resources used to load the database.


### TODO

* Goim doesn't currently support all available lists. Notable absences are 
  biographies, soundtracks, directors, writers, producers (and other crew 
  members).
* I am pleased with the search infrastructure, but there needs to be more
  options. For example, to search movie links, running times, release dates, 
  genres, etc.
* Look into searching plots/quotes. (Concern: how long will a fulltext index 
  take to build on these tables?)
* Expand the clever logic in the `rename` command to general searching.
  (Maybe. Not sure if I want to complicate the search too much more, but it
  would be nice to give a file name and, e.g., get back a plot.)
* With Goim's data and searching, can we easily connect it to other data 
  sources? (Schedules? Subtitles?)
* Investigate adding foreign key constraints to the schema.


### Motivation and comparison with similar tools

I tend to acquire a lot of media and it's a pain to keep up with correctly 
naming it. Many years ago, I spent a weekend hacking together a Python script 
to parse the `movies` IMDb list into a MySQL database and used that to rename 
files. But it was slow and the renaming script was terribly inflexible.
I wanted to make it better, so I embarked on a more disciplined approach to 
storing IMDb's data. I also find it incredibly useful to access most of IMDb 
instantly from the command line.

To the best of my knowledge, there are only two tools that claim to load a 
substantial fraction of IMDb's data into a relational database:
[IMDbPY](http://imdbpy.sourceforge.net) and
[JMDB](http://www.jmdb.de).
The source code of JMDB doesn't appear to ever have been released, and it looks 
like some sort of GUI tool. Truthfully, I haven't tried it.

IMDbPY has been around for a long time and is pretty similar to Goim. However, 
I found its loading procedure to be a bit awkward (a fast load seems to require 
some mangling with CSV files), and generally slower than Goim although I 
haven't done any rigorous benchmarks. (And I don't know enough about IMDbPY to 
know if the comparison would be fair.)

IMDbPY also seems to support MySQL. Goim does not. (And I don't have any 
particular plans to support it, but I'm not against it.)

It is entirely possible that I could have used IMDbPY to load a database and 
then built tools on top of it to do searching and renaming, but I'm much 
happier with a smaller and simpler piece of software to do the work for me.
Also, it's a lot more fun to design your own database.

### Licensing minutia

While IMDb is generous enough to provide an easily parseable dump of a subset 
of their data, they are pretty finicky with
[their licensing](http://www.imdb.com/help/show_leaf?usedatasoftware).

This project is not a commerical project. The **only source** of IMDb data in 
Goim is through the ["alternative interfaces" plain text data 
files](http://www.imdb.com/interfaces), which are expressly provided for 
non-commercial uses.

Point-by-point:

1. I agree to the terms of their
[copyright/terms of use](http://www.imdb.com/help/show_article?conditions). 
Namely, I am not using data mining, robots, screen scraping or any other 
mechanism to get IMDb data other than the aforementioned "alternative 
interfaces" plain text data dump. To the best of my knowledge, I am not using 
any framing techniques to enclose IMDb trademarks, logos or other proprietary 
information. I do not link to IMDb in any part of Goim, sans this README.
I am not using any IMDb software (not that it actually works), so the terms at 
the bottom don't apply.
2. As mentioned above, data is only taken from the plain text data from their 
"alternative interfaces," specifically from one of those listed FTP sites. Goim 
does not send any HTTP requests to IMDb proper. Goim does not attempt to 
recover any information on IMDb proper that is not available in the subset of 
data provided through the "alternative interface."
3. I am only using Goim for personal and non-commercial use. Each individual 
user of Goim has to build their own database. This has precedent with
[IMDbPY](http://imdbpy.sourceforge.net) and [JMDB](http://www.jmdb.de).
4. Information courtesy of IMDb (http://www.imdb.com). Used with permission.

