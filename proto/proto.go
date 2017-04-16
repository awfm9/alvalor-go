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

package proto

import (
	"io"

	"github.com/pkg/errors"
	capnp "zombiezen.com/go/capnproto2"

	"github.com/veltor/veltor-network/message"
)

// Codec struct.
type Codec struct{}

// Encode method.
func (c Codec) Encode(w io.Writer, i interface{}) error {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return errors.Wrap(err, "could not create proto message")
	}
	z, err := NewRootZ(seg)
	if err != nil {
		return errors.Wrap(err, "could not create proto wrapper")
	}
	switch v := i.(type) {
	case *message.Ping:
		var ping Ping
		ping, err = z.NewPing()
		if err != nil {
			return errors.Wrap(err, "could not create proto ping")
		}
		ping.SetNonce(v.Nonce)
	case *message.Pong:
		var pong Pong
		pong, err = z.NewPong()
		if err != nil {
			return errors.Wrap(err, "could not create proto pong")
		}
		pong.SetNonce(v.Nonce)
	case *message.Discover:
		_, err = z.NewDiscover()
		if err != nil {
			return errors.Wrap(err, "could not create proto discover")
		}
	case *message.Peers:
		var peers Peers
		peers, err = z.NewPeers()
		if err != nil {
			return errors.Wrap(err, "could not create proto peers")
		}
		var addrs capnp.TextList
		addrs, err = peers.NewAddresses(int32(len(v.Addresses)))
		if err != nil {
			return errors.Wrap(err, "could not create address list")
		}
		for i, address := range v.Addresses {
			addrs.Set(i, address)
		}
	case string:
		err = z.SetText(v)
		if err != nil {
			return errors.Wrap(err, "could not create proto text")
		}
	case []byte:
		err = z.SetData(v)
		if err != nil {
			return errors.Wrap(err, "could not create proto data")
		}
	default:
		return errors.Errorf("unknown proto type (%T)", i)
	}
	err = capnp.NewEncoder(w).Encode(msg)
	if err != nil {
		return errors.Wrap(err, "could not encode proto message")
	}
	return nil
}

// Decode method.
func (c Codec) Decode(r io.Reader) (interface{}, error) {
	msg, err := capnp.NewDecoder(r).Decode()
	if err != nil {
		return nil, errors.Wrap(err, "could not decode proto message")
	}
	z, err := ReadRootZ(msg)
	if err != nil {
		return nil, errors.Wrap(err, "could not read proto wrapper")
	}
	switch z.Which() {
	case Z_Which_ping:
		ping, err := z.Ping()
		if err != nil {
			return nil, errors.Wrap(err, "could not read proto ping")
		}
		v := message.Ping{
			Nonce: ping.Nonce(),
		}
		return &v, nil
	case Z_Which_pong:
		pong, err := z.Pong()
		if err != nil {
			return nil, errors.Wrap(err, "could not read proto pong")
		}
		v := message.Pong{
			Nonce: pong.Nonce(),
		}
		return &v, nil
	case Z_Which_discover:
		v := message.Discover{}
		return &v, nil
	case Z_Which_peers:
		peers, err := z.Peers()
		if err != nil {
			return nil, errors.Wrap(err, "could not read proto peers")
		}
		v := message.Peers{}
		addrs, err := peers.Addresses()
		if err != nil {
			return nil, errors.Wrap(err, "could not read address list")
		}
		for i := 0; i < addrs.Len(); i++ {
			addr, _ := addrs.At(i)
			v.Addresses = append(v.Addresses, addr)
		}
		return &v, nil
	case Z_Which_text:
		text, err := z.Text()
		if err != nil {
			return nil, errors.Wrap(err, "could not read proto text")
		}
		return text, nil
	case Z_Which_data:
		data, err := z.Data()
		if err != nil {
			return nil, errors.Wrap(err, "could not read proto data")
		}
		return data, nil
	default:
		return nil, errors.Errorf("invalid proto code (%v)", z.Which())
	}
}
