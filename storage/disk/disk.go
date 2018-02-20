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

package disk

import (
	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
)

// Disk represents a wrapper around an in-memory storage.
type Disk struct {
	db *badger.DB
}

// New creates a new cache wrapper.
func New(path string) (*Disk, error) {
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path
	db, err := badger.Open(opts)
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize badger database")
	}
	d := &Disk{
		db: db,
	}
	return d, nil
}

// Store will store the entity data under its hash.
func (m *Disk) Store(hash []byte, data []byte) error {
	err := m.db.Update(func(tx *badger.Txn) error {
		return tx.Set(hash, data)
	})
	if err != nil {
		return errors.Wrap(err, "could not store entity in DB")
	}
	return nil
}

// Retrieve will retrieve the entity data by its hash.
func (m *Disk) Retrieve(hash []byte) ([]byte, error) {
	var data []byte
	err := m.db.View(func(tx *badger.Txn) error {
		var inErr error
		item, inErr := tx.Get(hash)
		if inErr != nil {
			return errors.Wrapf(inErr, "could not get item by hash")
		}
		data, inErr = item.Value()
		if inErr != nil {
			return errors.Wrap(inErr, "could not get value from item")
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not get hash data")
	}
	return data, nil
}

// Delete will remove the entity data by its hash.
func (m *Disk) Delete(hash []byte) error {
	err := m.db.Update(func(tx *badger.Txn) error {
		return tx.Delete(hash)
	})
	if err != nil {
		return errors.Wrap(err, "could not delete entry")
	}
	return nil
}

// Close will shut down the database.
func (m *Disk) Close() error {
	return m.db.Close()
}
