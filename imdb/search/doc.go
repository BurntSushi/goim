/*
Package search provides a convenient interface that can quickly search an IMDb
database loaded with Goim. Each search result corresponds to exactly one entity
in the database, where an entity is (currently) either a movie, a TV show, an
episode or an actor/actress.

The search interface in this package has two forms. One of them is with regular
Go method calls:

	neo := New(db).Text("keanu reeves")
	s := New(db).Text("the matrix").Years(1999, 2003).Cast(neo)

And the other is with a special query string syntax:

	s, err := Query(db, "the matrix {years:1999-2003} {cast:keanu reeves}")

The above two queries are functionally equivalent. (They both return entities
with names similar to "the matrix", released in the years 1999-2003 where Keanu
Reeves was on the cast.) There are more elaborate examples included in this
package. There are even more examples in Goim, which can be seen in the usage
information for the search command.  See 'goim help search'.

Beta

Please consider this package as beta material. I am reasonably happy with what
is here so far and I don't expect to change it. But it hasn't been used much,
so I'd like to wait before declaring it stable.
*/
package search
