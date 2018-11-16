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

package assembly

import (
	"github.com/alvalor/alvalor-go/types"
	"github.com/pkg/errors"
)

// Manager is the manager to assemble and validate blocks.
type Manager struct {
	headers      Headers
	inventories  Inventories
	transactions Transactions
}

// Validate will assemble the block from our database and validate it.
func (am *Manager) Validate(hash types.Hash) error {

	// retrieve the header
	header, err := am.headers.Get(hash)
	if err != nil {
		return errors.Wrap(err, "could not retrieve header for block assembly")
	}

	// retrieve the inventory
	inv, err := am.inventories.Get(hash)
	if err != nil {
		return errors.Wrap(err, "could not retrieve inventory for block assembly")
	}

	// build block
	block := types.Block{
		Header: header,
	}
	for _, txHash := range inv.Hashes {
		tx, err := am.transactions.Get(txHash)
		if err != nil {
			return errors.Wrapf(err, "could not retrieve transaction for block assembly (%v)", txHash)
		}
		block.Transactions = append(block.Transactions, tx)
	}

	// TODO: validate block

	return nil
}
