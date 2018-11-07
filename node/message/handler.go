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

package message

import (
	"sync"

	"github.com/alvalor/alvalor-go/types"
	"github.com/rs/zerolog"
)

// Handler represents the handler for messages from the network stack.
type Handler struct {
	log          zerolog.Logger
	net          Network
	paths        Paths
	downloads    Downloads
	headers      Headers
	inventories  Inventories
	transactions Transactions
	peers        Peers
	entity       Entity
}

// Process processes a message from the network.
func (handler *Handler) Process(wg *sync.WaitGroup, address string, message interface{}) {
	wg.Add(1)
	switch msg := message.(type) {
	case *Status:
		go handler.processStatus(wg, address, msg)
	case *Sync:
		go handler.processSync(wg, address, msg)
	case *Path:
		go handler.processPath(wg, address, msg)
	case *GetInv:
		go handler.processGetInv(wg, address, msg)
	case *GetTx:
		go handler.processGetTx(wg, address, msg)
	case *types.Inventory:
		go handler.processInventory(wg, address, msg)
	case *types.Transaction:
		go handler.processTransaction(wg, address, msg)
	}
}
