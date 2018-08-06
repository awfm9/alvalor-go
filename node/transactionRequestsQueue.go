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

type transactionRequestsQueue interface {
	getData() map[string][]types.Hash
}

type simpleTransactionRequestsQueue struct {
	requests map[types.Hash][]string
}

func newTransactionRequestsQueue(requests map[types.Hash][]string) simpleTransactionRequestsQueue {
	return simpleTransactionRequestsQueue{requests: requests}
}

func (q *simpleTransactionRequestsQueue) getData() map[string][]types.Hash {
	data := q.transformToOutgoingRequests()
	return data
}

// tranforms map of [tx hashes -> peers addr which have it] to the map of [peer addr -> tx hashes to request]
func (q *simpleTransactionRequestsQueue) transformToOutgoingRequests() map[string][]types.Hash {
	result := make(map[string][]types.Hash)

	//map where key is an address and value is amount of planned hashes to send
	outgoingTxReqCountMap := make(map[string]int)
	for hash, addresses := range q.requests {
		//find address which has least amount of tx hashes to request
		minAddr := ""
		minAddrCount := 0
		for _, addr := range addresses {
			if count, ok := outgoingTxReqCountMap[addr]; ok {
				if minAddr == "" || count < minAddrCount {
					minAddr = addr
					minAddrCount = count
				}
				continue
			}
			//Found zero planned requests for this peer, therefore just selecting it
			minAddr = addr
			break
		}

		outgoingTxReqCountMap[minAddr] = minAddrCount + 1

		if outgoing, ok := result[minAddr]; ok {
			result[minAddr] = append(outgoing, hash)
		} else {
			result[minAddr] = []types.Hash{hash}
		}

	}
	return result
}
