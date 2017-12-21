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

package codec

import (
	"encoding/json"
	"io"

	"github.com/pkg/errors"
	"github.com/veltor/veltor-go/network"
)

// Enumeration of different entity types that we use to select the entity for decoding.
const (
	MsgPing = iota
	MsgPong
	MsgDiscover
	MsgPeers
	MsgString
	MsgBytes
)

// Simple is a network codec using a simple JSON format for encoding traffic
// over the network.
type Simple struct{}

// NewSimple will create a new simple codec.
func NewSimple() Simple {
	return Simple{}
}

// Encode will write the input message to the given writer using the simple JSON
// format.
func (s Simple) Encode(w io.Writer, i interface{}) error {
	code := make([]byte, 1)
	switch i.(type) {
	case *network.Ping:
		code[0] = MsgPing
	case *network.Pong:
		code[0] = MsgPong
	case *network.Discover:
		code[0] = MsgDiscover
	case *network.Peers:
		code[0] = MsgPeers
	case string:
		code[0] = MsgString
	case []byte:
		code[0] = MsgBytes
	default:
		return errors.Errorf("invalid message type (%T)", i)
	}
	_, err := w.Write(code)
	if err != nil {
		return errors.Wrap(err, "could not write message code")
	}
	enc := json.NewEncoder(w)
	err = enc.Encode(i)
	if err != nil {
		return errors.Wrap(err, "could not encode message data")
	}
	return nil
}

// Decode will read a message in simple JSON format from the given reader.
func (s Simple) Decode(r io.Reader) (interface{}, error) {
	code := make([]byte, 1)
	_, err := r.Read(code)
	if err != nil {
		return nil, errors.Wrap(err, "could not read message code")
	}
	dec := json.NewDecoder(r)
	var i interface{}
	switch code[0] {
	case MsgPing:
		var ping network.Ping
		err = dec.Decode(&ping)
		i = &ping
	case MsgPong:
		var pong network.Pong
		err = dec.Decode(&pong)
		i = &pong
	case MsgDiscover:
		var discover network.Discover
		err = dec.Decode(&discover)
		i = &discover
	case MsgPeers:
		var peers network.Peers
		err = dec.Decode(&peers)
		i = &peers
	case MsgString:
		var str string
		err = dec.Decode(&str)
		i = str
	case MsgBytes:
		var bytes []byte
		err = dec.Decode(&bytes)
		i = bytes
	default:
		return nil, errors.Errorf("invalid message code (%T)", code)
	}
	if err != nil {
		return nil, errors.Wrap(err, "could not decode message data")
	}
	return i, nil
}
