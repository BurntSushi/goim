package main

import (
	"bytes"
	"crypto/md5"
	"database/sql"

	"github.com/BurntSushi/csql"

	"github.com/BurntSushi/goim/imdb"
)

type tx struct {
	db     *imdb.DB
	closed bool
	*sql.Tx
}

func wrapTx(db *imdb.DB, trans *sql.Tx) *tx {
	return &tx{db, false, trans}
}

func (tx *tx) another() *tx {
	if tx.db.Driver == "sqlite3" {
		return tx
	}
	txx, err := tx.db.Begin()
	csql.Panic(err)
	return wrapTx(tx.db, txx)
}

func (tx *tx) Commit() error {
	if !tx.closed {
		tx.closed = true
		return tx.Tx.Commit()
	}
	return nil
}

type atomMap map[[md5.Size]byte]imdb.Atom

type atomizer struct {
	db     *imdb.DB
	atoms  atomMap
	nextId imdb.Atom
	ins    *csql.Inserter
}

func newAtomizer(db *imdb.DB, tx *sql.Tx) (*atomizer, error) {
	az := &atomizer{db, make(atomMap, 1000000), 0, nil}
	err := csql.Safe(func() {
		if tx != nil {
			var err error
			az.ins, err = csql.NewInserter(
				tx, db.Driver, 50, "atom", "id", "hash")
			csql.Panic(err)
		}

		rs := csql.Query(db, "SELECT id, hash FROM atom")
		csql.Panic(csql.ForRow(rs, az.readRow))
	})
	az.nextId++
	return az, err
}

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

func (az *atomizer) atom(key []byte) (imdb.Atom, bool, error) {
	hash := hashKey(key)
	if a, ok := az.atoms[hash]; ok {
		return a, true, nil
	}
	a, err := az.add(hash)
	return a, false, err
}

func (az *atomizer) atomOnlyIfExist(key []byte) (imdb.Atom, bool) {
	hash := hashKey(key)
	a, ok := az.atoms[hash]
	return a, ok
}

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

func (az *atomizer) Close() error {
	if az.ins != nil {
		ins := az.ins
		az.ins = nil
		return ins.Exec()
	}
	return nil
}

func hashKey(s []byte) [md5.Size]byte {
	sum := md5.Sum(bytes.TrimSpace(s))
	return sum
}
