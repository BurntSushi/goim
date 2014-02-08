package main

import (
	"bytes"
	"io"

	"github.com/BurntSushi/goim/imdb"
)

func listSoundMixes(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "sound_mix",
		"atom_id", "mix", "attrs")
	defer table.done()

	listAttrRowIds(r, table.atoms, func(id imdb.Atom, line, ent, row []byte) {
		var attrs []byte

		fields := splitListLine(row)
		if len(fields) == 0 {
			return
		}
		if len(fields) > 1 {
			attrs = fields[1]
		}
		table.add(line, id, unicode(fields[0]), unicode(attrs))
	})
}

func listGenres(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "genre", "atom_id", "name")
	defer table.done()

	listAttrRowIds(r, table.atoms, func(id imdb.Atom, line, ent, row []byte) {
		fields := splitListLine(row)
		if len(fields) == 0 {
			return
		}
		table.add(line, id, unicode(fields[0]))
	})
}

func listLanguages(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "language", "atom_id", "name", "attrs")
	defer table.done()

	listAttrRowIds(r, table.atoms, func(id imdb.Atom, line, ent, row []byte) {
		var attrs []byte
		fields := splitListLine(row)
		if len(fields) == 0 {
			return
		}
		if len(fields) > 1 {
			attrs = fields[1]
		}
		table.add(line, id, unicode(fields[0]), unicode(attrs))
	})
}

func listLocations(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "location", "atom_id", "place", "attrs")
	defer table.done()

	listAttrRowIds(r, table.atoms, func(id imdb.Atom, line, ent, row []byte) {
		var attrs []byte
		fields := splitListLine(row)
		if len(fields) == 0 {
			return
		}
		if len(fields) > 1 {
			attrs = fields[1]
		}
		table.add(line, id, unicode(fields[0]), unicode(attrs))
	})
}

func listTrivia(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "trivia", "atom_id", "entry")
	defer table.done()

	do := func(id imdb.Atom, item []byte) {
		table.add(item, id, unicode(item))
	}
	listPrefixItems(r, table.atoms, []byte{'#'}, []byte{'-'}, do)
}

func listAlternateVersions(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "alternate_version", "atom_id", "about")
	defer table.done()

	do := func(id imdb.Atom, item []byte) {
		table.add(item, id, unicode(item))
	}
	listPrefixItems(r, table.atoms, []byte{'#'}, []byte{'-'}, do)
}

func listTaglines(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "tagline", "atom_id", "tag")
	defer table.done()

	do := func(id imdb.Atom, item []byte) {
		table.add(item, id, unicode(item))
	}
	listPrefixItems(r, table.atoms, []byte{'#'}, []byte{'\t'}, do)
}

func listGoofs(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "goof", "atom_id", "goof_type", "entry")
	defer table.done()

	do := func(id imdb.Atom, item []byte) {
		sep := bytes.IndexByte(item, ':')
		if sep == -1 {
			table.add(item, id, "", unicode(item))
			return
		}
		goofType := bytes.TrimSpace(item[0:sep])
		item = bytes.TrimSpace(item[sep+1:])
		table.add(item, id, unicode(goofType), unicode(item))
	}
	listPrefixItems(r, table.atoms, []byte{'#'}, []byte{'-'}, do)
}

func listLiterature(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "literature", "atom_id", "lit_type", "ref")
	defer table.done()

	do := func(id imdb.Atom, item []byte) {
		sep := bytes.IndexByte(item, ':')
		if sep == -1 {
			logf("Badly formatted literature reference (skipping): '%s'", item)
			return
		}
		litType := bytes.TrimSpace(item[0:sep])
		item = bytes.TrimSpace(item[sep+1:])
		table.add(item, id, unicode(litType), unicode(item))
	}
	listPrefixItems(r, table.atoms, []byte("MOVI:"), nil, do)
}
