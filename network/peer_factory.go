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
	"time"

	"github.com/pierrec/lz4"
)

type PeerFactory struct {
	codec     Codec
	heartbeat time.Duration
	timeout   time.Duration
}

func (factory *PeerFactory) create(conn net.Conn, nonce []byte) peer {
	addr := conn.RemoteAddr().String()
	r := lz4.NewReader(conn)
	w := lz4.NewWriter(conn)
	outgoing := make(chan interface{}, 16)
	incoming := make(chan interface{}, 16)
	p := peer{
		conn:      conn,
		addr:      addr,
		nonce:     nonce,
		r:         r,
		w:         w,
		outgoing:  outgoing,
		incoming:  incoming,
		codec:     factory.codec,
		heartbeat: factory.heartbeat,
		timeout:   factory.timeout,
		hb:        time.NewTimer(factory.heartbeat),
	}
	return p
}
