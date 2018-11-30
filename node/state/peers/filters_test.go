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

package peers

import (
	"testing"

	"github.com/alvalor/alvalor-go/types"
	"github.com/stretchr/testify/assert"
)

func TestHasEntity(t *testing.T) {

	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}
	hash3 := types.Hash{0x3}

	vectors := map[string]struct {
		has  EntityHas
		hash types.Hash
		yes  map[types.Hash]struct{}
		no   map[types.Hash]struct{}
		ok   bool
	}{
		"yes_present": {
			has:  EntityYes,
			hash: hash1,
			yes:  map[types.Hash]struct{}{hash1: struct{}{}},
			no:   map[types.Hash]struct{}{hash2: struct{}{}},
			ok:   true,
		},
		"yes_absent": {
			has:  EntityYes,
			hash: hash2,
			yes:  map[types.Hash]struct{}{hash1: struct{}{}},
			no:   map[types.Hash]struct{}{hash2: struct{}{}},
			ok:   false,
		},
		"no_present": {
			has:  EntityNo,
			hash: hash2,
			yes:  map[types.Hash]struct{}{hash1: struct{}{}},
			no:   map[types.Hash]struct{}{hash2: struct{}{}},
			ok:   true,
		},
		"no_absent": {
			has:  EntityNo,
			hash: hash1,
			yes:  map[types.Hash]struct{}{hash1: struct{}{}},
			no:   map[types.Hash]struct{}{hash2: struct{}{}},
			ok:   false,
		},
		"maybe_none": {
			has:  EntityMaybe,
			hash: hash3,
			yes:  map[types.Hash]struct{}{hash1: struct{}{}},
			no:   map[types.Hash]struct{}{hash2: struct{}{}},
			ok:   true,
		},
		"maybe_yes": {
			has:  EntityMaybe,
			hash: hash1,
			yes:  map[types.Hash]struct{}{hash1: struct{}{}},
			no:   map[types.Hash]struct{}{hash2: struct{}{}},
			ok:   true,
		},
		"maybe_no": {
			has:  EntityMaybe,
			hash: hash2,
			yes:  map[types.Hash]struct{}{hash1: struct{}{}},
			no:   map[types.Hash]struct{}{hash2: struct{}{}},
			ok:   false,
		},
	}

	for name, vector := range vectors {
		p := &Peer{
			yes: vector.yes,
			no:  vector.no,
		}
		ok := HasEntity(vector.has, vector.hash)(p)
		assert.Equal(t, vector.ok, ok, name)
	}
}

func TestIsActive(t *testing.T) {
}
