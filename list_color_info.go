package main

import (
	"bytes"
	"io"

	"github.com/BurntSushi/goim/imdb"
)

func listColorInfo(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "color_info",
		"atom_id", "color", "attrs")
	defer table.done()

	listAttrRows(r, table.atoms, func(id imdb.Atom, line, entity, row []byte) {
		var (
			color bool
			attrs []byte
		)

		rowFields := splitListLine(row)
		if len(rowFields) == 0 {
			return // herp derp...
		}
		if !parseColorInfo(rowFields[0], &color) {
			return
		}
		if len(rowFields) > 1 {
			attrs = rowFields[1]
		}
		table.add(line, id, color, unicode(attrs))
	})
}

var (
	infoColor = []byte("Color")
	infoBandW = []byte("Black and White")
)

func parseColorInfo(text []byte, color *bool) bool {
	switch {
	case bytes.Equal(text, infoColor):
		*color = true
		return true
	case bytes.Equal(text, infoBandW):
		*color = false
		return true
	}
	logf("Could not parse '%s' as color information.", text)
	return false
}
