/*
Package imdb provides types and functions for retrieving data from an IMDb
database loaded by Goim. There are types for each entity and attribute, along
with some convenience functions for loading them from the database. While there
are a lot of types---since the database is large---this package is actually
fairly minimal. It is likely that you'll find the 'search' sub-package more
useful.

The database can be queried using the 'database/sql' package, but it is
strongly recommended that you use the Open function in this package (which will
give you access to a *sql.DB value). Namely, the Open function will perform a
migration on the schema of your database to make sure it is up to date with the
version of the 'imdb' package that you're using. (If the migration fails, it
will be rolled back and your database will be left untouched.)

Also, many of the functions here require values with types in my csql package:
https://github.com/BurntSushi/csql. Mostly, these types are interfaces that
types in the 'database/sql' package already satisfy. For example, a
csql.Queryer can be a *imdb.DB, *sql.DB, *sql.Stmt or a *sql.Tx, etc.
*/
package imdb
