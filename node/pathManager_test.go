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

func TestNewSimplePath(t *testing.T) {
	root := &types.Header{Hash: types.Hash{1}}
	sp := newSimplePaths(root)
	assert.NotNil(t, sp.headers, "header map not initialized")
	assert.NotNil(t, sp.children, "children map not initialized")
	assert.Equal(t, root.Hash, sp.root, "root hash not saved")
	if assert.NotEmpty(t, sp.headers, "root header not saved") {
		assert.Equal(t, root, sp.headers[root.Hash], "root header not at correct index")
	}
}

func TestSimplePathAddExistingHash(t *testing.T) {

	header := &types.Header{Hash: types.Hash{1}, Parent: types.Hash{0}}

	sp := &simplePath{
		headers:  make(map[types.Hash]*types.Header),
		children: make(map[types.Hash][]types.Hash),
	}

	sp.headers[header.Parent] = &types.Header{}
	sp.headers[header.Hash] = &types.Header{}
	err := sp.Add(header)

	assert.NotNil(t, err, "could insert duplicate header")
}

func TestSimplePathAddMissingParent(t *testing.T) {

	header := &types.Header{Hash: types.Hash{1}, Parent: types.Hash{0}}

	sp := &simplePath{
		headers:  make(map[types.Hash]*types.Header),
		children: make(map[types.Hash][]types.Hash),
	}

	err := sp.Add(header)

	assert.NotNil(t, err, "could insert with missing parent")
}

func TestSimplePathAddValidHeader(t *testing.T) {

	header1 := &types.Header{Hash: types.Hash{1}, Parent: types.Hash{0}}
	header2 := &types.Header{Hash: types.Hash{2}, Parent: types.Hash{1}}

	sp := &simplePath{
		headers:  make(map[types.Hash]*types.Header),
		children: make(map[types.Hash][]types.Hash),
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

func TestSimplePathLongest(t *testing.T) {
}
