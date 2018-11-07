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

func (handler *Handler) processGetTx(wg *sync.WaitGroup, address string, getTx *GetTx) {
	defer wg.Done()

	// configure logger
	log := handler.log.With().Str("component", "message").Str("address", address).Logger()
	log = log.With().Str("msg_type", "get_tx").Hex("hash", getTx.Hash[:]).Logger()
	log.Debug().Msg("message routine started")
	defer log.Debug().Msg("message routine stopped")

	// try to get the inventory
	tx, err := handler.transactions.Get(getTx.Hash)
	if err != nil {
		log.Error().Err(err).Msg("could not get transaction")
		return
	}

	// try to send the inventory
	err = handler.net.Send(address, tx)
	if err != nil {
		log.Error().Err(err).Msg("could not send transaction")
		return
	}

	log.Debug().Msg("processed get_tx message")
}

func (handler *Handler) processTransaction(wg *sync.WaitGroup, address string, tx *types.Transaction) {
	defer wg.Done()

	// configure logger
	log := handler.log.With().Str("component", "message").Str("address", address).Logger()
	log = log.With().Str("msg_type", "transaction").Hex("hash", tx.Hash[:]).Logger()
	log.Debug().Msg("message routine started")
	defer log.Debug().Msg("message routine stopped")

	// cancel any pending download retries for this transaction
	handler.downloads.Cancel(tx.Hash)

	// mark the inventory download as completed for the respective peer
	handler.peers.Received(address, tx.Hash)

	// handle the transaction entity
	handler.entity.Process(wg, tx)

	log.Debug().Msg("processed transaction message")
}
