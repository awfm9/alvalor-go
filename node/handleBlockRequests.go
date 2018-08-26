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

func handleBlockRequests(log zerolog.Logger, wg *sync.WaitGroup, net Network, requestsStream <-chan interface{}, stop <-chan struct{}) {
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
		requestsQueue := newBlockRequestsQueue(requestMessages)

		for addr, hashes := range requestsQueue.getData() {
			for _, hash := range hashes {
				confirm := &Confirm{Hash: hash}
				err := net.Send(addr, confirm)
				if err != nil {
					log.Error().Err(err).Msg("could not request transactions")
					return
				}
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
			case *blockRequest:
				{
					if hashAddr, ok := requestMessages[requestMsg.hash]; ok {
						requestMessages[requestMsg.hash] = append(hashAddr, requestMsg.addr)
					} else {
						requestMessages[requestMsg.hash] = []string{requestMsg.addr}
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
