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
	"sync"
	"time"

	"github.com/alvalor/alvalor-go/types"
	"github.com/rs/zerolog"
)

func handleTransactionRequests(log zerolog.Logger, wg *sync.WaitGroup, net Network, requestsStream <-chan interface{}, stop <-chan struct{}) {
	defer wg.Done()

	log = log.With().Str("component", "transaction requests").Logger()

	ticker := time.NewTicker(time.Millisecond * 100)
	for {
		select {
		case <-stop:
			ticker.Stop()
			return
		case <-ticker.C:
		}

		requestMessages := getRequestMessages(requestsStream)
		outgoingRequests := transformToOutgoingRequests(requestMessages)

		for addr, hashes := range outgoingRequests {
			request := &Request{Hashes: hashes}
			err := net.Send(addr, request)
			if err != nil {
				log.Error().Err(err).Msg("could not request transactions")
				return
			}
		}
	}
}

func getRequestMessages(requestsStream <-chan interface{}) map[types.Hash][]string {
	requestMessages := make(map[types.Hash][]string)
	ticker := time.NewTicker(time.Second * 1)
Loop:
	for {
		select {
		case <-ticker.C:
			//Collect request messages only for 1 second
			break Loop
		case msg, ok := <-requestsStream:
			if !ok {
				break Loop
			}

			switch requestMsg := msg.(type) {
			case *internalTransactionRequest:
				{
					for _, hash := range requestMsg.hashes {
						if hashAddr, ok := requestMessages[hash]; ok {
							requestMessages[hash] = append(hashAddr, requestMsg.addr)
							continue
						}

						requestMessages[hash] = []string{requestMsg.addr}
					}
				}
			}
		case <-time.After(100 * time.Millisecond):
			//Only wait for transaction requests for 100 ms
			break Loop
		}
	}
	return requestMessages
}

// tranforms map of [tx hashes -> peers addr which have it] to the map of [peer addr -> tx hashes to request]
func transformToOutgoingRequests(requests map[types.Hash][]string) map[string][]types.Hash {
	result := make(map[string][]types.Hash)

	//map where key is an address and value is amount of planned hashes to send
	outgoingTxReqCountMap := make(map[string]int)
	for hash, addresses := range requests {
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
