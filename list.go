package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"strconv"

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

func listLines(list io.ReadCloser, do func([]byte) bool) {
	dataStart, dataEnd := []byte("====="), []byte("----------")
	dataSection := false
	scanner := bufio.NewScanner(list)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
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
		if !do(line) {
			break
		}
	}
	csql.SQLPanic(scanner.Err())
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

type entity int

const (
	entityMovie = iota
	entityTvshow
	entityEpisode
)

func entityType(listName string, item []byte) entity {
	switch listName {
	case "movies", "release-dates":
		switch {
		case item[0] == '"':
			if item[len(item)-1] == '}' {
				return entityEpisode
			} else {
				return entityTvshow
			}
		default:
			return entityMovie
		}
	}
	panic("unrecognized list name " + listName)
}

func (e entity) String() string {
	switch e {
	case entityMovie:
		return "movie"
	case entityTvshow:
		return "tvshow"
	case entityEpisode:
		return "episode"
	}
	panic(sf("unrecognized entity %d", int(e)))
}
