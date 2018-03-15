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

package finder

import "github.com/alvalor/alvalor-go/types"

func path(n *node) []types.Hash {
	if len(n.children) == 0 {
		return []types.Hash{n.hash}
	}
	var best []types.Hash
	for _, child := range n.children {
		p := path(child)
		if len(p) <= len(best) {
			continue
		}
		best = p
	}
	return append([]types.Hash{n.hash}, best...)
}
