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
// GNU Affero General Public License for more detailb.
//
// You should have received a copy of the GNU Affero General Public License
// along with Alvalor.  If not, see <http://www.gnu.org/licenses/>.

package node

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alvalor/alvalor-go/types"
)

func TestNewSimplePathfinder(t *testing.T) {
	root := &types.Header{Hash: types.Hash{1}}
	sp := newSimplePathfinder(root)
	assert.NotNil(t, sp.headers, "header map not initialized")
	assert.NotNil(t, sp.children, "children map not initialized")
	assert.NotNil(t, sp.pending, "pending map not initialized")
	assert.Equal(t, root.Hash, sp.root, "root hash not saved")
	if assert.NotEmpty(t, sp.headers, "root header not saved") {
		assert.Equal(t, root, sp.headers[root.Hash], "root header not at correct index")
	}
}

func TestSimplPathfinderKnows(t *testing.T) {

	hash1 := types.Hash{1}
	hash2 := types.Hash{2}

	sp := &simplePathfinder{
		headers: make(map[types.Hash]*types.Header),
	}
	sp.headers[hash1] = &types.Header{}

	ok := sp.Knows(hash1)
	assert.True(t, ok, "hash one not known")

	ok = sp.Knows(hash2)
	assert.False(t, ok, "hash two known")
}

func TestSimplePathfinderHeader(t *testing.T) {

	hash1 := types.Hash{1}
	hash2 := types.Hash{2}

	header1 := &types.Header{Hash: hash1}

	sp := &simplePathfinder{
		headers: make(map[types.Hash]*types.Header),
	}
	sp.headers[header1.Hash] = header1

	result, err := sp.Header(hash1)
	if assert.Nil(t, err, "could not retrieve first header") {
		assert.Equal(t, header1, result, "retrieved header not equal")
	}

	_, err = sp.Header(hash2)
	assert.NotNil(t, err, "could retrieve second header")
}

func TestSimplePathAddExistingHash(t *testing.T) {

	header := &types.Header{Hash: types.Hash{1}, Parent: types.Hash{0}}

	sp := &simplePathfinder{
		headers: make(map[types.Hash]*types.Header),
	}

	sp.headers[header.Parent] = &types.Header{}
	sp.headers[header.Hash] = &types.Header{}

	err := sp.Add(header)
	assert.NotNil(t, err, "could insert duplicate header")
}

func TestSimplePathAddMissingParent(t *testing.T) {

	header := &types.Header{Hash: types.Hash{1}, Parent: types.Hash{0}}

	sp := &simplePathfinder{
		headers: make(map[types.Hash]*types.Header),
		pending: make(map[types.Hash][]*types.Header),
	}

	err := sp.Add(header)
	if assert.Nil(t, err, "could not insert with missing parent") {
		assert.NotEmpty(t, sp.pending, "header missing from pending")
		assert.Contains(t, sp.pending[header.Parent], header, "header not correctly listed in pending")
	}
}

func TestSimplePathAddValidHeader(t *testing.T) {

	header1 := &types.Header{Hash: types.Hash{1}, Parent: types.Hash{0}}
	header2 := &types.Header{Hash: types.Hash{2}, Parent: types.Hash{1}}

	sp := &simplePathfinder{
		headers:  make(map[types.Hash]*types.Header),
		children: make(map[types.Hash][]types.Hash),
		pending:  make(map[types.Hash][]*types.Header),
	}

	sp.headers[header1.Hash] = header1
	err := sp.Add(header2)

	assert.Nil(t, err, "could not insert valid header")
	assert.Len(t, sp.headers, 2, "header map wrong size")
	if assert.Contains(t, sp.headers, header2.Hash, "second header not saved") {
		assert.Equal(t, sp.headers[header2.Hash], header2, "second header incorrect")
	}
	assert.Len(t, sp.children, 1, "children map wrong size")
	if assert.Contains(t, sp.children, header2.Parent, "header child not saved") {
		assert.Contains(t, sp.children[header2.Parent], header2.Hash, "header child hash not saved")
	}
}

func TestSimplePathLongestRootOnly(t *testing.T) {

	header := &types.Header{Hash: types.Hash{1}, Parent: types.Hash{0}, Diff: 1}

	sp := newSimplePathfinder(header)

	path, _ := sp.Longest()

	assert.Equal(t, path, []types.Hash{header.Hash})
}

func TestSimplePathLongestLinearOnly(t *testing.T) {

	header1 := &types.Header{Hash: types.Hash{1}, Parent: types.Hash{0}, Diff: 1}
	header2 := &types.Header{Hash: types.Hash{2}, Parent: types.Hash{1}, Diff: 1}
	header3 := &types.Header{Hash: types.Hash{3}, Parent: types.Hash{2}, Diff: 1}
	header4 := &types.Header{Hash: types.Hash{4}, Parent: types.Hash{3}, Diff: 1}

	sp := newSimplePathfinder(header1)
	_ = sp.Add(header2)
	_ = sp.Add(header3)
	_ = sp.Add(header4)

	path, _ := sp.Longest()

	assert.Equal(t, path, []types.Hash{header4.Hash, header3.Hash, header2.Hash, header1.Hash})
}

func TestSimplePathLongestShortHeavy(t *testing.T) {

	header1 := &types.Header{Hash: types.Hash{1}, Parent: types.Hash{0}, Diff: 1}
	header2 := &types.Header{Hash: types.Hash{2}, Parent: types.Hash{1}, Diff: 1}
	header3 := &types.Header{Hash: types.Hash{3}, Parent: types.Hash{2}, Diff: 1}
	header4 := &types.Header{Hash: types.Hash{4}, Parent: types.Hash{3}, Diff: 1}
	header5 := &types.Header{Hash: types.Hash{5}, Parent: types.Hash{1}, Diff: 10}

	sp := newSimplePathfinder(header1)
	_ = sp.Add(header2)
	_ = sp.Add(header3)
	_ = sp.Add(header4)
	_ = sp.Add(header5)

	path, _ := sp.Longest()

	assert.Equal(t, path, []types.Hash{header5.Hash, header1.Hash})

}

func TestSimplePathLongestLongHeavy(t *testing.T) {

	header1 := &types.Header{Hash: types.Hash{1}, Parent: types.Hash{0}, Diff: 1}
	header2 := &types.Header{Hash: types.Hash{2}, Parent: types.Hash{1}, Diff: 5}
	header3 := &types.Header{Hash: types.Hash{3}, Parent: types.Hash{2}, Diff: 5}
	header4 := &types.Header{Hash: types.Hash{4}, Parent: types.Hash{3}, Diff: 5}
	header5 := &types.Header{Hash: types.Hash{5}, Parent: types.Hash{1}, Diff: 10}

	sp := newSimplePathfinder(header1)
	_ = sp.Add(header2)
	_ = sp.Add(header3)
	_ = sp.Add(header4)
	_ = sp.Add(header5)

	path, _ := sp.Longest()

	assert.Equal(t, path, []types.Hash{header4.Hash, header3.Hash, header2.Hash, header1.Hash})
}

func TestSimplePathLongestEqualHeavy(t *testing.T) {

	header1 := &types.Header{Hash: types.Hash{1}, Parent: types.Hash{0}, Diff: 1}

	header2 := &types.Header{Hash: types.Hash{2}, Parent: types.Hash{1}, Diff: 4}
	header3 := &types.Header{Hash: types.Hash{3}, Parent: types.Hash{1}, Diff: 16}
	header4 := &types.Header{Hash: types.Hash{4}, Parent: types.Hash{1}, Diff: 8}
	header5 := &types.Header{Hash: types.Hash{5}, Parent: types.Hash{1}, Diff: 32}

	header6 := &types.Header{Hash: types.Hash{6}, Parent: types.Hash{2}, Diff: 4}
	header7 := &types.Header{Hash: types.Hash{7}, Parent: types.Hash{2}, Diff: 16}
	header8 := &types.Header{Hash: types.Hash{8}, Parent: types.Hash{3}, Diff: 8}
	header9 := &types.Header{Hash: types.Hash{9}, Parent: types.Hash{3}, Diff: 32}
	header10 := &types.Header{Hash: types.Hash{10}, Parent: types.Hash{4}, Diff: 4}
	header11 := &types.Header{Hash: types.Hash{11}, Parent: types.Hash{4}, Diff: 16}
	header12 := &types.Header{Hash: types.Hash{12}, Parent: types.Hash{5}, Diff: 8}
	header13 := &types.Header{Hash: types.Hash{13}, Parent: types.Hash{5}, Diff: 32}

	sp := newSimplePathfinder(header1)
	_ = sp.Add(header2)
	_ = sp.Add(header3)
	_ = sp.Add(header4)
	_ = sp.Add(header5)
	_ = sp.Add(header6)
	_ = sp.Add(header7)
	_ = sp.Add(header8)
	_ = sp.Add(header9)
	_ = sp.Add(header10)
	_ = sp.Add(header11)
	_ = sp.Add(header12)
	_ = sp.Add(header13)

	path, _ := sp.Longest()

	assert.Equal(t, path, []types.Hash{header13.Hash, header5.Hash, header1.Hash})
}
