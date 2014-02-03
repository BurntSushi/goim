package main

import (
	"bytes"
	"io"

	"github.com/BurntSushi/goim/imdb"
)

func listMPAARatings(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "mpaa_rating", "atom_id", "rating", "reason")
	defer table.done()

	var curAtom imdb.Atom
	var curRating string
	var curReason []byte
	var ok bool
	reset := func() {
		curAtom, curRating, curReason = 0, "", nil
	}
	add := func(line []byte) {
		if len(curReason) > 0 {
			curReason = bytes.TrimSpace(curReason)
			table.add(line, curAtom, curRating, unicode(curReason))
			reset()
		}
	}
	listLines(r, func(line []byte) {
		if len(line) == 0 || line[0] == '-' {
			return
		}
		if bytes.HasPrefix(line, []byte("MV: ")) {
			add(line)
			entity := bytes.TrimSpace(line[3:])
			if curAtom, ok = table.atoms.AtomOnlyIfExist(entity); !ok {
				warnf("Could not find id for '%s'. Skipping.", entity)
				reset()
			}
			return
		}
		if curAtom == 0 || !bytes.HasPrefix(line, []byte("RE: ")) {
			return
		}
		line = bytes.TrimSpace(line[3:])
		if len(curReason) == 0 {
			if bytes.HasPrefix(line, []byte("PG")) {
				// Weird corner case for "The Honeymooners (2005)". Bah.
				line = bytes.TrimSpace(line[2:])
				curRating = "PG"
			} else {
				if !bytes.HasPrefix(line, []byte("Rated ")) &&
					!bytes.HasPrefix(line, []byte("rated ")) {
					logf("Could not find rating in '%s'. Skipping.", line)
					reset()
					return
				}
				line = bytes.TrimSpace(line[5:])
				nextSpace := bytes.IndexByte(line, ' ')
				if nextSpace == -1 {
					curRating = unicode(line)
				} else {
					curRating = unicode(line[:nextSpace])
					line = line[nextSpace+1:]
				}
			}
			switch curRating {
			case "G", "PG", "PG-13", "R", "NC-17": // ok
			default:
				logf("Unrecognized rating '%s' in '%s'. Skipping.",
					curRating, line)
				reset()
			}
		}
		curReason = append(curReason, line...)
		curReason = append(curReason, ' ')
	})
	add([]byte("UNKNOWN (last line?)"))
}
