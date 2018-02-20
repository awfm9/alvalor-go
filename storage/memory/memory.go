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

package memory

import (
	"errors"

	cache "github.com/patrickmn/go-cache"
)

// Memory represents a wrapper around an in-memory storage.
type Memory struct {
	cache *cache.Cache
}

// New creates a new cache wrapper.
func New() *Memory {
	m := &Memory{
		cache: cache.New(cache.NoExpiration, cache.NoExpiration),
	}
	return m
}

// Store will store the entity data under its hash.
func (m *Memory) Store(hash []byte, data []byte, force bool) error {
	_, ok := m.cache.Get(string(hash))
	if ok && !force {
		return errors.New("hash already exists")
	}
	m.cache.Set(string(hash), data, cache.NoExpiration)
	return nil
}

// Retrieve will retrieve the entity data by its hash.
func (m *Memory) Retrieve(hash []byte) ([]byte, error) {
	data, ok := m.cache.Get(string(hash))
	if !ok {
		return nil, errors.New("hash doesn't exist")
	}
	return data.([]byte), nil
}

// Delete will remove the entity data by its hash.
func (m *Memory) Delete(hash []byte) error {
	_, ok := m.cache.Get(string(hash))
	if !ok {
		return errors.New("hash doesn't exist")
	}
	m.cache.Delete(string(hash))
	return nil
}
