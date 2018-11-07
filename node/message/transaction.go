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
)

// The transaction message is a message containing a transaction.
func (handler *Handler) processTransaction(wg *sync.WaitGroup, address string, tx *types.Transaction) {
	defer wg.Done()

	// configure logger
	with := handler.log.With()
	with.Str("component", "message")
	with.Str("message_type", "transaction")
	with.Str("address", address)
	with.Hex("hash", tx.Hash[:])
	log := with.Logger()

	// wrap routine in start and stop messages
	log.Debug().Msg("routine started")
	defer log.Debug().Msg("routine stopped")

	// cancel any pending download retries for this transaction
	handler.downloads.Cancel(tx.Hash)

	// mark the inventory download as completed for the respective peer
	handler.peers.Received(address, tx.Hash)

	// handle the transaction entity
	handler.entity.Process(wg, tx)

	log.Debug().Msg("processed transaction message")
}
