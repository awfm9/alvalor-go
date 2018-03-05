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
)

// Store represents a store for different entities.
type Store struct {
	codec Codec
	kv    KV
}

// New creates a new store to store entities.
func New(codec Codec, kv KV) *Store {
	return &Store{
		codec: codec,
		kv:    kv,
	}
}

// Save will put a new entity into the store.
func (s *Store) Save(entity Entity) error {
	buf := &bytes.Buffer{}
	err := s.codec.Encode(buf, entity)
	if err != nil {
		return errors.Wrap(err, "could not encode entity")
	}
	id := entity.ID()
	data := buf.Bytes()
	err = s.kv.Put(id, data)
	if err != nil {
		return errors.Wrap(err, "could not put entity data")
	}
	return nil
}

// Retrieve will retrieve an entity from the store.
func (s *Store) Retrieve(id []byte) (interface{}, error) {
	data, err := s.kv.Get(id)
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
