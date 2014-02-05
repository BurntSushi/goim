package main

import (
	"bytes"
	"io"

	"github.com/BurntSushi/goim/imdb"
)

func listRatings(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "rating", "atom_id", "votes", "rank")
	defer table.done()

	done := false
	listLines(r, func(line []byte) {
		var (
			id    imdb.Atom
			ok    bool
			votes int
			rank  float64
		)
		if done {
			return
		}

		fields := bytes.Fields(line)
		if bytes.HasPrefix(line, []byte("REPORT FORMAT")) {
			done = true
			return
		}
		if len(fields) < 4 {
			return
		}
		if bytes.Equal(fields[0], []byte("New")) {
			return
		}

		entity := bytes.Join(fields[3:], []byte{' '})
		if id, ok = table.atoms.AtomOnlyIfExist(entity); !ok {
			warnf("Could not find id for '%s'. Skipping.", entity)
			return
		}
		if err := parseInt(fields[1], &votes); err != nil {
			logf("Could not parse integer '%s' in: '%s'", fields[1], line)
			return
		}
		if err := parseFloat(fields[2], &rank); err != nil {
			logf("Could not parse float '%s' in: '%s'", fields[2], line)
			return
		}
		table.add(line, id, votes, int(10*rank))
	})
}
