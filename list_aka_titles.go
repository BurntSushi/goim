package main

import (
	"bytes"
	"io"

	"github.com/BurntSushi/goim/imdb"
)

func listAkaTitles(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "aka_title", "atom_id", "title", "attrs")
	defer table.done()

	listAttrRows(r, table.atoms, func(id imdb.Atom, line, entity, row []byte) {
		var (
			title string
			attrs []byte
		)

		rowFields := splitListLine(row)
		if len(rowFields) == 0 {
			return // herp derp...
		}
		if !parseAkaTitle(rowFields[0], &title) {
			if !bytes.Contains(rowFields[0], []byte("(VG)")) {
				logf("Could not parse aka title from '%s'", rowFields[0])
			}
			return
		}
		if len(rowFields) > 1 {
			attrs = rowFields[1]
		}
		table.add(line, id, title, unicode(attrs))
	})
}

func parseAkaTitle(text []byte, title *string) bool {
	attrName, data, ok := parseNamedAttr(text)
	if !ok {
		return false
	}
	if !bytes.Equal(attrName, []byte("aka")) {
		return false
	}
	ent, ok := parseMediaEntity(data)
	if !ok {
		return false
	}
	*title = ent.Name()
	return true
}
