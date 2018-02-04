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
	"io"
	"net"
	"sync"

	"github.com/rs/zerolog"
)

type handlerManager interface {
	Listen()
	Accept(conn net.Conn)
	Connect(address string)
	Send(address string, output <-chan interface{}, w io.Writer)
	Process(address string, input <-chan interface{}, output chan<- interface{})
	Receive(address string, r io.Reader, input chan<- interface{})
}

type simpleHandlerManager struct {
	log       zerolog.Logger
	wg        *sync.WaitGroup
	cfg       *Config
	dialer    dialWrapper
	listener  listenWrapper
	addresses addressManager
	pending   pendingManager
	peers     peerManager
	rep       reputationManager
	stop      chan struct{}
}

func (hm *simpleHandlerManager) Drop() {
	go handleDropping(hm.log, hm.wg, hm.cfg, hm.peers, hm.stop)
}

func (hm *simpleHandlerManager) Serve() {
	go handleServing(hm.log, hm.wg, hm.cfg, hm.peers, hm, hm.stop)
}

func (hm *simpleHandlerManager) Dial() {
	go handleDialing(hm.log, hm.wg, hm.cfg, hm.peers, hm.pending, hm.addresses, hm.rep, hm, hm.stop)
}

func (hm *simpleHandlerManager) Listen() {
	go handleListening(hm.log, hm.wg, hm.cfg, hm, hm.listener, hm.stop)
}

func (hm *simpleHandlerManager) Accept(conn net.Conn) {
	go handleAccepting(hm.log, hm.wg, hm.cfg, hm.pending, hm.peers, hm.rep, conn)
}

func (hm *simpleHandlerManager) Connect(address string) {
	go handleConnecting(hm.log, hm.wg, hm.cfg, hm.pending, hm.peers, hm.rep, hm.dialer, address)
}

func (hm *simpleHandlerManager) Send(address string, output <-chan interface{}, w io.Writer) {
	go handleSending(hm.log, hm.wg, hm.cfg, hm.peers, hm.rep, address, output, w)
}

func (hm *simpleHandlerManager) Process(address string, input <-chan interface{}, output chan<- interface{}) {
	go handleProcessing(hm.log, hm.wg, hm.cfg, hm.addresses, hm.peers, address, input, output)
}

func (hm *simpleHandlerManager) Receive(address string, r io.Reader, input chan<- interface{}) {
	go handleReceiving(hm.log, hm.wg, hm.cfg, hm.peers, hm.rep, address, r, input)
}

func (hm *simpleHandlerManager) Stop() {
	close(hm.stop)
	addresses := hm.peers.Addresses()
	for _, address := range addresses {
		hm.peers.Drop(address)
	}
	hm.wg.Wait()
}
