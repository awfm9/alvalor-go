// Copyright (c) 2017 The Alvalor Authors
//
// This file is part of Alvalor.
//
// Alvalor is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Alvalor is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Alvalor.  If not, see <http://www.gnu.org/licenses/>.

package blockchain

import (
	"github.com/alvalor/alvalor-go/trie"
	"github.com/dgraph-io/badger/badger"
	"github.com/pkg/errors"
)

// DB is a blockchain database that syncs the trie with the hard disk persistence store.
type DB struct {
	kv *badger.KV
	tr *trie.Trie
}

// NewDB creates a new blockchain DB on the disk.
func NewDB(kv *badger.KV) (*DB, error) {
	tr := trie.New()
	itr := kv.NewIterator(badger.DefaultIteratorOptions)
	for itr.Rewind(); itr.Valid(); itr.Next() {
		item := itr.Item()
		key := item.Key()
		val := item.Value()
		ok := tr.Put(key, val, false)
		if !ok {
			return nil, errors.Errorf("could not insert key %x", key)
		}
	}
	itr.Close()
	db := &DB{kv: kv, tr: tr}
	return db, nil
}

// Insert will insert a new entity into the trie.
func (db *DB) Insert(entity Entity) error {
	data := entity.Bytes()
	hash := hash(data)
	err := db.kv.Set(hash, data)
	if err != nil {
		return errors.Wrap(err, "could not save entity on disk")
	}
	id := entity.ID()
	ok := db.tr.Put(id, hash, false)
	if !ok {
		return errors.Errorf("could not insert entity %x into trie", entity.ID())
	}
	return nil
}

// Retrieve will retrieve an entity from the trie.
func (db *DB) Retrieve(id []byte, entity Entity) error {
	hash, ok := db.tr.Get(id)
	if !ok {
		return errors.Errorf("could not find entity %x in trie", id)
	}
	var kv badger.KVItem
	err := db.kv.Get(hash, &kv)
	if err != nil {
		return errors.Wrap(err, "could not retrieve entity from disk")
	}
	data := kv.Value()
	err = entity.FromBytes(data)
	if err != nil {
		return errors.Wrap(err, "could not decode entity")
	}
	return nil
}
