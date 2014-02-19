goim is a command line utility for maintaining and querying the [Internet Movie 
Database (IMDb)](http://www.imdb.com). Goim automatically downloads IMDb's data 
[in plain text format](http://www.imdb.com/interfaces) and loads it into a 
relational database. Goim can then interact with the data in the database in 
various ways, fuzzy (with trigrams) searching, naming TV series episode files, 
etc.

Goim currently supports both SQLite and PostgreSQL. By default, it uses SQLite. 
For Goim, SQLite is slower and smaller while PostgreSQL is faster and larger.
SQLite is intended to be a convenience for those that do not want to run a
database server, whereas usage with PostgreSQL should be fast.
In the author's opinion, the biggest difference between using SQLite and
PostgreSQL is the lack of fuzzy searching with trigrams in SQLite.
(SQLite still supports wild card searching.)


### Database loading time and size

The following benchmarks were measured with data downloaded from IMDb on 
February 3, 2014 (872MB compressed). The specs of my machine: Intel i7 3930K 
(12 logical CPUs) with 32GB of DDR3 1600MHz RAM. Both PostgreSQL and SQLite 
databases were stored on a Crucial M4 128GB solid state drive (CT128M4SSD2).

A complete database (with indices) for SQLite uses approximately 3GB 
of space on disk. A complete load (with all IMDb downloaded first) took about 
12 minutes. Note that since this is SQLite, this did not use any concurrent 
updating. After completion, a search query of `%Matrix%` takes approximate 0.5 
seconds.

A complete database (with indices) for PostgreSQL 9.3 (using a default 
configuration) uses approximately 5.5GB of space on disk. A complete load (with 
all IMDb downloaded first) took about 7.5 minutes. There is a significant speed 
boost from parallel table updates, although about half the time is spent 
building indices (the trigram indices take especially long). After completion, 
a search query of `%Matrix%` takes approximately 0.18 seconds. A search query 
of `matrix` (using the trigram indices) takes approximately 1 second. (Searches 
were done only when the Postgres autovacuum appeared to be idling. On my 
system, it tends to run for a few minutes after a full load of the database.)

A PostgreSQL database with just movies/TV shows/episodes takes about 1.5 
minutes to load completely, including indices.

Loading all attribute lists (excludes only `movies` and `actors`/`actresses`)
into a PostgreSQL database takes about 2 minutes to load completely, including
indices.

The point of these benchmarks is not to be rigorous, but to give you a general 
ballpark of the sorts of resources used to load the database.


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
3. I am only using Goim for personal and non-commercial use. I do not make 
the database created by Goim public. Instead, each individual user of Goim has 
to build their own. This has precedent with
[IMDbPY](http://imdbpy.sourceforge.net) and [JMDB](http://www.jmdb.de).
4. Information courtesy of IMDb (http://www.imdb.com). Used with permission.

My interpretation of IMDb's fastidious legalese prevents me from distributing a 
SQL dump of a Goim database (whether it be PostgreSQL or SQLite). This is 
unfortunate, because it would likely be much faster to load than downloading 
and inserting each individual IMDb list file.

