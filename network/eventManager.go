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
	"time"

	"github.com/rs/zerolog/log"
)

type eventManager interface {
	Disconnected(addr string)
	Connected(addr string)
	Received(addr string, msg interface{})
}

type simpleEventManager struct {
	subscriber chan interface{}
}

func (mgr *simpleEventManager) Disconnected(address string) {
	mgr.subscriber <- Disconnected{Address: address, Timestamp: time.Now()}
}

func (mgr *simpleEventManager) Connected(address string) {
	mgr.subscriber <- Connected{Address: address, Timestamp: time.Now()}
}

func (mgr *simpleEventManager) Received(address string, msg interface{}) {
	select {
	case mgr.subscriber <- Received{Address: address, Message: msg, Timestamp: time.Now()}:
	default:
		log.Debug().Msg("subscriber stalling")
	}
}
