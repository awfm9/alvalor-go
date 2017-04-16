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
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	logger "github.com/veltor/veltor-logger"
	network "github.com/veltor/veltor-network"
)

func main() {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT)
	done := make(chan struct{})
	var nodes []*network.Node
	sub := make(chan interface{})
	beacon := "127.0.0.1:10000"
	log := logger.New()
	log.SetLevel(logger.Debug)
	for i := 0; i < 16; i++ {
		book := network.NewSimpleBook()
		book.Add(beacon)
		addr := fmt.Sprintf("127.0.0.1:%v", 10000+i)
		node := network.NewNode(
			network.SetLog(log),
			network.SetBook(book),
			network.SetSubscriber(sub),
			network.SetServer(true),
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
			case <-time.After(time.Second * 1):
				msg := strconv.FormatUint(uint64(rand.Uint32()), 10)
				node := nodes[rand.Int()%len(nodes)]
				err := node.Send(beacon, msg)
				if err != nil {
					log.Warningf("message send failed: %v", err)
				}
			case packet := <-sub:
				msg := packet.(*network.Packet).Message.(string)
				log.Infof("received message: %v", msg)
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
