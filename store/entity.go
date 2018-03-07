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

// Entity represents an entity that has a unique ID to use as key to store.
type Entity interface {
	Hash() []byte
}

// EntityStore represents a store to store entities by unique ID.
type EntityStore struct {
	kv     KV
	codec  Codec
	prefix []byte
}

// NewEntity creates a new store for entities, using the given codec to encode them, the key-value store to save the
// data and the prefix to add to the key for differentiation with other entities in the same store.
func NewEntity(kv KV, codec Codec, prefix string) *EntityStore {
	return &EntityStore{
		kv:     kv,
		codec:  codec,
		prefix: []byte(prefix),
	}
}

// Save will put a new entity into the store.
func (es *EntityStore) Save(entity Entity) error {
	buf := &bytes.Buffer{}
	err := es.codec.Encode(buf, entity)
	if err != nil {
		return errors.Wrap(err, "could not encode entity")
	}
	hash := entity.Hash()
	data := buf.Bytes()
	key := append(es.prefix, hash...)
	err = es.kv.Put(key, data)
	if err != nil {
		return errors.Wrap(err, "could not put entity data")
	}
	return nil
}

// Retrieve will retrieve an entity from the store.
func (es *EntityStore) Retrieve(id []byte) (interface{}, error) {
	key := append(es.prefix, id...)
	data, err := es.kv.Get(key)
	if err != nil {
		return nil, errors.Wrap(err, "could not get entity data")
	}
	buf := bytes.NewBuffer(data)
	entity, err := es.codec.Decode(buf)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode entity")
	}
	return entity, nil
}
