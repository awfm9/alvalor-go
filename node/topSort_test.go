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

	"github.com/alvalor/alvalor-go/types"
)

func TestDistances(t *testing.T) {
	root := types.Header{Hash: [32]byte{0x1}, Parent: types.ZeroHash}
	headers := []types.Header{
		{Hash: [32]byte{0x2}, Parent: [32]byte{0x1}, Diff: 10},
		{Hash: [32]byte{0x3}, Parent: [32]byte{0x2}, Diff: 10},
		{Hash: [32]byte{0x4}, Parent: [32]byte{0x3}, Diff: 10},
		{Hash: [32]byte{0x5}, Parent: [32]byte{0x4}, Diff: 10},
		{Hash: [32]byte{0x6}, Parent: [32]byte{0x5}, Diff: 10},
		{Hash: [32]byte{0x7}, Parent: [32]byte{0x6}, Diff: 10},
	}
	g := NewGraph(&root)
	for _, header := range headers {
		err := g.AddHeader(&header)
		if err != nil {
			t.Fatal(err)
		}
	}
	distances := g.LongestPath()
	t.Log(distances)
}
