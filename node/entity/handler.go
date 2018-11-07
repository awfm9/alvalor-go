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

package entity

import (
	"sync"

	"github.com/alvalor/alvalor-go/types"
	"github.com/rs/zerolog"
)

// Handler is the handler for entities. We use a struct rather than a
// function so we can mock it easier for testing.
type Handler struct {
	log          zerolog.Logger
	net          Network
	paths        Paths
	events       Events
	headers      Headers
	transactions Transactions
	peers        Peers
}

// Process is the entity handler's function for processing a new entity.
func (handler *Handler) Process(wg *sync.WaitGroup, entity types.Entity) {
	wg.Add(1)
	switch e := entity.(type) {
	case *types.Header:
		go handler.processHeader(wg, e)
	case *types.Transaction:
		go handler.processTransaction(wg, e)
	}
}
