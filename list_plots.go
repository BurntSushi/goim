package main

import (
	"bytes"
	"io"

	"github.com/BurntSushi/goim/imdb"
)

func listPlots(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "plot", "atom_id", "entry", "by")
	defer table.done()

	var curAtom imdb.Atom
	var curPlot []byte
	var curBy []byte
	var ok bool
	add := func(line []byte) {
		if curAtom > 0 && len(curPlot) > 0 {
			plot := unicode(bytes.TrimSpace(curPlot))
			by := unicode(bytes.TrimSpace(curBy))
			table.add(line, curAtom, plot, by)
		}
		curPlot, curBy = nil, nil
	}
	listLines(r, func(line []byte) {
		if bytes.HasPrefix(line, []byte("MV:")) {
			if len(curPlot) > 0 {
				add(line)
			}
			entity := bytes.TrimSpace(line[3:])
			if curAtom, ok = table.atoms.AtomOnlyIfExist(entity); !ok {
				warnf("Could not find id for '%s'. Skipping.", entity)
				curAtom, curPlot, curBy = 0, nil, nil
			}
			return
		}
		if len(line) == 0 {
			return
		}
		if bytes.HasPrefix(line, []byte("PL:")) {
			curPlot = append(curPlot, bytes.TrimSpace(line[3:])...)
			curPlot = append(curPlot, ' ')
			return
		}
		if bytes.HasPrefix(line, []byte("BY:")) {
			curBy = line[3:]
			add(line)
			return
		}
	})
	add([]byte("UNKNOWN (last line?)"))
}
