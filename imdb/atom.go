package imdb

import (
	"bytes"
	"crypto/md5"

	"github.com/BurntSushi/csql"
)

type Atomer interface {
	Atom(key []byte) (Atom, bool, error)
}

type Atom int32

type atomMap map[[md5.Size]byte]Atom

type Atomizer struct {
	db     *DB
	atoms  atomMap
	nextId Atom
	ins    *Inserter
}

func (db *DB) NewAtomizer(tx *Tx) (*Atomizer, error) {
	az := &Atomizer{db, make(atomMap, 1000000), 0, nil}
	err := csql.Safe(func() {
		if tx != nil {
			var err error
			az.ins, err = db.NewInserter(tx, 50, "atom", "id", "hash")
			csql.SQLPanic(err)
		}

		rs := csql.Query(db, "SELECT id, hash FROM atom ORDER BY id ASC")
		csql.SQLPanic(csql.ForRow(rs, func(scanner csql.RowScanner) {
			var id Atom
			var hashBytes, hashBytesScan []byte
			csql.Scan(scanner, &id, &hashBytesScan)

			var hash [md5.Size]byte
			hashBytes = hash[:]
			copy(hashBytes, hashBytesScan)
			az.atoms[hash] = id
			az.nextId = id
		}))
	})
	az.nextId++
	return az, err
}

func (az *Atomizer) Atom(key []byte) (Atom, bool, error) {
	hash := hashKey(key)
	if a, ok := az.atoms[hash]; ok {
		return a, true, nil
	}
	a, err := az.add(hash)
	return a, false, err
}

func (az *Atomizer) AtomOnlyIfExist(key []byte) (Atom, bool) {
	hash := hashKey(key)
	a, ok := az.atoms[hash]
	return a, ok
}

func (az *Atomizer) add(hash [md5.Size]byte) (Atom, error) {
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

func (az *Atomizer) Close() error {
	if az.ins != nil {
		return az.ins.Close()
	}
	return nil
}

func (a Atom) String() string {
	return sf("%d", a)
}

func hashKey(s []byte) [md5.Size]byte {
	sum := md5.Sum(bytes.TrimSpace(s))
	return sum
}
