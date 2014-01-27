package main

import (
	"bytes"
	"io"
	"time"

	"github.com/BurntSushi/csql"
	"github.com/BurntSushi/goim/imdb"
)

func listReleases(db *imdb.DB, releases io.ReadCloser) {
	logf("Reading release dates list...")
	addedDates := 0

	// It's easier to just blow away the dates table and reconstruct it.
	csql.SQLPanic(csql.Truncate(db, db.Driver, "release"))

	txDates, err := db.Begin()
	csql.SQLPanic(err)

	dateIns, err := db.NewInserter(txDates, 50, "release",
		"atom_id", "outlet", "country", "released", "attrs")

	atoms, err := db.NewAtomizer(nil)
	csql.SQLPanic(err)

	insert := func(line []byte, id imdb.Atom, o, c, a string, date time.Time) {
		if err := dateIns.Exec(id, o, c, date, a); err != nil {
			logf("Full release date info (that failed to add): "+
				"id:%d, outlet:%s, country:%s, date:%s, attrs:'%s'",
				id, o, c, date, a)
			csql.SQLPanic(ef("Error adding date '%s': %s", line, err))
		}
	}
	listLines(releases, func(line []byte) bool {
		var (
			id      imdb.Atom
			ok      bool
			country string
			date    time.Time
			attrs []byte
		)

		fields := splitListLine(line)
		item, value := fields[0], fields[1]
		if len(fields) == 3 {
			attrs = bytes.TrimSpace(fields[2])
		}
		if id, ok = atoms.AtomOnlyIfExist(item); !ok {
			logf("Could not find id for '%s'. Skipping.", item)
			return true
		}
		if !parseReleaseDate(value, &country, &date) {
			pef("Could not extract date from '%s'. Skipping.", line)
			return true
		}

		ent := entityType("release-dates", item)
		insert(line, id, ent.String(), country, unicode(attrs), date)
		addedDates++
		return true
	})
	logf("Done. Added %d release dates.", addedDates)
}

func parseReleaseDate(text []byte, country *string, released *time.Time) bool {
	sep := bytes.IndexByte(text, ':')
	var date []byte
	if sep > -1 {
		*country = unicode(bytes.TrimSpace(text[:sep]))
		date = bytes.TrimSpace(text[sep+1:])
	} else {
		*country = ""
		date = bytes.TrimSpace(text)
	}

	var layout string
	switch spaces := len(bytes.Fields(date)); spaces {
	case 3:
		layout = "2 January 2006"
	case 2:
		layout = "January 2006"
	case 1:
		layout = "2006"
	default:
		pef("Too many elements in date '%s' (%d) in '%s'", date, spaces, text)
		return false
	}

	t, err := time.Parse(layout, unicode(date))
	if err != nil {
		pef("Could not parse date '%s': %s", date, err)
		return false
	}
	*released = t.UTC()
	return true
}
