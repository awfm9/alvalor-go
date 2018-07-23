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

	assert.Equal(t, root.Hash, sp.root, "root header hash not saved")
	if assert.NotEmpty(t, sp.headers, "root header not saved") {
		assert.Equal(t, root, sp.headers[root.Hash], "root header not in map")
	}
}

func TestSimplePathAdd(t *testing.T) {

	header1 := &types.Header{Hash: types.Hash{1}}
	header2 := &types.Header{Hash: types.Hash{2}, Parent: types.Hash{1}}
	header3 := &types.Header{Hash: types.Hash{3}, Parent: types.Hash{2}}
	header4 := &types.Header{Hash: types.Hash{4}}

	sp := simplePath{
		headers:  map[types.Hash]*types.Header{header1.Hash: header1},
		children: make(map[types.Hash][]types.Hash),
	}

	err := sp.Add(header2)
	if assert.Nil(t, err, "adding header2 failed") {
		assert.Len(t, sp.headers, 2, "header not added to map")
	}

	err = sp.Add(header3)
	if assert.Nil(t, err, "adding header3 failed") {
		assert.Len(t, sp.headers, 3, "header not added to map")
	}

	err = sp.Add(header4)
	if assert.NotNil(t, err, "adding header4 succeeded") {
		assert.Len(t, sp.headers, 3, "header added to map")
	}

	err = sp.Add(header1)
	if assert.NotNil(t, err, "adding header1 succeeded") {
		assert.Len(t, sp.headers, 3, "header added to map")
	}
}

func TestSimplePathLongest(t *testing.T) {
}
