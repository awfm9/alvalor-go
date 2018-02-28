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
	"io"

	"github.com/pkg/errors"
	capnp "zombiezen.com/go/capnproto2"

	"github.com/alvalor/alvalor-go/network"
	"github.com/alvalor/alvalor-go/node"
	"github.com/alvalor/alvalor-go/types"
)

// Proto represents the capnproto serialization module.
type Proto struct{}

// NewProto will return a new proto codec.
func NewProto() Proto {
	return Proto{}
}

// Encode will serialize the provided entity by writing the binary format into the provided writer.
// It will fail if the entity type is unknown.
func (p Proto) Encode(w io.Writer, entity interface{}) error {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return errors.Wrap(err, "could not create message")
	}
	z, err := NewRootZ(seg)
	if err != nil {
		return errors.Wrap(err, "could not create wrapper")
	}
	switch e := entity.(type) {
	case *network.Ping:
		_, err = encodePing(seg, createRootPing(z), e)
	case *network.Pong:
		_, err = encodePong(seg, createRootPong(z), e)
	case *network.Discover:
		_, err = encodeDiscover(seg, createRootDiscover(z), e)
	case *network.Peers:
		_, err = encodePeers(seg, createRootPeers(z), e)
	case *types.Transaction:
		_, err = encodeTransaction(seg, createRootTransaction(z), e)
	case *node.Mempool:
		_, err = encodeMempool(seg, createRootMempool(z), e)
	case *node.Inventory:
		_, err = encodeInventory(seg, createRootInventory(z), e)
	case *node.Request:
		_, err = encodeRequest(seg, createRootRequest(z), e)
	case *node.Batch:
		_, err = encodeBatch(seg, createRootBatch(z), e)
	default:
		return errors.Errorf("unknown message type (%T)", e)
	}
	if err != nil {
		return err
	}
	err = capnp.NewEncoder(w).Encode(msg)
	if err != nil {
		return errors.Wrap(err, "could not encode message")
	}
	return nil
}

// Decode will decode the binary data of the given reader into the original entity.
func (p Proto) Decode(r io.Reader) (interface{}, error) {
	msg, err := capnp.NewDecoder(r).Decode()
	if err != nil {
		return nil, errors.Wrap(err, "could not decode message")
	}
	z, err := ReadRootZ(msg)
	if err != nil {
		return nil, errors.Wrap(err, "could not read wrapper")
	}
	switch z.Which() {
	case Z_Which_ping:
		return decodePing(readRootPing(z))
	case Z_Which_pong:
		return decodePong(readRootPong(z))
	case Z_Which_discover:
		return decodeDiscover(readRootDiscover(z))
	case Z_Which_peers:
		return decodePeers(readRootPeers(z))
	case Z_Which_transaction:
		return decodeTransaction(readRootTransaction(z))
	case Z_Which_mempool:
		return decodeMempool(readRootMempool(z))
	case Z_Which_inventory:
		return decodeInventory(readRootInventory(z))
	case Z_Which_request:
		return decodeRequest(readRootRequest(z))
	case Z_Which_batch:
		return decodeBatch(readRootBatch(z))
	default:
		return nil, errors.Errorf("unknown message code (%v)", z.Which())
	}
}
