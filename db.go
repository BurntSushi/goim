package main

import (
	"bytes"
	"crypto/md5"
	"database/sql"

	"github.com/BurntSushi/csql"

	"github.com/BurntSushi/goim/imdb"
)

// tx wraps a regular SQL transaction and provides two abstractions:
// 1) Commit may be called on tx multiple times. Each time after the first is
// a no-op (instead of an error).
// 2) A tx can "create" provide another transaction for simultaneous use.
// With the SQLite driver, this actually returns the same tx. But with
// PostgreSQL, a new tx is made.
type tx struct {
	db     *imdb.DB
	closed bool
	*sql.Tx
}

// wrapTx makes a new tx from a database connection and an existing SQL
// transaction.
func wrapTx(db *imdb.DB, trans *sql.Tx) *tx {
	return &tx{db, false, trans}
}

// another produces a transaction from tx. It may or may not return the same
// transaction depending on the driver being used.
func (tx *tx) another() *tx {
	if tx.db.Driver == "sqlite3" {
		return tx
	}
	txx, err := tx.db.Begin()
	csql.Panic(err)
	return wrapTx(tx.db, txx)
}

// Commit commits the transaction to the database. It is safe to call it more
// than once.
func (tx *tx) Commit() error {
	if !tx.closed {
		tx.closed = true
		return tx.Tx.Commit()
	}
	return nil
}

// atomMap stores a mapping from md5 hashes (in binary) to atom integer ids.
type atomMap map[[md5.Size]byte]imdb.Atom

// atomizer provides a readable/writable abstraction for accessing and creating
// new atom identifiers.
type atomizer struct {
	db     *imdb.DB
	atoms  atomMap
	nextId imdb.Atom
	ins    *csql.Inserter
}

// newAtomizer returns an atomizer that can be used to access or create new
// atom identifiers. Note that if tx is nil, then the atomizer returned is
// read-only (attempting to write will cause a panic).
//
// A read-only atomizer may be accessed from multiple goroutines
// simultaneously, but a read/write atomizer may NOT.
//
// If a read/write atomizer is created, then the caller is responsible for
// closing the transaction (which should be done immediately after a call to
// atomizer.Close).
//
// Note that this function loads the entire set of atoms from the database
// into memory, so it is costly.
func newAtomizer(db *imdb.DB, tx *sql.Tx) (az *atomizer, err error) {
	defer csql.Safe(&err)

	az = &atomizer{db, make(atomMap, 1000000), 0, nil}
	if tx != nil {
		var err error
		az.ins, err = csql.NewInserter(
			tx, db.Driver, "atom", "id", "hash")
		csql.Panic(err)
	}

	rs := csql.Query(db, "SELECT id, hash FROM atom ORDER BY id ASC")
	csql.Panic(csql.ForRow(rs, az.readRow))
	az.nextId++
	return
}

// readRow scans a row from the atom table into an atomMap.
func (az *atomizer) readRow(scanner csql.RowScanner) {
	var id imdb.Atom
	var rawBytes sql.RawBytes
	csql.Scan(scanner, &id, &rawBytes)

	var hash [md5.Size]byte
	hashBytes := hash[:]
	copy(hashBytes, rawBytes)
	az.atoms[hash] = id
	az.nextId = id
}

// atom returns the atom associated with the key string given, along with
// whether it already existed or not. If it didn't exist, then a new atom is
// created and returned (along with an error if there was a problem creating
// the atom).
func (az *atomizer) atom(key []byte) (imdb.Atom, bool, error) {
	hash := hashKey(key)
	if a, ok := az.atoms[hash]; ok {
		return a, true, nil
	}
	a, err := az.add(hash)
	return a, false, err
}

// atomOnlyIfExist returns an atom id for the key string given only if that
// key string has already been atomized. If it doesn't exist, then the zero
// atom is returned along with false. Otherwise, the atom id is returned along
// with true.
func (az *atomizer) atomOnlyIfExist(key []byte) (imdb.Atom, bool) {
	hash := hashKey(key)
	a, ok := az.atoms[hash]
	return a, ok
}

// add always adds the given hash to the database with a fresh and unique
// atom identifier.
func (az *atomizer) add(hash [md5.Size]byte) (imdb.Atom, error) {
	if az.ins == nil {
		panic("cannot add atoms when opened read-only")
	}
	a := az.nextId
	if err := az.ins.Exec(a, hash[:]); err != nil {
		return 0, err
	}
	az.atoms[hash] = a
	az.nextId++
	return a, nil
}

// Close inserts any new atoms lingering in the buffer into the database.
// This does NOT commit the transaction.
// If the atomizer is read-only, this is a no-op.
func (az *atomizer) Close() error {
	if az.ins != nil {
		ins := az.ins
		az.ins = nil
		return ins.Exec()
	}
	return nil
}

// hashKey returns a byte array corresponding to the md5 hash of the key
// string given.
func hashKey(s []byte) [md5.Size]byte {
	h := md5.New()
	h.Write(bytes.TrimSpace(s))
	slice := h.Sum(nil)

	var sum [md5.Size]byte
	for i := 0; i < md5.Size; i++ {
		sum[i] = slice[i]
	}
	return sum
}

// listTables itemizes the tables that are updated for each list name.
var listTables = map[string][]string{
	"movies": []string{
		"atom", "name", "movie", "tvshow", "episode",
	},
	"actors":               []string{"atom", "name", "actor", "credit"},
	"sound-mix":            []string{"sound_mix"},
	"genres":               []string{"genre"},
	"language":             []string{"language"},
	"locations":            []string{"location"},
	"trivia":               []string{"trivia"},
	"alternate-versions":   []string{"alternate_version"},
	"taglines":             []string{"tagline"},
	"goofs":                []string{"goof"},
	"literature":           []string{"literature"},
	"running-times":        []string{"running_time"},
	"ratings":              []string{"rating"},
	"aka-titles":           []string{"aka_title"},
	"movie-links":          []string{"link"},
	"color-info":           []string{"color_info"},
	"mpaa-ratings-reasons": []string{"mpaa_rating"},
	"release-dates":        []string{"release_date"},
	"quotes":               []string{"quote"},
	"plot":                 []string{"plot"},
}
