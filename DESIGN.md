This document describes the architecture of Goim and some key design decisions 
for the database.


### Overview

The purpose of the main `goim` command (the `main` package) is to provide a 
command line interface for managing a Goim database. This includes searching 
and updating. It also provides some convenience commands like `rename` that use 
IMDb data to help you manage your media.

The logic for searching and querying the database is contained inside the 
`imdb` and `imdb/search` sub-packages. The logic for parsing IMDb lists and 
updating the database is all inside the `main` package. Stated differently, the 
`imdb` and `imdb/search` packages do not contain any logic for adding data to 
the Goim database. All operations that update the database are done in 
transactions so that if they fail, the original data should be preserved.

The `imdb` sub-package contains schema definitions and indices for both SQLite 
and PostgreSQL. Opening a Goim databse through the `imdb` package automatically 
makes sure the schema is up-to-date with the package version.


### Database

The design of the database is very simple. In general, there are three types of 
tables:

  - Tables for storing identifiers/names (e.g., atom, name)
  - Tables for storing entities (e.g., movie, tvshow, episode, actor)
  - Tables for storing attributes about entities (e.g., plot, goof)

All entity and attribute tables are completely wiped whenever they are updated. 
The atom and name tables are not. This can result in stale but benign rows in 
the atom and name tables, but it preserves primary keys across updates.

All rows in all tables (sans migration meta data) are immutable.


### Unique identifiers

The data from IMDb does not include any numeric unique identifiers for each 
movie, TV show, episode, actor, etc. While such identifiers clearly exist on 
their web site, this data is explicitly not included in the "alternative 
interfaces" plain text data dump. Moreover, IMDb's terms of use explicitly
forbids trying to recover this data.

Instead, there appears to be an undocumented guarantee that the name of an 
entity, combined with some miscellaneous attributes, will be unique. For 
example, the full unique string for The Matrix is:

    The Matrix (1999)

This string is then used in other list files to identify this movie uniquely. 
The string may include attributes like "(TV)" or "(V)" for "made for TV" and 
"made for video", respectively.

In the rare case where two movies share the same name and were released in the 
same year, a roman numeral indicates uniqueness:

    The Maury Island Incident (2014/I)
    The Maury Island Incident (2014/II)

These unique strings also apply to TV shows, episodes, actors, etc. For 
example, the full unique string for The Simpsons episode "Lisa the Iconoclast" 
is:

    "The Simpsons" (1989) {Lisa the Iconoclast (#7.16)}

Since these strings seem to be guaranteed to be unique in the source data set, 
Goim also uses these as the ultimate source of a primary key in the database. 
The key is stored as an md5 (binary) hash of the string, which uses 16 bytes. 
This key is then mapped to a unique 32-bit monotonically increasing integer.

This design decision has several effects:

  - There is a possibility of a collision, which will violate a key invariant
    assumed by Goim. (See issue #1.)
  - Mapping each md5 hash to a unique 32 bit integer means that we drastically
    decrease storage requirements, since each row in each attribute table
    requires a uniquely identifying key for an entity.
  - The mapping between md5 hash and 32 bit integer can be reasonably stored in 
    main memory of most systems (requires at least 100MB for 5,000,000 
    entities). This mapping is a crucial piece of inserting data quickly, as it 
    eliminates the need for round-tripping to the database for each attribute 
    row insert.
  - The average length of an IMDb entity name, I believe, requires 40 bytes of 
    storage. All md5 hashes require 16 bytes of storage. So we get some space
    savings. (And this is more important for storing the id<->md5 mapping in 
    memory, rather than with respect to the database.)

One last caveat needs to be mentioned. IMDb frequently pre-loads their database 
with data about media that isn't completely known yet. For example, the 
database might include the first episode of the 26th season of The Simpsons, 
but not yet have an episode name. This will be loaded by Goim. At some point, 
IMDb will update this episode with its name. The old entry will be removed the 
episode table, but it will remain in the atom and name table. The new entry 
will be added to the episode, atom and name tables. Therefore, the atom/name 
table cannot be considered as an oracle of the data in the database (a select 
from atom/name MUST join with an entity table).


### Why is there both an atom and name table? Why not merge them?

Given the current schema, the atom and name tables must be in precise 
correspondence. Given that correspondence, it might make sense to merge them 
together.

However, entities may have alternate names. Therefore, some day, Goim might 
support more than one name for each entity.

