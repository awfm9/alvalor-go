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

package download

import "math"

// Select will return the best download candidate from a list of candidates
// who certainly have or possibly have an entity.
func Select(has []string, may []string, count map[string]uint) string {

	// decide whether we select from certain or potential candidates
	candidates := has
	if len(candidates) == 0 {
		candidates = may
	}

	// select the available peer with the least amount of pending download
	var address string
	best := uint(math.MaxUint32)
	for _, candidate := range candidates {
		if count[candidate] >= best {
			continue
		}
		best = count[candidate]
		address = candidate
	}

	return address
}
