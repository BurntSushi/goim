goim is a command line utility for maintaining and querying the [Internet Movie 
Database (IMDb)](http://www.imdb.com). Goim automatically downloads IMDb's data 
[in plain text format](http://www.imdb.com/interfaces) and loads it into a 
relational database. Goim can then interact with the data in the database in 
various ways, fuzzy (with trigrams) searching, naming TV series episode files, 
etc.

Goim currently supports both SQLite and PostgreSQL. By default, it uses SQLite. 
For Goim, SQLite is slower and smaller while PostgreSQL is faster and larger.
SQLite is intended to be a convenience, whereas usage with PostgreSQL should be 
fast.
Notably, fuzzy searching with trigrams only works with PostgreSQL.
(SQLite still supports wild card searching.)


Database size
=============
A complete database (with indices) for PostgreSQL appears to take approximately 
5-6GB of space on disk.


Under construction
==================
Goim is currently under construction, but if you install it, you can see a list 
of available commands with `goim help`.

For example, to load all movies/TV shows/episodes into a SQLite file:

    goim write-config
    goim load -lists movies

See `goim help load` for more details.

