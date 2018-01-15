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
	"net"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type listenerActions interface {
	StartAcceptor(conn net.Conn)
}

type listenFunc func(addr *net.TCPAddr) (Listener, error)

func handleListening(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, actions listenerActions, listener listenFunc, stop <-chan struct{}) {
	defer wg.Done()

	// extract the config parameters we are interested in
	var (
		address = cfg.address
	)

	// configure the component logger and set start/stop messages
	log = log.With().Str("component", "listener").Str("address", address).Logger()
	log.Info().Msg("listening routine started")
	defer log.Info().Msg("listening routine stopped")

	// resolve and start listening on the address
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Error().Err(err).Msg("could not resolve listen address")
		return
	}

	ln, err := listener(addr)
	if err != nil {
		log.Error().Err(err).Msg("could not listen on address")
		return
	}

Loop:
	for {

		// keep checking if we should quit
		select {
		case <-stop:
			break Loop
		default:
		}

		// if not try to accept a new connection with a low enough timeout so
		// quiting doesn't block too long due to long for loop iterations
		ln.SetDeadline(time.Now().Add(time.Millisecond * 100))
		var conn net.Conn
		conn, err = ln.Accept()
		if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
			// this is the default timeout we get with the deadline, so just iterate
			continue
		}
		if err != nil {
			log.Error().Err(err).Msg("could not accept connection")
			break
		}

		// we should handle onboarding on a new goroutine to avoid blocking
		// on listening, and as well so we can release slots with defer
		actions.StartAcceptor(conn)
	}

	// ordered to quit, we close the listener down
	err = ln.Close()
	if err != nil {
		log.Error().Err(err).Msg("could not close listener")
		return
	}
}

type Listener interface {
	Accept() (net.Conn, error)
	Close() error
	SetDeadline(time.Time) error
}
