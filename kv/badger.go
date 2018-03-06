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

package kv

import (
	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
)

// Badger is a wrapper around the badger key-value stores that implements our database interface.
type Badger struct {
	kv *badger.DB
}

// NewBadger creates a new KV store wraper around a badger database.
func NewBadger(kv *badger.DB) *Badger {
	return &Badger{kv: kv}
}

// Put will insert a key-value pair into the database.
func (b *Badger) Put(key []byte, val []byte) error {
	err := b.kv.Update(func(tx *badger.Txn) error {
		err := tx.Set(key, val)
		if err != nil {
			return errors.Wrap(err, "could not set value for key")
		}
		return nil
	})
	return err
}

// Has will check if there is a value for the given key in the database.
func (b *Badger) Has(key []byte) (bool, error) {
	var ok bool
	err := b.kv.View(func(tx *badger.Txn) error {
		_, err := tx.Get(key)
		if err != nil {
			return errors.Wrap(err, "could not get item for key")
		}
		ok = true
		return nil
	})
	return ok, err
}

// Get will return the value for a given key, if it exists.
func (b *Badger) Get(key []byte) ([]byte, error) {
	var val []byte
	err := b.kv.View(func(tx *badger.Txn) error {
		var err error
		item, err := tx.Get(key)
		if err != nil {
			return errors.Wrap(err, "could not get item for key")
		}
		val, err = item.Value()
		if err != nil {
			return errors.Wrap(err, "could not get value from item")
		}
		return nil
	})
	return val, err
}

// Del will delete the key-value pair for the given key from the database, if it exists.
func (b *Badger) Del(key []byte) error {
	err := b.kv.Update(func(tx *badger.Txn) error {
		err := tx.Delete(key)
		if err != nil {
			return errors.Wrap(err, "could not delete value for key")
		}
		return nil
	})
	return err
}
