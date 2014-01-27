goim is a command line utility for maintaining and querying the [Internet Movie 
Database (IMDb)](http://www.imdb.com). Goim automatically downloads IMDb's data 
[in plain text format](http://www.imdb.com/interfaces) and loads it into a 
relational database. Goim can then interact with the data in the database in 
various ways, including general purpose searching, fuzzy matching, naming 
TV series episode files, etc.

Goim currently supports both SQLite and PostgreSQL. By default, it uses SQLite. 
For Goim, SQLite is slower and smaller while PostgreSQL is faster and larger.
SQLite is intended to be a convenience, whereas usage with PostgreSQL should be 
fast.


Under construction
==================
Goim is currently under construction, but if you install it, you can see a list 
of available commands with `goim help`.

For example, to load all movies/TV shows/episodes into a SQLite file:

    goim write-config
    goim load -lists movies

See `goim help load` for more details.

