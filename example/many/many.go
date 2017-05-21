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

package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"go.uber.org/zap"
	"github.com/alvalor/alvalor-go/network"
)

func main() {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT)
	done := make(chan struct{})
	var nodes []*network.Node
	sub := make(chan interface{}, 1000)
	beacon := "127.0.0.1:10000"
	log, _ := zap.NewDevelopment()
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
					log.Warn("message send failed", zap.Error(err))
				}
			case event := <-sub:
				switch e := event.(type) {
				case *network.Connected:
					log.Info("connected to peer", zap.String("addr", e.Address))
				case *network.Disconnected:
					log.Info("disconnected from peer", zap.String("addr", e.Address))
				case *network.Message:
					txt := e.Value.(string)
					log.Info("received message", zap.String("message", txt))
				}
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
