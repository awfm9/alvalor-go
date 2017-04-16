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

package network

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type peer struct {
	mutex     sync.Mutex
	conn      net.Conn
	addr      string
	nonce     []byte
	r         io.Reader
	w         io.Writer
	out       chan *Packet
	err       error
	codec     Codec
	timeout   time.Duration
	heartbeat time.Duration
	hb        *time.Timer
}

func (p *peer) receive() {
	for {
		p.conn.SetReadDeadline(time.Now().Add(p.timeout))
		i, err := p.codec.Decode(p.r)
		if err != nil {
			p.err = errors.Wrap(err, "could not decode message")
			close(p.out)
			break
		}
		p.hb.Stop()
		p.hb.Reset(p.heartbeat)
		pk := Packet{
			Address: p.addr,
			Message: i,
		}
		p.out <- &pk
	}
}

func (p *peer) send(i interface{}) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	err := p.codec.Encode(p.w, i)
	if err != nil {
		return errors.Wrap(err, "could not encode message")
	}
	return nil
}

func (p *peer) close() error {
	return p.conn.Close()
}
