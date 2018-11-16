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

// State represents the state of the currently followed path.
type State struct {
	current []types.Hash
}

// Current returns the current path.
func (st *State) Current() []types.Hash {
	return st.current
}

// Set sets the path to be followed and returns the deltas between old and new.
func (st *State) Set(path []types.Hash) ([]types.Hash, []types.Hash) {
	cancel, start := Diff(st.current, path)
	st.current = path
	return cancel, start
}
