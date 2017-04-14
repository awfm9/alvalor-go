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
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	network "github.com/veltor/veltor-network"
)

func main() {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT)
	done := make(chan struct{})
	var nodes []*network.Node
	sub := make(chan interface{})
	beacon := "127.0.0.1:10000"
	for i := 0; i < 16; i++ {
		book := network.NewSimpleBook()
		book.Add(beacon)
		addr := fmt.Sprintf("127.0.0.1:%v", 10000+i)
		log1 := network.NewSimpleLog(fmt.Sprintf("node-%v", i))
		node := network.NewNode(
			network.SetLog(log1),
			network.SetBook(book),
			network.SetSubscriber(sub),
			network.SetListen(true),
			network.SetAddress(addr),
			network.SetMinPeers(4),
			network.SetMaxPeers(15),
		)
		nodes = append(nodes, node)
	}
	go func() {
	Loop:
		for {
			select {
			case <-done:
				break Loop
			case <-time.After(time.Second * 5):
				msg := strconv.FormatUint(uint64(rand.Uint32()), 10)
				node := nodes[rand.Int()%len(nodes)]
				err := node.Send(beacon, msg)
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
	for _, node := range nodes {
		_ = node
	}
	os.Exit(0)
}
