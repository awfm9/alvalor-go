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

package headers

import (
	"testing"

	"github.com/alvalor/alvalor-go/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestRepoAddExisting(t *testing.T) {

	// initialize the repository with required maps
	hr := &Repo{
		headers: make(map[types.Hash]*types.Header),
	}

	// create entities and set up state
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}

	header := &types.Header{Hash: hash1, Parent: hash2}

	hr.headers[header.Hash] = &types.Header{}

	// try adding header already known and check outcome
	err := hr.Add(header)
	if assert.NotNil(t, err) {
		assert.Equal(t, ErrExist, errors.Cause(err))
	}
	if assert.Len(t, hr.headers, 1) {
		assert.NotEqual(t, hr.headers[header.Hash], header)
	}
}

func TestRepoAddPending(t *testing.T) {

	// initialize the repository with required maps
	hr := &Repo{
		headers: make(map[types.Hash]*types.Header),
		pending: make(map[types.Hash][]*types.Header),
	}

	// create entities and set up state
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}

	header := &types.Header{Hash: hash1, Parent: hash2}

	// try adding header with missing parent and check outcome
	err := hr.Add(header)
	assert.Nil(t, err)
	assert.Empty(t, hr.headers)
	if assert.Len(t, hr.pending, 1) {
		assert.ElementsMatch(t, hr.pending[header.Parent], []*types.Header{header})
	}
}

func TestRepoAddValid(t *testing.T) {

	// initialize the repository with required maps
	hr := &Repo{
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
	assert.Nil(t, err)
	if assert.Len(t, hr.headers, 2) {
		assert.Equal(t, hr.headers[header.Hash], header)
	}
	if assert.Len(t, hr.children, 1) {
		assert.ElementsMatch(t, hr.children[parent.Hash], []types.Hash{header.Hash})
	}
	assert.Empty(t, hr.pending)
}

func TestRepoAddValidWithPending(t *testing.T) {

	// initialize the repository with required maps
	hr := &Repo{
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
	assert.Nil(t, err)
	if assert.Len(t, hr.headers, 4) {
		assert.Equal(t, hr.headers[hash1], header)
		assert.Equal(t, hr.headers[hash3], child1)
		assert.Equal(t, hr.headers[hash4], child2)
	}
	if assert.Len(t, hr.children, 2) {
		assert.ElementsMatch(t, hr.children[parent.Hash], []types.Hash{header.Hash})
		assert.ElementsMatch(t, hr.children[header.Hash], []types.Hash{child1.Hash, child2.Hash})
	}
	assert.Empty(t, hr.pending)
}

func TestRepoHasExisting(t *testing.T) {

	// initialize the repository with required maps
	hr := &Repo{
		headers: make(map[types.Hash]*types.Header),
	}

	// create entities and set up state
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}

	header := &types.Header{Hash: hash1, Parent: hash2}

	hr.headers[header.Hash] = header

	// try adding header already known and check outcome
	ok := hr.Has(header.Hash)
	assert.True(t, ok)
}

func TestRepoHasMissing(t *testing.T) {

	// initialize the repository with required maps
	hr := &Repo{
		headers: make(map[types.Hash]*types.Header),
	}

	// create entities and set up state
	hash1 := types.Hash{0x1}

	// try adding header already known and check outcome
	ok := hr.Has(hash1)
	assert.False(t, ok)
}

func TestRepoGetExisting(t *testing.T) {

	// initialize the repository with required maps
	hr := &Repo{
		headers: make(map[types.Hash]*types.Header),
	}

	// create entities and set up state
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}

	header := &types.Header{Hash: hash1, Parent: hash2}

	hr.headers[header.Hash] = header

	// try adding header already known and check outcome
	output, err := hr.Get(header.Hash)
	if assert.Nil(t, err) {
		assert.Equal(t, header, output)
	}
}

func TestRepoGetMissing(t *testing.T) {

	// initialize the repository with required maps
	hr := &Repo{
		headers: make(map[types.Hash]*types.Header),
	}

	// create entities and set up state
	hash1 := types.Hash{0x1}

	// try adding header already known and check outcome
	_, err := hr.Get(hash1)
	if assert.NotNil(t, err) {
		assert.Equal(t, ErrNotExist, errors.Cause(err))
	}
}
