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

	hr.headers[header.Hash] = header

	// try adding header already known and check outcome
	err := hr.Add(header)
	if assert.NotNil(t, err) {
		assert.Equal(t, ErrExist, errors.Cause(err))
	}
	if assert.Len(t, hr.headers, 1) {
		assert.Equal(t, hr.headers[header.Hash], header)
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

func TestRepoPathRoot(t *testing.T) {

	// root
	hash0 := types.Hash{0x6}

	// first level
	hash1 := types.Hash{0x1}

	// second level
	hash11 := types.Hash{0x11}
	hash12 := types.Hash{0x12}

	// third level
	hash111 := types.Hash{0x11, 0x1}
	hash121 := types.Hash{0x12, 0x1}
	hash122 := types.Hash{0x12, 0x2}

	// fourth level
	hash1111 := types.Hash{0x11, 0x11}
	hash1211 := types.Hash{0x12, 0x11}
	hash1212 := types.Hash{0x12, 0x12}

	// fifth level
	hash11111 := types.Hash{0x11, 0x11, 0x1}

	// initialize root
	header0 := &types.Header{Hash: hash0, Diff: 1} // 1

	// initialize level one
	header1 := &types.Header{Hash: hash1, Parent: hash0, Diff: 10} // 11

	// initialize level two
	header11 := &types.Header{Hash: hash11, Parent: hash1, Diff: 10} // 31
	header12 := &types.Header{Hash: hash12, Parent: hash1, Diff: 20} // 41

	// initialize level three
	header111 := &types.Header{Hash: hash111, Parent: hash11, Diff: 10} // 41
	header121 := &types.Header{Hash: hash121, Parent: hash12, Diff: 10} // 51
	header122 := &types.Header{Hash: hash122, Parent: hash12, Diff: 20} // 61

	// initialize level four
	header1111 := &types.Header{Hash: hash1111, Parent: hash111, Diff: 10} // 51
	header1211 := &types.Header{Hash: hash1211, Parent: hash121, Diff: 10} // 61
	header1212 := &types.Header{Hash: hash1212, Parent: hash121, Diff: 20} // 71

	// initialize level five
	header11111 := &types.Header{Hash: hash11111, Parent: hash1111, Diff: 50} // 91

	vectors := map[string]struct {
		headers  []*types.Header
		path     []types.Hash
		distance uint64
	}{
		"empty": {
			headers: []*types.Header{
				header0,
			},
			path: []types.Hash{
				hash0,
			},
			distance: 1,
		},
		"level_one": {
			headers: []*types.Header{
				header0,
				header1,
			},
			path: []types.Hash{
				hash1,
				hash0,
			},
			distance: 11,
		},
		"level_two": {
			headers: []*types.Header{
				header0,
				header1,
				header11,
				header12,
			},
			path: []types.Hash{
				hash12,
				hash1,
				hash0,
			},
			distance: 31,
		},
		"level_three": {
			headers: []*types.Header{
				header0,
				header1,
				header11,
				header12,
				header111,
				header121,
				header122,
			},
			path: []types.Hash{
				hash122,
				hash12,
				hash1,
				hash0,
			},
			distance: 51,
		},
		"level_four": {
			headers: []*types.Header{
				header0,
				header1,
				header11,
				header12,
				header111,
				header121,
				header122,
				header1111,
				header1211,
				header1212,
			},
			path: []types.Hash{
				hash1212,
				hash121,
				hash12,
				hash1,
				hash0,
			},
			distance: 61,
		},
		"level_five": {
			headers: []*types.Header{
				header0,
				header1,
				header11,
				header12,
				header111,
				header121,
				header122,
				header1111,
				header1211,
				header1212,
				header11111,
			},
			path: []types.Hash{
				hash11111,
				hash1111,
				hash111,
				hash11,
				hash1,
				hash0,
			},
			distance: 91,
		},
	}

	// loop through the test vectors
	for name, vector := range vectors {

		// initialize the repository with required maps
		hr := &Repo{
			root:     hash0,
			headers:  make(map[types.Hash]*types.Header),
			pending:  make(map[types.Hash][]*types.Header),
			children: make(map[types.Hash][]types.Hash),
		}

		// add the first header as root
		hr.headers[vector.headers[0].Hash] = vector.headers[0]

		// add the rest of the headers
		for i := 1; i < len(vector.headers); i++ {
			_ = hr.Add(vector.headers[i])
		}

		// get the distance/path and compare
		path, distance := hr.Path()
		assert.Equal(t, vector.path, path, name)
		assert.Equal(t, vector.distance, distance, name)
	}

}
