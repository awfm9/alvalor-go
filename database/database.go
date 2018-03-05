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

package database

import (
	"bytes"
	"hash"

	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"

	"github.com/alvalor/alvalor-go/trie"
)

// DB is a blockchain database that syncs the trie with the persistent key-value store on disk.
type DB struct {
	h  hash.Hash
	cd Codec
	tr *trie.Trie
	kv *badger.DB
}

// Load loads a blockchain database from disk.
func Load(h hash.Hash, cd Codec, kv *badger.DB) (*DB, error) {
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
			err = tr.Put(key, value)
			if err != nil {
				return errors.Wrapf(err, "could not put value (%x) with key (%x)", value, key)
			}
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not execute iteration transaction")
	}
	db := &DB{h: h, cd: cd, tr: tr, kv: kv}
	return db, nil
}

// Insert will insert an entity that has a unique ID into the database.
func (db *DB) Insert(entity Entity) error {
	buf := &bytes.Buffer{}
	err := db.cd.Encode(buf, entity)
	if err != nil {
		return errors.Wrap(err, "could not encode entity")
	}
	data := buf.Bytes()
	db.h.Reset()
	db.h.Write(data)
	hash := db.h.Sum(nil)
	err = db.kv.Update(func(tx *badger.Txn) error {
		return tx.Set(hash, data)
	})
	if err != nil {
		return errors.Wrap(err, "could not store data for hash")
	}
	key := entity.ID()
	err = db.tr.Put(key, hash)
	if err != nil {
		return errors.Wrapf(err, "could not store hash for key (%x)", hash)
	}
	return nil
}

// Retrieve will retrieve an entity from the database by its unique ID.
func (db *DB) Retrieve(key []byte) (Entity, error) {
	hash, err := db.tr.Get(key)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve hash for key")
	}
	var data []byte
	err = db.kv.View(func(tx *badger.Txn) error {
		var inErr error
		item, inErr := tx.Get(hash)
		if inErr != nil {
			return errors.Wrapf(inErr, "could not retrieve item for hash (%x)", hash)
		}
		data, inErr = item.Value()
		if inErr != nil {
			return errors.Wrap(inErr, "could not retrieve value for item")
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve data for hash")
	}
	buf := bytes.NewBuffer(data)
	entity, err := db.cd.Decode(buf)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode entity")
	}
	return entity.(Entity), nil
}
