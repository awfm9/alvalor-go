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
	"os"
	"os/signal"
	"sync"

	"github.com/alvalor/alvalor-go/network"
	"go.uber.org/zap"
)

func main() {

	// catch signals
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	// initialize the structured & highly efficient zap logger
	// NOTE: it's super hard to abstract a typed structured logger into an
	// interface, so we just inject the concrete type everywhere
	log, err := zap.NewProduction()
	if err != nil {
		os.Exit(1)
	}

	// initialize the channels we use to plug the network modules together
	events := make(chan interface{})
	addresses := make(chan string, 16)
	subscriber := make(chan interface{})

	// initialize waitgroup for the network modules
	wg := &sync.WaitGroup{}

	// initialize the network modules
	book := &network.SimpleBook{}
	mgr := network.NewManager(log, wg, book, events, addresses, subscriber)
	cli := network.NewClient(log, wg, addresses, events)
	svr := network.NewServer(log, wg, addresses, events)

	// initialize drivers
	bal := network.NewBalancer(events)

	// launch the drivers
	go bal.Start()

	// TODO: stopping logic
	<-c

	// stop the drivers
	bal.Close()

	// stop the network components
	mgr.Close()
	cli.Close()
	svr.Close()

	// wait for all modules to shut down
	wg.Wait()
	close(events)
}
