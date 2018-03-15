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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewPeerManager(t *testing.T) {
	handlers := &HandlerManagerMock{}
	min := uint(1)
	max := uint(2)
	peers := newSimplePeerManager(handlers, min, max)
	assert.Equal(t, handlers, peers.handlers)
	assert.Equal(t, min, peers.min)
	assert.Equal(t, max, peers.max)
	assert.NotZero(t, peers.buffer)
	assert.NotNil(t, peers.reg)
}

func TestPeerManagerAdd(t *testing.T) {
	nonce := []byte{1, 2, 3, 4, 5}
	address := "192.0.2.100:1337"
	addr := &AddrMock{}
	addr.On("String").Return(address)
	conn := &ConnMock{}
	conn.On("RemoteAddr").Return(addr)
	handlers := &HandlerManagerMock{}
	handlers.On("Sender", mock.Anything, mock.Anything, mock.Anything)
	handlers.On("Processor", mock.Anything, mock.Anything, mock.Anything)
	handlers.On("Receiver", mock.Anything, mock.Anything, mock.Anything)
	peers := &simplePeerManager{
		reg:      make(map[string]*peer),
		handlers: handlers,
	}

	peers.max = 0
	err := peers.Add(conn, nonce)
	assert.NotNil(t, err)
	assert.Empty(t, peers.reg)

	peers.max = 2
	peers.reg[address] = &peer{}
	err = peers.Add(conn, nonce)
	assert.NotNil(t, err)
	assert.Len(t, peers.reg, 1)

	delete(peers.reg, address)
	err = peers.Add(conn, nonce)
	assert.Nil(t, err)
	if assert.Contains(t, peers.reg, address) {
		p := peers.reg[address]
		assert.Equal(t, conn, p.conn)
		assert.Equal(t, nonce, p.nonce)
		handlers.AssertCalled(t, "Sender", address, mock.Anything, mock.Anything)
		handlers.AssertCalled(t, "Processor", address, mock.Anything, mock.Anything)
		handlers.AssertCalled(t, "Receiver", address, mock.Anything, mock.Anything)
	}
}

func TestPeerManagerDrop(t *testing.T) {
	address := "192.0.2.100:1337"
	conn := &ConnMock{}
	p := &peer{conn: conn}
	peers := &simplePeerManager{reg: make(map[string]*peer)}

	err := peers.Drop(address)
	assert.NotNil(t, err)

	conn.On("Close").Return(errors.New("could not close connection")).Once()
	peers.reg[address] = p
	err = peers.Drop(address)
	assert.NotNil(t, err)
	assert.Empty(t, peers.reg)

	conn.On("Close").Return(nil)
	peers.reg[address] = p
	err = peers.Drop(address)
	assert.Nil(t, err)
	assert.Empty(t, peers.reg)
}

func TestPeerManagerCount(t *testing.T) {
	address := "192.0.2.100:1337"
	peers := &simplePeerManager{reg: make(map[string]*peer)}

	assert.Equal(t, uint(0), peers.Count())

	peers.reg[address] = &peer{}
	assert.Equal(t, uint(1), peers.Count())
}

func TestPeerManagerKnown(t *testing.T) {
	address := "192.0.2.100:1337"
	nonce := []byte{1, 2, 3, 4, 5}
	p := &peer{nonce: nonce}
	peers := &simplePeerManager{reg: make(map[string]*peer)}

	ok := peers.Known(nonce)
	assert.False(t, ok)

	peers.reg[address] = p
	ok = peers.Known(nonce)
	assert.True(t, ok)

	ok = peers.Known([]byte{0, 0, 0, 0, 0})
	assert.False(t, ok)
}

func TestPeerManagerAddresses(t *testing.T) {
	peers := &simplePeerManager{reg: make(map[string]*peer)}

	addresses := peers.Addresses()
	assert.Empty(t, addresses)

	address1 := "192.0.2.100:1337"
	address2 := "192.0.2.200:1337"
	peers.reg[address1] = &peer{}
	peers.reg[address2] = &peer{}
	addresses = peers.Addresses()
	assert.ElementsMatch(t, []string{address1, address2}, addresses)
}

func TestPeerManagerSend(t *testing.T) {

	msg := "message"
	address := "192.0.2.100:1337"

	peers := &simplePeerManager{reg: make(map[string]*peer)}

	err := peers.Send(address, msg)
	assert.NotNil(t, err)

	output := make(chan interface{}, 1)
	p := &peer{output: output}
	peers.reg[address] = p
	err = peers.Send(address, msg)
	assert.Nil(t, err)
	select {
	case received := <-output:
		assert.Equal(t, msg, received)
	default:
		t.Error("no message in output channel")
	}

	peers.reg[address].output = nil
	err = peers.Send(address, msg)
	assert.NotNil(t, err)
}
