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
	"github.com/stretchr/testify/assert"
)

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
