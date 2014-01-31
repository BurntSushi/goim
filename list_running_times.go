package main

import (
	"bytes"
	"io"
	"strconv"

	"github.com/BurntSushi/csql"
	"github.com/BurntSushi/goim/imdb"
)

func listRunningTimes(db *imdb.DB, times io.ReadCloser) {
	defer idxs(db, "running_time").drop().create()
	defer func() { csql.Panic(db.CloseInserters()) }()

	logf("Reading running times list...")
	addedTimes := 0

	// It's easier to just blow away the times table and reconstruct it.
	csql.Panic(csql.Truncate(db, db.Driver, "running_time"))

	txDates, err := db.Begin()
	csql.Panic(err)

	timeIns, err := db.NewInserter(txDates, 50, "running_time",
		"atom_id", "outlet", "country", "minutes", "attrs")

	atoms, err := db.NewAtomizer(nil)
	csql.Panic(err)

	insert := func(line []byte, id imdb.Atom, o, c, a string, min int) {
		if err := timeIns.Exec(id, o, c, min, a); err != nil {
			logf("Full running time info (that failed to add): "+
				"id:%d, outlet:%s, country:%s, minutes:%d, attrs:'%s'",
				id, o, c, min, a)
			csql.Panic(ef("Error adding time '%s': %s", line, err))
		}
	}
	listLines(times, func(line []byte) bool {
		var (
			id      imdb.Atom
			ok      bool
			country string
			minutes int
			attrs   []byte
		)

		fields := splitListLine(line)
		if len(fields) <= 1 {
			// herp derp...
			return true
		}
		item, value := fields[0], fields[1]
		if len(fields) == 3 {
			attrs = bytes.TrimSpace(fields[2])
		}
		if id, ok = atoms.AtomOnlyIfExist(item); !ok {
			warnf("Could not find id for '%s'. Skipping.", item)
			return true
		}
		if !parseRunningTime(value, &country, &minutes) {
			return true
		}

		ent := entityType("running-times", item)
		insert(line, id, ent.String(), country, unicode(attrs), minutes)
		addedTimes++
		return true
	})
	logf("Done. Added %d running times.", addedTimes)
}

func parseRunningTime(text []byte, country *string, minutes *int) bool {
	sep := bytes.IndexByte(text, ':')
	var runtime []byte
	if sep > -1 {
		*country = unicode(bytes.TrimSpace(text[:sep]))
		runtime = bytes.TrimSpace(text[sep+1:])
	} else {
		*country = ""
		runtime = bytes.TrimSpace(text)
	}

	var err error
	*minutes, err = strconv.Atoi(unicode(runtime))
	if err != nil {
		// There are a lot of these.
		// From the looks of it, IMDb's web site just ignores them.
		// It's almost like it's freeform text... Yikes.
		return false
	}
	return true
}
