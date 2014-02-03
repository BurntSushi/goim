package main

import (
	"bytes"
	"io"
	"time"

	"github.com/BurntSushi/goim/imdb"
)

func listReleaseDates(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "release_date",
		"atom_id", "country", "released", "attrs")
	defer table.done()

	listAttrRows(r, table.atoms, func(id imdb.Atom, line, entity, row []byte) {
		var (
			country string
			date    time.Time
			attrs   string
		)

		rowFields := splitListLine(row)
		if !parseReleaseDate(rowFields[0], &country, &date) {
			pef("Could not extract date from '%s'. Skipping.", line)
			return
		}
		if len(rowFields) > 1 {
			attrs = unicode(rowFields[1])
		}
		table.add(line, id, country, date, attrs)
	})
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
