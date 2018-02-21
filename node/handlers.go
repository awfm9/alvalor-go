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

	"github.com/rs/zerolog"
)

type handlerManager interface {
	Process(entity Entity)
	Propagate(entity Entity)
}

type simpleHandlerManager struct {
	log   zerolog.Logger
	wg    *sync.WaitGroup
	net   networkManager
	state stateManager
	pool  poolManager
}

func (hm *simpleHandlerManager) Process(entity Entity) {
	hm.wg.Add(1)
	go handleProcessing(hm.log, hm.wg, hm.pool, hm.state, hm, entity)
}

func (hm *simpleHandlerManager) Propagate(entity Entity) {
	hm.wg.Add(1)
	go handlePropagating(hm.log, hm.wg, hm.state, hm.net, entity)
}
