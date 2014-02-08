package main

import (
	"bytes"
	"io"

	"github.com/BurntSushi/csql"
	"github.com/BurntSushi/goim/imdb"
)

func listActors(db *imdb.DB, ractor, ractress io.ReadCloser) {
	defer idxs(db, "atom", "name", "actor", "credit").drop().create()
	defer func() { csql.Panic(db.CloseInserters()) }()

	logf("Reading actors list...")

	// Postgresql wants different transactions for each inserter.
	// SQLite can't handle them.
	txactor, err := db.Begin()
	csql.Panic(err)
	txcredit, err := txactor.Another()
	csql.Panic(err)
	txname, err := txactor.Another()
	csql.Panic(err)
	txatom, err := txactor.Another()
	csql.Panic(err)

	// We don't refresh the actor table, but we do need to rebuild credits.
	csql.Panic(csql.Truncate(txcredit, db.Driver, "credit"))

	batchSize := 50
	actIns, err := db.NewInserter(txactor, batchSize, "actor",
		"atom_id", "sequence")
	csql.Panic(err)
	credIns, err := db.NewInserter(txcredit, batchSize, "credit",
		"actor_atom_id", "media_atom_id", "character", "position", "attrs")
	csql.Panic(err)
	nameIns, err := db.NewInserter(txname, batchSize, "name",
		"atom_id", "name")
	csql.Panic(err)
	atoms, err := db.NewAtomizer(txatom)
	csql.Panic(err)

	var nacts1, ncreds1, nacts2, ncreds2 int
	nacts1, ncreds1 = listActs(db, ractress, atoms, actIns, credIns, nameIns)
	nacts2, ncreds2 = listActs(db, ractor, atoms, actIns, credIns, nameIns)

	logf("Done. Added %d actors/actresses and %d credits.",
		nacts1+nacts2, ncreds1+ncreds2)
}

func listActs(
	db *imdb.DB,
	r io.ReadCloser,
	atoms *imdb.Atomizer,
	actIns, credIns, nameIns *imdb.Inserter,
) (addedActors, addedCredits int) {
	bunkName, bunkTitles := []byte("Name"), []byte("Titles")
	bunkLines1, bunkLines2 := []byte("----"), []byte("------")

	listAttrRows(r, atoms, func(line, idstr, row []byte) {
		if bytes.Equal(idstr, bunkName) && bytes.Equal(row, bunkTitles) {
			return
		}
		if bytes.Equal(idstr, bunkLines1) && bytes.Equal(row, bunkLines2) {
			return
		}

		var a imdb.Actor
		existed, err := parseId(atoms, idstr, &a.Id)
		if err != nil {
			csql.Panic(err)
		}
		if !existed {
			if !parseActorName(idstr, &a) {
				logf("Could not parse actor name '%s' in '%s'.", idstr, line)
				return
			}
			if err := actIns.Exec(a.Id, a.Sequence); err != nil {
				csql.Panic(ef("Could not add actor info '%#v' from '%s': %s",
					a, line, err))
			}
			if err := nameIns.Exec(a.Id, a.FullName); err != nil {
				csql.Panic(ef("Could not add actor name '%s' from '%s': %s",
					idstr, line, err))
			}
			addedActors++
		}

		// Reading this list always refreshes the credits.
		var c imdb.Credit
		c.ActorId = a.Id
		if !parseCredit(atoms, row, &c) {
			// messages are emitted in parseCredit if something is worth
			// reporting
			return
		}
		err = credIns.Exec(c.ActorId, c.MediaId,
			c.Character, c.Position, c.Attrs)
		if err != nil {
			csql.Panic(ef("Could not add credit '%s' for '%s': %s",
				row, idstr, err))
		}
		addedCredits++
	})
	return
}

func parseActorName(idstr []byte, a *imdb.Actor) bool {
	var name, sequence []byte
	if idstr[len(idstr)-1] == ')' {
		fields := bytes.Fields(idstr)
		last := fields[len(fields)-1]
		if last[0] == '(' && last[len(last)-1] == ')' {
			name = bytes.Join(fields[0:len(fields)-1], []byte{' '})
			sequence = last[1 : len(last)-1]
		} else {
			name = idstr
		}
	} else {
		name = idstr
	}
	sep := bytes.IndexByte(name, ',')
	if sep > -1 {
		var flipped []byte
		l, f := bytes.TrimSpace(name[0:sep]), bytes.TrimSpace(name[sep+1:])
		flipped = append(flipped, f...)
		flipped = append(flipped, ' ')
		flipped = append(flipped, l...)
		name = flipped
	}

	a.FullName = unicode(name)
	a.Sequence = unicode(sequence)
	return true
}

func parseCredit(atoms *imdb.Atomizer, row []byte, c *imdb.Credit) bool {
	pieces := bytes.Split(row, []byte{' ', ' '})
	ent := bytes.TrimSpace(pieces[0])
	if id, ok := atoms.AtomOnlyIfExist(ent); !ok {
		warnf("Could not find media id for '%s'. Skipping.", ent)
		return false
	} else {
		c.MediaId = id
	}
	for _, f := range pieces[1:] {
		f = bytes.TrimSpace(f)
		if len(f) < 3 {
			continue
		}
		switch {
		case f[0] == '<' && f[len(f)-1] == '>':
			if err := parseInt(f[1:len(f)-1], &c.Position); err != nil {
				pef("Could not parse '%s' as integer in '%s': %s", f, row, err)
				return false
			}
		case f[0] == '[' && f[len(f)-1] == ']':
			c.Character = unicode(bytes.TrimSpace(f[1 : len(f)-1]))
		case f[0] == '(' && f[len(f)-1] == ')':
			c.Attrs = unicode(f)
		}
	}
	return true
}
