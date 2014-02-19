package search

import (
	"log"

	"github.com/BurntSushi/goim/imdb"
)

// Example New finds the top 10 ranked Simpsons episodes with 500 or more
// votes using methods on a Searcher value.
func ExampleNew() {
	var db *imdb.DB // needs to be created with imdb.Open

	s := New(db)
	s.Votes(500, -1).Sort("rank", "desc").Sort("votes", "desc").Limit(10)
	s.Tvshow(New(db).Text("the simpsons"))

	results, err := s.Results()
	if err != nil {
		log.Fatal(err)
	}
	for _, result := range results {
		log.Println(result)
	}
}


// Example Query finds the top 10 ranked Simpsons episodes with 500 or more
// votes using a query string.
func ExampleQuery() {
	var db *imdb.DB // needs to be created with imdb.Open

	q := `
		{show:the simpsons} {votes:500-}
		{sort:rank desc} {sort:votes desc} {limit:10}
	`
	s, err := Query(db, q)
	if err != nil {
		log.Fatal(err)
	}

	results, err := s.Results()
	if err != nil {
		log.Fatal(err)
	}
	for _, result := range results {
		log.Println(result)
	}
}

