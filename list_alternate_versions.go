package main

import (
	"bytes"
	"io"

	"github.com/BurntSushi/goim/imdb"
)

func listAlternateVersions(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "alternate_version", "atom_id", "about")
	defer table.done()

	var curAtom imdb.Atom
	var curAbout []byte
	var ok bool
	add := func(line []byte) {
		if len(curAbout) > 0 {
			curAbout = bytes.TrimSpace(curAbout)
			table.add(line, curAtom, unicode(curAbout))
			curAbout = nil
		}
	}
	listLines(r, func(line []byte) {
		if len(line) == 0 {
			return
		}
		if line[0] == '#' {
			add(line)
			entity := bytes.TrimSpace(line[1:])
			if curAtom, ok = table.atoms.AtomOnlyIfExist(entity); !ok {
				warnf("Could not find id for '%s'. Skipping.", entity)
				curAtom, curAbout = 0, nil
			}
			return
		}
		if curAtom == 0 {
			return
		}
		if line[0] == '-' {
			if len(curAbout) > 0 {
				add(line)
			}
			line = line[1:]
		}
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			return
		}
		curAbout = append(curAbout, line...)
		curAbout = append(curAbout, ' ')
	})
	add([]byte("UNKNOWN (last line?)"))
}
