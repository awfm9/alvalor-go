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
	"github.com/willf/bloom"

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

// Path message shares the partial path from a header hash to our best header.
type Path struct {
	Hashes []types.Hash
}

// RequestHeaders requests a number of headers for download.
type RequestHeaders struct {
	Hashes []types.Hash
}

// Headers shares a number of headers.
type Headers struct {
	Headers []*types.Header
}

// Mempool is a message containing details about the memory pool.
type Mempool struct {
	Bloom *bloom.BloomFilter
}

// Inventory is a message containing a list of transaction hashes.
type Inventory struct {
	Hashes []types.Hash
}

// Request requests a number of transactions for the memory pool.
type Request struct {
	Hashes []types.Hash
}

// Batch is a batch of transactions to send as one message.
type Batch struct {
	Transactions []*types.Transaction
}

type internalTransactionRequest struct {
	hashes []types.Hash
	addr   string
}
