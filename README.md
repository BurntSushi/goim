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


Database loading time and size
==============================
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


Under construction
==================
Goim is currently under construction, but if you install it, you can see a list 
of available commands with `goim help`.

For example, to load all movies/TV shows/episodes into a SQLite file:

    goim write-config
    goim load -lists movies

See `goim help load` for more details.

