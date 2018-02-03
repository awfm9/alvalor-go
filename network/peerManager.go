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
	"bytes"
	"net"
	"sync"

	"github.com/pierrec/lz4"
	"github.com/pkg/errors"
)

type peerManager interface {
	Add(conn net.Conn, nonce []byte) error
	Known(nonce []byte) bool
	Count() uint
	Drop(address string) error
	Addresses() []string
	DropAll()
}

type peer struct {
	conn   net.Conn
	input  chan interface{}
	output chan interface{}
	nonce  []byte
}

type simplePeerManager struct {
	sync.Mutex
	handlers handlerManager
	min      uint
	max      uint
	buffer   uint
	reg      map[string]*peer
}

func newSimplePeerManager(handlers handlerManager, min uint, max uint) *simplePeerManager {
	return &simplePeerManager{
		handlers: handlers,
		min:      min,
		max:      max,
		buffer:   16,
		reg:      make(map[string]*peer),
	}
}

func (pm *simplePeerManager) Add(conn net.Conn, nonce []byte) error {
	pm.Lock()
	defer pm.Unlock()

	// make sure we can still add peers
	if uint(len(pm.reg)) >= pm.max {
		return errors.New("maximum number of peers reached")
	}

	// check if we already know the peer
	address := conn.RemoteAddr().String()
	_, ok := pm.reg[address]
	if ok {
		return errors.New("peer with nonce already known")
	}

	// initialize the peer
	p := &peer{
		conn:   conn,
		input:  make(chan interface{}, pm.buffer),
		output: make(chan interface{}, pm.buffer),
		nonce:  nonce,
	}

	// initialize the readers and writers
	r := lz4.NewReader(conn)
	w := lz4.NewWriter(conn)

	// launch the message processing routines
	// TODO: figure out if it's cleaner & better for testing to pass address & rest or just conn
	pm.handlers.Send(address, p.output, w)
	pm.handlers.Process(address, p.input, p.output)
	pm.handlers.Receive(address, r, p.input)

	pm.reg[address] = p
	return nil
}

func (pm *simplePeerManager) DropAll() {
	pm.Lock()
	defer pm.Unlock()
	for address := range pm.reg {
		pm.Drop(address)
	}
}

func (pm *simplePeerManager) Drop(address string) error {
	pm.Lock()
	defer pm.Unlock()
	p, ok := pm.reg[address]
	if !ok {
		return errors.New("peer unknown")
	}
	err := p.conn.Close()
	if err != nil {
		return errors.Wrap(err, "could not close peer connection")
	}
	return nil
}

func (pm *simplePeerManager) Count() uint {
	pm.Lock()
	defer pm.Unlock()
	return uint(len(pm.reg))
}

func (pm *simplePeerManager) Known(nonce []byte) bool {
	pm.Lock()
	defer pm.Unlock()
	for _, p := range pm.reg {
		if bytes.Equal(p.nonce, nonce) {
			return true
		}
	}
	return false
}

func (pm *simplePeerManager) Addresses() []string {
	pm.Lock()
	defer pm.Unlock()
	addresses := make([]string, 0, len(pm.reg))
	for _, peer := range pm.reg {
		addresses = append(addresses, peer.conn.RemoteAddr().String())
	}
	return addresses
}
