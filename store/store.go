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

package store

import (
	"bytes"

	"github.com/pkg/errors"

	"github.com/alvalor/alvalor-go/types"
)

// Store represents a store to store entities by unique ID.
type Store struct {
	kv     KV
	codec  Codec
	prefix []byte
}

// New creates a new store for entities, using the given codec to encode them, the key-value store to save the
// data and the prefix to add to the key for differentiation with other entities in the same store.
func New(kv KV, codec Codec, prefix string) *Store {
	return &Store{
		kv:     kv,
		codec:  codec,
		prefix: []byte(prefix),
	}
}

// Save will put a new entity into the store.
func (s *Store) Save(hash types.Hash, entity interface{}) error {
	buf := &bytes.Buffer{}
	err := s.codec.Encode(buf, entity)
	if err != nil {
		return errors.Wrap(err, "could not encode entity")
	}
	data := buf.Bytes()
	key := append(s.prefix, hash[:]...)
	err = s.kv.Put(key, data)
	if err != nil {
		return errors.Wrap(err, "could not put entity data")
	}
	return nil
}

// Retrieve will retrieve an entity from the store.
func (s *Store) Retrieve(hash types.Hash) (interface{}, error) {
	key := append(s.prefix, hash[:]...)
	data, err := s.kv.Get(key)
	if err != nil {
		return nil, errors.Wrap(err, "could not get entity data")
	}
	buf := bytes.NewBuffer(data)
	entity, err := s.codec.Decode(buf)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode entity")
	}
	return entity, nil
}
