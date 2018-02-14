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
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/pflag"

	"github.com/alvalor/alvalor-go/codec"
	"github.com/alvalor/alvalor-go/network"
)

func main() {

	// set up a channel to catch system signals to cleanly shut down
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	// initialize default configuration
	cfg := DefaultConfig

	// apply the command line parameters to configuration
	pflag.Uint16Var(&cfg.Port, "port", 21517, "listen port for incoming connections")
	pflag.Parse()

	// seed the random generator
	rand.Seed(time.Now().UnixNano())

	// initialize the logger to standard error as output
	log := zerolog.New(os.Stderr)

	// use our efficient capnproto codec for network communication
	cod := codec.NewProto()

	// initialize the network component to create our p2p network node
	net := network.New(log, cod,
		network.SetListen(cfg.Listen),
		network.SetAddress(fmt.Sprintf("%v:%v", cfg.IP, cfg.Port)),
		network.SetMinPeers(1),
	)

	// add the bootstrapping nodes
	for _, address := range cfg.Bootstrap {
		net.Add(address)
	}

	// wait for a stop signal to initialize shutdown
	<-sig
	signal.Stop(sig)
	close(sig)

	// shut down the p2p network node
	net.Stop()
}
