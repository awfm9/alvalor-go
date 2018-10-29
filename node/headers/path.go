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

import "github.com/alvalor/alvalor-go/types"

// Path returns the best path of the graph by total difficulty.
func (hr *Repo) Path() ([]types.Hash, uint64) {

	// create a topological sort and get distance for each header
	var hash types.Hash
	var max uint64
	var best types.Hash
	sorted := make([]types.Hash, 0, len(hr.headers))
	queue := []types.Hash{hr.root}
	distances := make(map[types.Hash]uint64)
	for len(queue) > 0 {
		hash, queue = queue[0], queue[1:]
		sorted = append(sorted, hash)
		queue = append(queue, hr.children[hash]...)
		header := hr.headers[hash]
		distance := distances[header.Parent] + header.Diff
		if distance > max {
			max = distance
			best = hash
		}
		distances[hash] = distance
	}

	// iterate back to parents from best child
	header := hr.headers[best]
	path := []types.Hash{header.Hash}
	for header.Parent != types.ZeroHash {
		header = hr.headers[header.Parent]
		path = append(path, header.Hash)
	}

	return path, max
}
