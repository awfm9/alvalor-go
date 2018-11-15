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

package paths

import (
	"testing"

	"github.com/alvalor/alvalor-go/types"
	"github.com/stretchr/testify/assert"
)

func TestDiff(t *testing.T) {

	// initialize parameters
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}
	hash3 := types.Hash{0x3}

	// initialize vectors
	vectors := map[string]struct {
		old    []types.Hash
		new    []types.Hash
		cancel []types.Hash
		start  []types.Hash
	}{
		"only_new": {
			old:    []types.Hash{},
			new:    []types.Hash{hash1, hash2, hash3},
			cancel: []types.Hash{},
			start:  []types.Hash{hash1, hash2, hash3},
		},
		"only_old": {
			old:    []types.Hash{hash1, hash2, hash3},
			new:    []types.Hash{},
			cancel: []types.Hash{hash1, hash2, hash3},
			start:  []types.Hash{},
		},
		"only_cancel": {
			old:    []types.Hash{hash1, hash2, hash3},
			new:    []types.Hash{hash1, hash3},
			cancel: []types.Hash{hash2},
			start:  []types.Hash{},
		},
		"only_start": {
			old:    []types.Hash{hash1, hash3},
			new:    []types.Hash{hash1, hash2, hash3},
			cancel: []types.Hash{},
			start:  []types.Hash{hash2},
		},
		"start_cancel": {
			old:    []types.Hash{hash2, hash3},
			new:    []types.Hash{hash1, hash2},
			cancel: []types.Hash{hash3},
			start:  []types.Hash{hash1},
		},
		"identity": {
			old:    []types.Hash{hash1, hash2, hash3},
			new:    []types.Hash{hash1, hash3, hash2},
			cancel: []types.Hash{},
			start:  []types.Hash{},
		},
	}

	// iterate vectors
	for name, vector := range vectors {

		// execute diff
		cancel, start := Diff(vector.old, vector.new)

		// check conditions
		assert.ElementsMatch(t, vector.cancel, cancel, name)
		assert.ElementsMatch(t, vector.start, start, name)
	}
}
