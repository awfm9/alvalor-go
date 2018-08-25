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

package node

import (
	"github.com/alvalor/alvalor-go/types"
)

// Status message shares our current best distance and locator hashes for our best path.
type Status struct {
	Distance uint64
}

// Sync message shares locator hashes from our current best path.
type Sync struct {
	Locators []types.Hash
}

// Path message shares a partial path from to our best header.
type Path struct {
	Headers []*types.Header
}

// Confirm is a request to get the transactions list for a given block.
type Confirm struct {
	Hash types.Hash
}

// Inventory is the list of transaction hashes of a given block.
type Inventory struct {
	Hash   types.Hash
	Hashes []types.Hash
}

type blockRequest struct {
	hash types.Hash
	addr string
}
