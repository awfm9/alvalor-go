// Copyright (c) 2017 The Veltor Authors
//
// This file is part of Veltor.
//
// Veltor is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Veltor is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Veltor.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"go.uber.org/zap"
	"github.com/veltor/veltor-go/network"
	"github.com/veltor/veltor-go/protocol"
)

func main() {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT)
	done := make(chan struct{})
	addr := "127.0.0.1:10000"
	sub := make(chan interface{}, 1000)
	codec := protocol.Codec{}
	log, _ := zap.NewDevelopment()
	node1 := network.NewNode(
		network.SetLog(log),
		network.SetCodec(codec),
		network.SetSubscriber(sub),
		network.SetServer(true),
		network.SetAddress(addr),
		network.SetMinPeers(1),
		network.SetMaxPeers(1),
	)
	book := network.NewSimpleBook()
	book.Add(addr)
	node2 := network.NewNode(
		network.SetLog(log),
		network.SetCodec(codec),
		network.SetBook(book),
		network.SetSubscriber(sub),
		network.SetMinPeers(1),
		network.SetMaxPeers(1),
	)
	go func() {
	Loop:
		for {
			select {
			case <-done:
				break Loop
			case <-time.After(2 * time.Second):
				msg := strconv.FormatUint(uint64(rand.Uint32()), 10)
				err := node2.Send(addr, msg)
				if err != nil {
					log.Warn("message send failed", zap.Error(err))
				}
			case event := <-sub:
				switch e := event.(type) {
				case *network.Connected:
					log.Info("connected to peer on address", zap.String("addr", e.Address))
				case *network.Disconnected:
					log.Info("disconnected from peer on address", zap.String("addr", e.Address))
				case *network.Message:
					txt := e.Value.(string)
					log.Info("received message", zap.String("message", txt))
				}
			}
		}
	}()
	<-sig
	close(done)
	_ = node1
	_ = node2
	os.Exit(0)
}
