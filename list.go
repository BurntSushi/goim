package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"strconv"

	"github.com/BurntSushi/ty/fun"

	"github.com/BurntSushi/csql"

	"github.com/BurntSushi/goim/imdb"
)

var (
	tab      = []byte{'\t'}
	space    = []byte{' '}
	hypen    = []byte{'-'}
	openHash = []byte{'(', '#'}
)

type listHandler func(db *imdb.DB, list io.ReadCloser)

func listLoad(db *imdb.DB, list io.ReadCloser, handler listHandler) error {
	gzlist, err := gzip.NewReader(list)
	if err != nil {
		return err
	}
	defer list.Close()
	defer gzlist.Close()
	return csql.Safe(func() { handler(db, gzlist) })
}

// listAttrRows is a convenience function for traversing lines in IMDb
// lists that provide multiple instances of attributes for any particular
// entity. For example, the 'aka-titles' list has this format:
//
//	Mysteries of Egypt (1998)
//		(aka Egypt (1998))	(USA) (short title)
//		(aka Ägypten - Geheimnisse der Pharaonen (1998))	(Germany)
//
// The function given will called twice---for each attribute row---and will be
// supplied with the atom ID for "Mysteries of Egypt" along with the bytes for
// each attribute row. The entity name is also included for debugging purposes
// or in case the caller needs to look for extra information.
//
// Note that this formatting would produce an equivalent result:
//
//	Mysteries of Egypt (1998)	(aka Egypt (1998))	(USA) (short title)
//		(aka Ägypten - Geheimnisse der Pharaonen (1998))	(Germany)
//
// (Note the tab character following "Mysteries of Egypt (1998)".)
//
// Finally, if an atom ID cannot be found, the entry is skipped.
//
// (For the particular format described above, you'll likely find
// 'parseNamedAttr' useful.)
func listAttrRows(
	list io.ReadCloser,
	atoms *imdb.Atomizer,
	do func(id imdb.Atom, line, entity, row []byte),
) {
	var curEntity []byte
	var curAtom imdb.Atom
	var ok bool
	listLines(list, func(line []byte) {
		// Safe to ignore new lines here, since we can tell where we are by
		// the character in the first column.
		if len(line) == 0 {
			return
		}

		var row []byte
		if line[0] == ' ' || line[0] == '\t' { // just an attr row
			row = bytes.TrimSpace(line)
		} else { // specifying a new entity
			// If there's an attr row with the entity, separate it.
			entity := bytes.TrimSpace(line)
			sep := bytes.IndexByte(line, '\t')
			if sep > -1 {
				if sep+1 < len(entity) {
					row = bytes.TrimSpace(entity[sep+1:])
				}
				entity = bytes.TrimSpace(entity[0:sep])
			}
			if curAtom, ok = atoms.AtomOnlyIfExist(entity); !ok {
				warnf("Could not find id for '%s'. Skipping.", entity)
				curAtom, curEntity = 0, nil // indicates skipping
			} else {
				curEntity = entity
			}
		}
		// If no atom could be found, then we're skipping.
		if curAtom == 0 {
			warnf("No atom id found, so skipping: '%s'", line)
			return
		}
		// An attr row can be on a line by itself, or it can be on the same
		// line as the entity (delimited by a tab).
		if len(row) > 0 {
			// line != row when row is on same line as entity.
			do(curAtom, line, curEntity, row)
		}
	})
}

// listLines is a convenience function for traversing lines in most IMDb
// plain text list files. In particular, it ignores lines in
// the header/footer and lines containing the text '{{SUSPENDED}}'.
//
// Lines are not trimmed. Empty lines are NOT ignored.
func listLines(list io.ReadCloser, do func([]byte)) {
	dataStart, dataEnd := []byte("====="), []byte("----------")
	dataSection := false
	scanner := bufio.NewScanner(list)
	for scanner.Scan() {
		line := scanner.Bytes()
		if !dataSection {
			if bytes.HasPrefix(line, dataStart) {
				dataSection = true
			}
			continue
		}
		if dataSection && bytes.HasPrefix(line, dataEnd) {
			break
		}
		if bytes.Contains(line, attrSuspended) {
			continue
		}
		do(line)
	}
	csql.Panic(scanner.Err())
}

// splitListLine returns fields of the given line determined by tab characters.
// Note that this removes empty field, since an unpredictable number of tab
// characters often separates fields in list files.
func splitListLine(line []byte) [][]byte {
	fields := bytes.Split(line, tab)
	for i := len(fields) - 1; i >= 0; i-- { // go backwards to delete in place
		if len(fields[i]) == 0 {
			fields = append(fields[:i], fields[i+1:]...)
		}
	}
	return fields
}

// parseMediaEntity returns either a imdb.Movie, imdb.Tvshow or imdb.Episode
// based on the data in the text provided. Note that the text should correspond
// to the contents of the entire entity. For example, for the Simpsons episode
// "Lisa the Iconoclast", the entity string is:
//
//	"The Simpsons" (1989) {Lisa the Iconoclast (#7.16)}
//
// And this function will return it as a valid imdb.Episode.
//
// If the entity isn't a valid movie/tvshow/episode, then the boolean returned
// will be false.
//
// The 'Id' field of the returned entity is always zero. Also, if the entity
// is an episode, the TV show ID will be zero too.
func parseMediaEntity(entity []byte) (imdb.Entity, bool) {
	switch ent := entityType("media", entity); ent {
	case imdb.EntityMovie:
		var e imdb.Movie
		if !parseMovie(entity, &e) {
			return nil, false
		}
		return &e, true
	case imdb.EntityTvshow:
		var e imdb.Tvshow
		if !parseTvshow(entity, &e) {
			return nil, false
		}
		return &e, true
	case imdb.EntityEpisode:
		var e imdb.Episode
		if !parseEpisode(nil, entity, &e) {
			return nil, false
		}
		return &e, true
	default:
		return nil, false
	}
}

// parseNamedAttr returns the contents of text in the form
// '(attr-name {DATA})'. The 'attr-name' is returned first and the '{DATA}'
// is returned second.
// If there was a problem parsing this as a named attr, then the boolean is
// returned as false.
func parseNamedAttr(namedAttr []byte) ([]byte, []byte, bool) {
	if len(namedAttr) < 5 {
		return nil, nil, false
	}
	if namedAttr[0] != '(' && namedAttr[len(namedAttr)-1] != ')' {
		return nil, nil, false
	}
	namedAttr = namedAttr[1 : len(namedAttr)-1]
	pieces := bytes.SplitN(namedAttr, []byte{' '}, 2)
	if len(pieces) <= 1 {
		return nil, nil, false
	}
	return bytes.TrimSpace(pieces[0]), bytes.TrimSpace(pieces[1]), true
}

// parseId attempts to retrieve a uniquely identifying integer for this
// record. If one doesn't exist, it is created and returned. Otherwise, the
// existing one is returned.
//
// The boolean returned is true if and only if the atom previously existed.
// (e.g., This is useful information because it allows you to quit parsing some
// lines if you know their data has already been recorded.)
//
// If there was an error, it is returned and the atom is considered to not
// have existed.
func parseId(az imdb.Atomer, idStr []byte, id *imdb.Atom) (bool, error) {
	atom, existed, err := az.Atom(idStr)
	if err != nil {
		return false, ef("Could not atomize '%s': %s", idStr, err)
	}
	*id = atom
	return existed, nil
}

func parseEntryYear(inParens []byte, store *int, sequence *string) error {
	if inParens[0] == '(' && inParens[len(inParens)-1] == ')' {
		inParens = inParens[1 : len(inParens)-1]
	}
	if !bytes.Equal(inParens[0:4], attrUnknownYear) {
		n, err := strconv.Atoi(string(inParens[0:4]))
		if err != nil {
			return err
		}
		*store = int(n)
	}
	if sequence != nil && len(inParens) > 4 && inParens[4] == '/' {
		*sequence = unicode(inParens[5:])
	}
	return nil
}

func parseInt(bs []byte, store *int) error {
	n, err := strconv.Atoi(string(bs))
	if err != nil {
		return err
	}
	*store = int(n)
	return nil
}

func assertInsert(
	ins *imdb.Inserter,
	line []byte,
	what string,
	args ...interface{},
) {
	if err := ins.Exec(args...); err != nil {
		toStr := func(v interface{}) string { return sf("%#v", v) }
		logf("Full %s info (that failed to add): %s",
			what, fun.Map(toStr, args).([]string))
		csql.Panic(ef("Error adding to %s table '%s': %s", what, line, err))
	}
}

func unicode(latin1 []byte) string {
	runes := make([]rune, len(latin1))
	for i := range latin1 {
		runes[i] = rune(latin1[i])
	}
	return string(runes)
}

// hasEntryYear returns true if and only if
// 'f' is of the form '(YYYY[/RomanNumeral])'.
func hasEntryYear(f []byte) bool {
	return len(f) >= 6 && f[0] == '(' && f[len(f)-1] == ')'
}

func entityType(listName string, item []byte) imdb.EntityKind {
	switch listName {
	case "media":
		switch {
		case item[0] == '"':
			if item[len(item)-1] == '}' {
				return imdb.EntityEpisode
			} else {
				return imdb.EntityTvshow
			}
		default:
			return imdb.EntityMovie
		}
	}
	panic("BUG: unrecognized list name " + listName)
}

type simpleLoad struct {
	db    *imdb.DB
	table string
	count int
	ins   *imdb.Inserter
	atoms *imdb.Atomizer
}

func startSimpleLoad(db *imdb.DB, table string, columns ...string) *simpleLoad {
	logf("Reading list to populate table %s...", table)
	idxs(db, table).drop()
	csql.Panic(csql.Truncate(db, db.Driver, table))

	tx, err := db.Begin()
	csql.Panic(err)
	ins, err := db.NewInserter(tx, 50, table, columns...)
	csql.Panic(err)
	atoms, err := db.NewAtomizer(nil) // read only
	csql.Panic(err)
	return &simpleLoad{db, table, 0, ins, atoms}
}

func (sl *simpleLoad) add(line []byte, args ...interface{}) {
	assertInsert(sl.ins, line, sl.table, args...)
	sl.count++
}

func (sl *simpleLoad) done() {
	csql.Panic(sl.db.CloseInserters()) // must come first (to let idxs get lock)
	idxs(sl.db, sl.table).create()
	logf("Done with table %s. Inserted %d rows.", sl.table, sl.count)
}
