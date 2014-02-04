package main

import (
	"bytes"
	"io"

	"github.com/BurntSushi/goim/imdb"
)

func listMovieLinks(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "link", "atom_id",
		"link_type", "link_atom_id", "entity")
	defer table.done()

	listAttrRows(r, table.atoms, func(id imdb.Atom, line, entity, row []byte) {
		var (
			linkType   string
			linkAtom   imdb.Atom
			linkEntity imdb.EntityKind
		)

		fields := splitListLine(row)
		if len(fields) == 0 {
			return
		}
		if bytes.Contains(fields[0], []byte("(VG)")) {
			return
		}
		ok := parseMovieLink(table.atoms, fields[0],
			&linkType, &linkAtom, &linkEntity)
		if !ok {
			return
		}
		table.add(line, id, linkType, linkAtom, linkEntity.String())
	})
}

func parseMovieLink(
	atoms *imdb.Atomizer,
	text []byte,
	linkType *string,
	linkAtom *imdb.Atom,
	linkEntity *imdb.EntityKind,
) bool {
	attrName, data, ok := parseNamedAttr(text)
	if !ok {
		logf("Could not parse named attribute '%s'. Skipping.", text)
		return false
	}
	id, ok := atoms.AtomOnlyIfExist(data)
	if !ok {
		warnf("Could not find id for '%s'. Skipping.", data)
		return false
	}
	ent, ok := parseMediaEntity(data)
	if !ok {
		logf("Could not find entity type for '%s'. Skipping.", data)
		return false
	}
	*linkType = unicode(attrName)
	*linkAtom = id
	*linkEntity = ent.Type()
	return true
}
