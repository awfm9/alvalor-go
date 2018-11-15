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

package path

import "github.com/alvalor/alvalor-go/types"

// Diff gives us the deltas between an old and a new path of headers.
func Diff(old []types.Hash, new []types.Hash) ([]types.Hash, []types.Hash) {

	// create a lookup map for the old path
	lookupOld := make(map[types.Hash]struct{}, len(old))
	for _, hash := range old {
		lookupOld[hash] = struct{}{}
	}

	// create a lookup map for the new path
	lookupNew := make(map[types.Hash]struct{}, len(new))
	for _, hash := range new {
		lookupNew[hash] = struct{}{}
	}

	// find the hashes on the old path that are not on the new one (cancel)
	var cancel []types.Hash
	for _, hash := range old {
		_, ok := lookupNew[hash]
		if ok {
			continue
		}
		cancel = append(cancel, hash)
	}

	// find the hashes on the new path that are not on the old one (start)
	var start []types.Hash
	for _, hash := range new {
		_, ok := lookupOld[hash]
		if ok {
			continue
		}
		start = append(start, hash)
	}

	return cancel, start
}
