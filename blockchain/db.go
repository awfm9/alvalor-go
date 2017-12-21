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
	"bytes"

	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"

	"github.com/alvalor/alvalor-go/hasher"
	"github.com/alvalor/alvalor-go/trie"
)

// DB is a blockchain database that syncs the trie with the persistent key-value store on disk.
type DB struct {
	kv *badger.DB
	tr *trie.Trie
	cd Codec
}

// NewDB creates a new blockchain DB on the disk.
func NewDB(kv *badger.DB) (*DB, error) {
	tr := trie.New()
	err := kv.View(func(tx *badger.Txn) error {
		it := tx.NewIterator(badger.DefaultIteratorOptions)
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			value, err := item.Value()
			if err != nil {
				return errors.Wrapf(err, "could not retrieve item value (%x)", key)
			}
			ok := tr.Put(key, value, false)
			if !ok {
				return errors.Wrapf(err, "could not put value (%x) with key (%x)", value, key)
			}
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not execute iteration transaction")
	}
	db := &DB{kv: kv, tr: tr}
	return db, nil
}

// Insert will insert a new key and hash into the trie after storing the related hash and data on
// disk.
func (db *DB) Insert(id []byte, entity interface{}) error {
	buf := &bytes.Buffer{}
	err := db.cd.Encode(buf, entity)
	if err != nil {
		return errors.Wrap(err, "could not serialize entity")
	}
	data := buf.Bytes()
	hash := hasher.Sum256(data)
	err = db.kv.Update(func(tx *badger.Txn) error {
		return tx.Set(hash, data)
	})
	if err != nil {
		return errors.Wrap(err, "could not save entity on disk")
	}
	ok := db.tr.Put(id, hash, false)
	if !ok {
		return errors.Errorf("could not insert entity %x into trie", id)
	}
	return nil
}

// Retrieve will retrieve an entity from the key-value store by looking up the associated hash in
// the trie.
func (db *DB) Retrieve(id []byte) (interface{}, error) {
	hash, ok := db.tr.Get(id)
	if !ok {
		return nil, errors.New("could not get hash for id")
	}
	var value []byte
	err := db.kv.View(func(tx *badger.Txn) error {
		item, err := tx.Get(hash)
		if err != nil {
			return errors.Wrapf(err, "could not get item for hash (%x)", hash)
		}
		val, err := item.Value()
		if err != nil {
			return errors.Wrap(err, "could not get value for item")
		}
		value = val
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not lookup hash on key-value store")
	}
	buf := bytes.NewBuffer(value)
	entity, err := db.cd.Decode(buf)
	if err != nil {
		return nil, errors.Wrap(err, "could not deserialize entity")
	}
	return entity, nil
}
