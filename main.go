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
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/alvalor/alvalor-go/network"
	"github.com/rs/zerolog"
)

func main() {

	rand.Seed(time.Now().UnixNano())

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	log := zerolog.New(os.Stderr)

	port := flag.Int("port", 31337, "server port number")
	flag.Parse()

	wg := &sync.WaitGroup{}
	mgr := network.NewManager(log,
		network.SetListen(true),
		network.SetAddress(fmt.Sprintf("127.0.0.1:%v", *port)),
	)

	<-c

	mgr.Stop()

	wg.Wait()
}
