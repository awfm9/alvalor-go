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

import "errors"

// Memory is a wrapper around an in-memory key-value store.
type Memory struct {
	kv map[string][]byte
}

// NewMemory creates a new in-memory KV store.
func NewMemory() *Memory {
	return &Memory{kv: make(map[string][]byte)}
}

// Put will store the given value under the given key.
func (m *Memory) Put(key []byte, val []byte) error {
	m.kv[string(key)] = val
	return nil
}

// Has will check whether a value exists under the given key.
func (m *Memory) Has(key []byte) (bool, error) {
	_, ok := m.kv[string(key)]
	return ok, nil
}

// Get will return the value stored under the given key.
func (m *Memory) Get(key []byte) ([]byte, error) {
	val, ok := m.kv[string(key)]
	if !ok {
		return nil, errors.New("key not found")
	}
	return val, nil
}

// Del will delete the key-value pair with the given key.
func (m *Memory) Del(key []byte) error {
	delete(m.kv, string(key))
	return nil
}
