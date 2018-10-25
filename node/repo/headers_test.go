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

package repo

import (
	"testing"

	"github.com/alvalor/alvalor-go/types"
	"github.com/stretchr/testify/assert"
)

func TestHeadersAddExisting(t *testing.T) {

	// initialize the repository with required maps
	hr := &Headers{
		headers: make(map[types.Hash]*types.Header),
	}

	// create entities and set up state
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}

	header := &types.Header{Hash: hash1, Parent: hash2}

	hr.headers[header.Hash] = header

	// try adding header already known and check outcome
	err := hr.Add(header)
	assert.NotNil(t, err, "could add existing header")
	assert.Len(t, hr.headers, 1, "header map not one element")
}

func TestHeadersAddPending(t *testing.T) {

	// initialize the repository with required maps
	hr := &Headers{
		headers: make(map[types.Hash]*types.Header),
		pending: make(map[types.Hash][]*types.Header),
	}

	// create entities and set up state
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}

	header := &types.Header{Hash: hash1, Parent: hash2}

	// try adding header with missing parent and check outcome
	err := hr.Add(header)
	assert.Nil(t, err, "could not add pending")
	assert.Empty(t, hr.headers, "headers map not empty")
	if assert.Len(t, hr.pending, 1, "pending map not one element") {
		assert.Equal(t, hr.pending[header.Parent], []*types.Header{header}, "header not with correct key in pending map")
	}
}

func TestHeadersAddValid(t *testing.T) {

	// initialize the repository with required maps
	hr := &Headers{
		headers:  make(map[types.Hash]*types.Header),
		pending:  make(map[types.Hash][]*types.Header),
		children: make(map[types.Hash][]types.Hash),
	}

	// create entities and set up state
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}

	header := &types.Header{Hash: hash1, Parent: hash2}
	parent := &types.Header{Hash: hash2}

	hr.headers[parent.Hash] = parent

	// try adding header with existing parent and check outcome
	err := hr.Add(header)
	assert.Nil(t, err, "could not add valid")
	if assert.Len(t, hr.headers, 2, "headers map not two elements") {
		assert.Equal(t, hr.headers[header.Hash], header, "header not with correct key in header map")
	}
	if assert.Len(t, hr.children, 1, "children map not one element") {
		assert.Equal(t, hr.children[parent.Hash], []types.Hash{header.Hash})
	}
	assert.Empty(t, hr.pending, "pending map not empty")
}

func TestHeadersAddValidWithPending(t *testing.T) {

	// initialize the repository with required maps
	hr := &Headers{
		headers:  make(map[types.Hash]*types.Header),
		pending:  make(map[types.Hash][]*types.Header),
		children: make(map[types.Hash][]types.Hash),
	}

	// create entities and set up state
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}
	hash3 := types.Hash{0x3}
	hash4 := types.Hash{0x4}

	header := &types.Header{Hash: hash1, Parent: hash2}
	parent := &types.Header{Hash: hash2}
	child1 := &types.Header{Hash: hash3, Parent: hash1}
	child2 := &types.Header{Hash: hash4, Parent: hash1}

	hr.headers[parent.Hash] = parent
	hr.pending[header.Hash] = []*types.Header{child1, child2}

	// try adding header with existing parent and pending children and check outcome
	err := hr.Add(header)
	assert.Nil(t, err, "could not add valid")
	if assert.Len(t, hr.headers, 4, "headers map not four elements") {
		assert.Equal(t, hr.headers[hash1], header, "header not with correct key in header map")
		assert.Equal(t, hr.headers[hash3], child1, "child1 not with correct key in header map")
		assert.Equal(t, hr.headers[hash4], child2, "child2 not with correct key in header map")
	}
	if assert.Len(t, hr.children, 2, "children map not two elements") {
		assert.Equal(t, hr.children[parent.Hash], []types.Hash{header.Hash}, "parent child not header hash")
		assert.Equal(t, hr.children[header.Hash], []types.Hash{child1.Hash, child2.Hash}, "header children not child hashes")
	}
	assert.Empty(t, hr.pending, "pending map not empty")
}

func TestHeadersHasExisting(t *testing.T) {

	// initialize the repository with required maps
	hr := &Headers{
		headers: make(map[types.Hash]*types.Header),
	}

	// create entities and set up state
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}

	header := &types.Header{Hash: hash1, Parent: hash2}

	hr.headers[header.Hash] = header

	// try adding header already known and check outcome
	ok := hr.Has(header.Hash)
	assert.True(t, ok, "could not confirm existing header")
}

func TestHeadersHasMissing(t *testing.T) {

	// initialize the repository with required maps
	hr := &Headers{
		headers: make(map[types.Hash]*types.Header),
	}

	// create entities and set up state
	hash1 := types.Hash{0x1}

	// try adding header already known and check outcome
	ok := hr.Has(hash1)
	assert.False(t, ok, "could not confirm missing header")
}

func TestHeadersGet(t *testing.T) {
}

func TestHeadersPath(t *testing.T) {
}
