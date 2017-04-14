// Copyright (c) 2017 The Veltor Authors
//
// This file is part of Veltor.
//
// Veltor Network is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Veltor Network is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Veltor Network.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	network "github.com/veltor/veltor-network"
	"github.com/veltor/veltor-network/proto"
)

func main() {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT)
	done := make(chan struct{})
	addr := "127.0.0.1:10000"
	sub := make(chan interface{})
	codec := proto.Codec{}
	log1 := network.NewSimpleLog("A")
	node1 := network.NewNode(
		network.SetLog(log1),
		network.SetCodec(codec),
		network.SetSubscriber(sub),
		network.SetListen(true),
		network.SetAddress(addr),
		network.SetMinPeers(1),
		network.SetMaxPeers(1),
	)
	book := network.NewSimpleBook()
	book.Add(addr)
	log2 := network.NewSimpleLog("B")
	node2 := network.NewNode(
		network.SetLog(log2),
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
					log.Printf("message send failed: %v", err)
				}
			case packet := <-sub:
				msg := packet.(*network.Packet).Message.(string)
				log.Printf("received message: %v", msg)
			}
		}
	}()
	<-sig
	close(done)
	_ = node1
	_ = node2
	os.Exit(0)
}
