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

package download

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/alvalor/alvalor-go/node/handlers/message"
	"github.com/alvalor/alvalor-go/node/state/peers"
	"github.com/alvalor/alvalor-go/types"
)

// Manager implement a simple download manager.
type Manager struct {
	sync.Mutex
	net   Network
	peers Peers
	invs  map[types.Hash]string
	txs   map[types.Hash]string
}

// StartInv starts the download of a block inventory.
func (mgr *Manager) StartInv(hash types.Hash) error {

	// if we are already downloading the inventory, skip
	_, ok := mgr.invs[hash]
	if ok {
		return errors.New("inventory download already pending")
	}

	// get all active peers that have the desired entity
	has := mgr.peers.Addresses(peers.IsActive(true), peers.HasEntity(true, hash))
	may := mgr.peers.Addresses(peers.IsActive(true), peers.MayEntity(hash))
	if len(has) == 0 && len(may) == 0 {
		return errors.New("no active peers with entity available")
	}

	// send the request to the best candidate
	address := Select(has, may, mgr.count())
	msg := &message.GetInv{Hash: hash}
	err := mgr.net.Send(address, msg)
	if err != nil {
		return errors.Wrap(err, "could not send inventory request")
	}

	// mark the request as pending for this peer
	mgr.invs[hash] = address

	// TODO: add timeout & retry functionality

	return nil
}

// StartTx starts the download of a transaction.
func (mgr *Manager) StartTx(hash types.Hash) error {

	// if we are already downloading the transaction, skip
	_, ok := mgr.txs[hash]
	if ok {
		return errors.New("inventory download already pending")
	}

	// get all active peers that have the desired entity
	has := mgr.peers.Addresses(peers.IsActive(true), peers.HasEntity(true, hash))
	may := mgr.peers.Addresses(peers.IsActive(true), peers.MayEntity(hash))
	if len(has) == 0 && len(may) == 0 {
		return errors.New("no active peers with entity available")
	}

	// send the request to the best candidate
	address := Select(has, may, mgr.count())
	msg := &message.GetTx{Hash: hash}
	err := mgr.net.Send(address, msg)
	if err != nil {
		return errors.Wrap(err, "could not send inventory request")
	}

	// mark the request as pending for this peer
	mgr.txs[hash] = address

	// TODO: add timeout & retry functionality

	return nil
}

// counts returns the number of pending downloads per address.
func (mgr *Manager) count() map[string]uint {
	count := make(map[string]uint)
	for _, address := range mgr.invs {
		count[address]++
	}
	for _, address := range mgr.txs {
		count[address]++
	}
	return count
}

// HasInv checks whether we are currently trying to download an inventory.
func (mgr *Manager) HasInv(hash types.Hash) bool {
	_, ok := mgr.invs[hash]
	return ok
}

// HasTx checks whether we are currently trying to download a transaction.
func (mgr *Manager) HasTx(hash types.Hash) bool {
	_, ok := mgr.txs[hash]
	return ok
}

// CancelInv cancels the download of a block inventory.
func (mgr *Manager) CancelInv(hash types.Hash) error {

	// find which peer is currently pending for this download
	_, ok := mgr.invs[hash]
	if !ok {
		return errors.Wrap(ErrNotExist, "inventory download not found")
	}

	// TODO: cancel timeout and abort download

	// remove the pending entry
	delete(mgr.invs, hash)

	return nil
}

// CancelTx cancels the download of a block inventory.
func (mgr *Manager) CancelTx(hash types.Hash) error {

	// find which peer is currently pending for this download
	_, ok := mgr.txs[hash]
	if !ok {
		return errors.New("transaction download not found")
	}

	// TODO: cancel timeout and abort download

	// remove the pending entry
	delete(mgr.txs, hash)

	return nil
}
