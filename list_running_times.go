package main

import (
	"bytes"
	"io"
	"strconv"

	"github.com/BurntSushi/goim/imdb"
)

func listRunningTimes(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "running_time",
		"atom_id", "country", "minutes", "attrs")
	defer table.done()

	listAttrRows(r, table.atoms, func(id imdb.Atom, line, entity, row []byte) {
		var (
			country string
			minutes int
			attrs   []byte
		)

		rowFields := splitListLine(row)
		if len(rowFields) == 0 {
			return // herp derp...
		}
		if !parseRunningTime(rowFields[0], &country, &minutes) {
			return
		}
		if len(rowFields) > 1 {
			attrs = rowFields[1]
		}
		table.add(line, id, country, minutes, unicode(attrs))
	})
}

func parseRunningTime(text []byte, country *string, minutes *int) bool {
	sep := bytes.IndexByte(text, ':')
	var runtime []byte
	if sep > -1 {
		*country = unicode(bytes.TrimSpace(text[:sep]))
		runtime = bytes.TrimSpace(text[sep+1:])
	} else {
		*country = ""
		runtime = bytes.TrimSpace(text)
	}

	var err error
	*minutes, err = strconv.Atoi(unicode(runtime))
	if err != nil {
		// There are a lot of these.
		// From the looks of it, IMDb's web site just ignores them.
		// It's almost like it's freeform text... Yikes.
		return false
	}
	return true
}
