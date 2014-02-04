package main

import (
	"io"

	"github.com/BurntSushi/goim/imdb"
)

func listSoundMixes(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "sound_mix",
		"atom_id", "mix", "attrs")
	defer table.done()

	listAttrRows(r, table.atoms, func(id imdb.Atom, line, entity, row []byte) {
		var attrs []byte

		rowFields := splitListLine(row)
		if len(rowFields) == 0 {
			return // herp derp...
		}
		if len(rowFields) > 1 {
			attrs = rowFields[1]
		}
		table.add(line, id, unicode(rowFields[0]), unicode(attrs))
	})
}

func listGenres(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "genre", "atom_id", "name")
	defer table.done()

	listAttrRows(r, table.atoms, func(id imdb.Atom, line, entity, row []byte) {
		fields := splitListLine(row)
		if len(fields) == 0 {
			return // herp derp...
		}
		table.add(line, id, unicode(fields[0]))
	})
}

func listTrivia(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "trivia", "atom_id", "entry")
	defer table.done()

	listPrefixItems(r, table.atoms, '-', func(id imdb.Atom, item []byte) {
		table.add(item, id, unicode(item))
	})
}

func listAlternateVersions(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "alternate_version", "atom_id", "about")
	defer table.done()

	listPrefixItems(r, table.atoms, '-', func(id imdb.Atom, item []byte) {
		table.add(item, id, unicode(item))
	})
}

func listTaglines(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "tagline", "atom_id", "tag")
	defer table.done()

	listPrefixItems(r, table.atoms, '\t', func(id imdb.Atom, item []byte) {
		table.add(item, id, unicode(item))
	})
}
