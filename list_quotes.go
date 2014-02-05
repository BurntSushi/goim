package main

import (
	"bytes"
	"io"

	"github.com/BurntSushi/goim/imdb"
)

func listQuotes(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "quote", "atom_id", "entry")
	defer table.done()

	var curAtom imdb.Atom
	var curQuote []byte
	var ok bool
	add := func(line []byte) {
		if curAtom > 0 && len(curQuote) > 0 {
			table.add(line, curAtom, unicode(bytes.TrimSpace(curQuote)))
		}
		curQuote = nil
	}
	listLines(r, func(line []byte) {
		if bytes.HasPrefix(line, []byte{'#'}) {
			add(line)
			entity := bytes.TrimSpace(line[1:])
			if curAtom, ok = table.atoms.AtomOnlyIfExist(entity); !ok {
				warnf("Could not find id for '%s'. Skipping.", entity)
				curAtom, curQuote = 0, nil
			}
			return
		}
		if len(line) == 0 {
			add(line)
			return
		}
		// If the line starts with a space, then it's a continuation.
		// So keep it as one line in the database. We do this by prefixing
		// a new line character whenever we add a new character quote.
		if line[0] != ' ' && len(curQuote) > 0 {
			curQuote = append(curQuote, '\n')
		}
		curQuote = append(curQuote, bytes.TrimSpace(line)...)
		curQuote = append(curQuote, ' ')
	})
	add([]byte("UNKNOWN (last line?)"))
}
