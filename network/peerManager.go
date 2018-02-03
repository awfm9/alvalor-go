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
	"encoding/hex"
	"errors"
	"net"
	"sync"

	"github.com/pierrec/lz4"
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
	min    uint
	max    uint
	buffer uint
	reg    map[string]*peer
	lookup map[string]string
}

func newSimplePeerManager(min uint, max uint) *simplePeerManager {
	return &simplePeerManager{
		min:    min,
		max:    max,
		buffer: 16,
		reg:    make(map[string]*peer),
		lookup: make(map[string]string),
	}
}

func (pm *simplePeerManager) Add(conn net.Conn, nonce []byte) error {
	pm.Lock()
	defer pm.Unlock()
	if uint(len(pm.reg)) >= pm.max {
		return errors.New("maximum number of peers reached")
	}
	address := conn.RemoteAddr().String()
	_, ok := pm.reg[address]
	if ok {
		return errors.New("peer with nonce already known")
	}
	p := &peer{
		conn:   conn,
		input:  make(chan interface{}, pm.buffer),
		output: make(chan interface{}, pm.buffer),
		nonce:  nonce,
	}

	// initialize the readers and writers
	_ = lz4.NewReader(conn)
	_ = lz4.NewWriter(conn)

	// launch the message processing routines
	// TODO: launch handlers to send, receive & process messages

	pm.reg[address] = p
	return nil
}

func (pm *simplePeerManager) DropAll() {
	pm.Lock()
	defer pm.Unlock()
	for _, peer := range pm.reg {
		peer.conn.Close()
	}
}

func (pm *simplePeerManager) Drop(address string) error {
	return nil
}

func (pm *simplePeerManager) Count() uint {
	pm.Lock()
	defer pm.Unlock()
	return uint(len(pm.reg))
}

func (pm *simplePeerManager) Known(nonce []byte) bool {
	hexNonce := hex.EncodeToString(nonce)
	_, ok := pm.reg[hexNonce]
	return ok
}

func (pm *simplePeerManager) Addresses() []string {
	addresses := make([]string, 0, len(pm.reg))
	for _, peer := range pm.reg {
		addresses = append(addresses, peer.conn.RemoteAddr().String())
	}
	return addresses
}
