package main

import (
	"bytes"
	"io"

	"github.com/BurntSushi/csql"
	"github.com/BurntSushi/goim/imdb"
)

func listActors(db *imdb.DB, ractor, ractress io.ReadCloser) (err error) {
	defer csql.Safe(&err)

	logf("Reading actors list...")

	// PostgreSQL wants different transactions for each inserter.
	// SQLite can't handle them. The wrapper type here ensures that
	// PostgreSQL gets multiple transactions while SQLite only gets one.
	tx, err := db.Begin()
	csql.Panic(err)

	txactor := wrapTx(db, tx)
	txcredit := txactor.another()
	txname := txactor.another()
	txatom := txactor.another()

	// Drop data from the actor and credit tables. They will be rebuilt below.
	// The key here is to leave the atom and name tables alone. Invariably,
	// they will contain stale data. But the only side effect, I think, is
	// taking up space.
	// (Stale data can be removed with 'goim clean'.)
	csql.Panic(csql.Truncate(txactor, db.Driver, "actor"))
	csql.Panic(csql.Truncate(txcredit.Tx, db.Driver, "credit"))

	actIns, err := csql.NewInserter(txactor.Tx, db.Driver, "actor",
		"atom_id", "sequence")
	csql.Panic(err)
	credIns, err := csql.NewInserter(txcredit.Tx, db.Driver, "credit",
		"actor_atom_id", "media_atom_id", "character", "position", "attrs")
	csql.Panic(err)
	nameIns, err := csql.NewInserter(txname.Tx, db.Driver, "name",
		"atom_id", "name")
	csql.Panic(err)
	atoms, err := newAtomizer(db, txatom.Tx)
	csql.Panic(err)

	// Unfortunately, it looks like credits for an actor can appear in
	// multiple locations. (Or there are different actors that erroneously
	// have the same name.)
	added := make(map[imdb.Atom]struct{}, 3000000)
	n1, nc1 := listActs(db, ractress, atoms, added, actIns, credIns, nameIns)
	n2, nc2 := listActs(db, ractor, atoms, added, actIns, credIns, nameIns)

	csql.Panic(actIns.Exec())
	csql.Panic(credIns.Exec())
	csql.Panic(nameIns.Exec())
	csql.Panic(atoms.Close())

	csql.Panic(txactor.Commit())
	csql.Panic(txcredit.Commit())
	csql.Panic(txname.Commit())
	csql.Panic(txatom.Commit())

	logf("Done. Added %d actors/actresses and %d credits.", n1+n2, nc1+nc2)
	return
}

type credit struct {
	ActorId   imdb.Atom
	MediaId   imdb.Atom
	Character string
	Position  int
	Attrs     string
}

func listActs(
	db *imdb.DB,
	r io.ReadCloser,
	atoms *atomizer,
	added map[imdb.Atom]struct{},
	actIns, credIns, nameIns *csql.Inserter,
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

			// We only add a name when we've added an atom.
			if err := nameIns.Exec(a.Id, a.FullName); err != nil {
				csql.Panic(ef("Could not add actor name '%s' from '%s': %s",
					idstr, line, err))
			}
		}

		// If we haven't seen this actor before, then insert into actor table.
		if _, ok := added[a.Id]; !ok {
			if len(a.FullName) == 0 {
				if !parseActorName(idstr, &a) {
					logf("Could not get actor name '%s' in '%s'.", idstr, line)
					return
				}
			}
			if err := actIns.Exec(a.Id, a.Sequence); err != nil {
				csql.Panic(ef("Could not add actor info '%#v' from '%s': %s",
					a, line, err))
			}
			added[a.Id] = struct{}{}
			addedActors++
		}

		// Reading this list always refreshes the credits.
		var c credit
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

func parseCredit(atoms *atomizer, row []byte, c *credit) bool {
	pieces := bytes.Split(row, []byte{' ', ' '})
	ent := bytes.TrimSpace(pieces[0])
	if id, ok := atoms.atomOnlyIfExist(ent); !ok {
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
