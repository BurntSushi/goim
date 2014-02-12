package main

import (
	"bytes"
	"database/sql"
	"io"
	"strconv"
	"time"

	"github.com/BurntSushi/ty/fun"

	"github.com/BurntSushi/csql"

	"github.com/BurntSushi/goim/imdb"
)

type simpleLoad struct {
	db    *imdb.DB
	tx    *sql.Tx
	table string
	count int
	ins   *csql.Inserter
	atoms *atomizer
}

func startSimpleLoad(db *imdb.DB, table string, columns ...string) *simpleLoad {
	logf("Reading list to populate table %s...", table)
	idxs(db, table).drop()

	tx, err := db.Begin()
	csql.Panic(err)
	csql.Panic(csql.Truncate(tx, db.Driver, table))
	ins, err := csql.NewInserter(tx, db.Driver, 50, table, columns...)
	csql.Panic(err)
	atoms, err := newAtomizer(db, nil) // read only
	csql.Panic(err)
	return &simpleLoad{db, tx, table, 0, ins, atoms}
}

func (sl *simpleLoad) add(line []byte, args ...interface{}) {
	if err := sl.ins.Exec(args...); err != nil {
		toStr := func(v interface{}) string { return sf("%#v", v) }
		logf("Full %s info (that failed to add): %s",
			sl.table, fun.Map(toStr, args).([]string))
		logf("Context: %s", line)
		csql.Panic(ef("Error adding to %s table: %s", sl.table, err))
	}
	sl.count++
}

func (sl *simpleLoad) done() {
	csql.Panic(sl.ins.Exec()) // inserts anything left in the buffer
	csql.Panic(sl.tx.Commit())
	idxs(sl.db, sl.table).create()
	logf("Done with table %s. Inserted %d rows.", sl.table, sl.count)
}

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

func listRunningTimes(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "running_time",
		"atom_id", "country", "minutes", "attrs")
	defer table.done()

	parseRunningTime := func(text []byte, country *string, minutes *int) bool {
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

	listAttrRowIds(r, table.atoms, func(id imdb.Atom, line, ent, row []byte) {
		var (
			country string
			minutes int
			attrs   []byte
		)

		rowFields := splitListLine(row)
		if len(rowFields) == 0 {
			return // herp derp...
		}
		if !parseRunningTime(rowFields[0], &country, &minutes) {
			return
		}
		if len(rowFields) > 1 {
			attrs = rowFields[1]
		}
		table.add(line, id, country, minutes, unicode(attrs))
	})
}

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
		if id, ok = table.atoms.atomOnlyIfExist(entity); !ok {
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

func listAkaTitles(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "aka_title", "atom_id", "title", "attrs")
	defer table.done()

	parseAkaTitle := func(text []byte, title *string) bool {
		attrName, data, ok := parseNamedAttr(text)
		if !ok {
			return false
		}
		if !bytes.Equal(attrName, []byte("aka")) {
			return false
		}
		ent, ok := parseMediaEntity(data)
		if !ok {
			return false
		}
		*title = ent.Name()
		return true
	}

	listAttrRowIds(r, table.atoms, func(id imdb.Atom, line, ent, row []byte) {
		var (
			title string
			attrs []byte
		)

		fields := splitListLine(row)
		if len(fields) == 0 {
			return // herp derp...
		}
		if !parseAkaTitle(fields[0], &title) {
			if !bytes.Contains(fields[0], []byte("(VG)")) {
				logf("Could not parse aka title from '%s'", fields[0])
			}
			return
		}
		if len(fields) > 1 {
			attrs = fields[1]
		}
		table.add(line, id, title, unicode(attrs))
	})
}

func listMovieLinks(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "link", "atom_id",
		"link_type", "link_atom_id", "entity")
	defer table.done()

	parseMovieLink := func(
		atoms *atomizer,
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
		id, ok := atoms.atomOnlyIfExist(data)
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

	listAttrRowIds(r, table.atoms, func(id imdb.Atom, line, ent, row []byte) {
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

func listColorInfo(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "color_info",
		"atom_id", "color", "attrs")
	defer table.done()

	var (
		infoColor = []byte("Color")
		infoBandW = []byte("Black and White")
	)

	parseColorInfo := func(text []byte, color *bool) bool {
		switch {
		case bytes.Equal(text, infoColor):
			*color = true
			return true
		case bytes.Equal(text, infoBandW):
			*color = false
			return true
		}
		logf("Could not parse '%s' as color information.", text)
		return false
	}

	listAttrRowIds(r, table.atoms, func(id imdb.Atom, line, ent, row []byte) {
		var (
			color bool
			attrs []byte
		)

		rowFields := splitListLine(row)
		if len(rowFields) == 0 {
			return // herp derp...
		}
		if !parseColorInfo(rowFields[0], &color) {
			return
		}
		if len(rowFields) > 1 {
			attrs = rowFields[1]
		}
		table.add(line, id, color, unicode(attrs))
	})
}

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
			if curAtom, ok = table.atoms.atomOnlyIfExist(entity); !ok {
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

func listReleaseDates(db *imdb.DB, r io.ReadCloser) {
	table := startSimpleLoad(db, "release_date",
		"atom_id", "country", "released", "attrs")
	defer table.done()

	parseDate := func(text []byte, country *string, released *time.Time) bool {
		sep := bytes.IndexByte(text, ':')
		var date []byte
		if sep > -1 {
			*country = unicode(bytes.TrimSpace(text[:sep]))
			date = bytes.TrimSpace(text[sep+1:])
		} else {
			*country = ""
			date = bytes.TrimSpace(text)
		}

		var layout string
		switch spaces := len(bytes.Fields(date)); spaces {
		case 3:
			layout = "2 January 2006"
		case 2:
			layout = "January 2006"
		case 1:
			layout = "2006"
		default:
			pef("Too many fields in date '%s' (%d) in '%s'", date, spaces, text)
			return false
		}

		t, err := time.Parse(layout, unicode(date))
		if err != nil {
			pef("Could not parse date '%s': %s", date, err)
			return false
		}
		*released = t.UTC()
		return true
	}

	listAttrRowIds(r, table.atoms, func(id imdb.Atom, line, ent, row []byte) {
		var (
			country string
			date    time.Time
			attrs   string
		)

		rowFields := splitListLine(row)
		if !parseDate(rowFields[0], &country, &date) {
			pef("Could not extract date from '%s'. Skipping.", line)
			return
		}
		if len(rowFields) > 1 {
			attrs = unicode(rowFields[1])
		}
		table.add(line, id, country, date, attrs)
	})
}

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
			if curAtom, ok = table.atoms.atomOnlyIfExist(entity); !ok {
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
			if curAtom, ok = table.atoms.atomOnlyIfExist(entity); !ok {
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
