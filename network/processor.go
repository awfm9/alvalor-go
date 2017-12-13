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

package network

import (
	"sync"

	"github.com/rs/zerolog"
)

// Processor is all the dependencies for a processing routine.
type Processor interface {
	Peer(address string) (Peer, error)
	Protocol(version string) (Protocol, error)
	State() (State, error)
	Send(address string, message interface{}) error
}

func handleProcessing(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, mgr Processor, messages <-chan Message) {
	defer wg.Done()
	for message := range messages {
		state, err := mgr.State()
		if err != nil {
			log.Error().Err(err).Msg("could not get local state")
			continue
		}
		address := message.Address
		peer, err := mgr.Peer(address)
		if err != nil {
			log.Error().Str("address", address).Err(err).Msg("could not get peer")
			continue
		}
		version := peer.Version
		protocol, err := mgr.Protocol(peer.Version)
		if err != nil {
			log.Error().Str("address", address).Str("version", version).Err(err).Msg("could not get protocol")
			continue
		}
		responses, recipients, err := protocol.Process(message, state, peer)
		if err != nil {
			log.Error().Str("address", address).Err(err).Msg("could not process message")
			continue
		}
		for i, response := range responses {
			recipient := recipients[i]
			err := mgr.Send(recipient, response)
			if err != nil {
				log.Error().Str("address", address).Err(err).Msg("could not send message")
				continue
			}
		}
	}
}
