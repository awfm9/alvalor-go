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
	"bytes"
	"io"

	"github.com/pkg/errors"
	"github.com/willf/bloom"
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
		return errors.Wrap(err, "could not initialize message")
	}
	z, err := NewRootZ(seg)
	if err != nil {
		return errors.Wrap(err, "could not initialize wrapper")
	}
	switch e := entity.(type) {
	case *network.Ping:
		_, err = encodePing(seg, rootPing(z), e)
	case *network.Pong:
		_, err = encodePong(seg, rootPong(z), e)
	case *network.Discover:
		_, err = encodeDiscover(seg, rootDiscover(z), e)
	case *network.Peers:
		_, err = encodePeers(seg, rootPeers(z), e)
	case *types.Transaction:
		_, err = encodeTransaction(seg, rootTransaction(z), e)
	case *node.Mempool:
		_, err = encodeMempool(seg, rootMempool(z), e)
	case *node.Inventory:
		_, err = encodeInventory(seg, rootInventory(z), e)
	case *node.Request:
		_, err = encodeRequest(seg, rootRequest(z), e)
	default:
		return errors.Errorf("unknown message type (%T)", e)
	}
	if err != nil {
		return errors.Wrap(err, "could not encode entity")
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
		v := network.Ping{
			Nonce: ping.Nonce(),
		}
		return &v, nil
	case Z_Which_pong:
		pong, err := z.Pong()
		if err != nil {
			return nil, errors.Wrap(err, "could not read proto pong")
		}
		v := network.Pong{
			Nonce: pong.Nonce(),
		}
		return &v, nil
	case Z_Which_discover:
		v := network.Discover{}
		return &v, nil
	case Z_Which_peers:
		peers, err := z.Peers()
		if err != nil {
			return nil, errors.Wrap(err, "could not read proto peers")
		}
		addrs, err := peers.Addresses()
		if err != nil {
			return nil, errors.Wrap(err, "could not read address list")
		}
		v := network.Peers{}
		for i := 0; i < addrs.Len(); i++ {
			addr, _ := addrs.At(i)
			v.Addresses = append(v.Addresses, addr)
		}
		return &v, nil
	case Z_Which_transaction:
		transaction, err := z.Transaction()
		if err != nil {
			return nil, errors.Wrap(err, "could not read proto transaction")
		}
		transfers, err := transaction.Transfers()
		if err != nil {
			return nil, errors.Wrap(err, "could not read transfer list")
		}
		fees, err := transaction.Fees()
		if err != nil {
			return nil, errors.Wrap(err, "could not read fee list")
		}
		data, err := transaction.Data()
		if err != nil {
			return nil, errors.Wrap(err, "could not read data")
		}
		sigs, err := transaction.Signatures()
		if err != nil {
			return nil, errors.Wrap(err, "could not read sigs")
		}
		v := types.Transaction{
			Transfers:  make([]types.Transfer, 0, transfers.Len()),
			Fees:       make([]types.Fee, 0, fees.Len()),
			Data:       data,
			Signatures: make([][]byte, 0, sigs.Len()),
		}
		for i := 0; i < transfers.Len(); i++ {
			transfer := transfers.At(i)
			from, err := transfer.From()
			if err != nil {
				return nil, errors.Wrap(err, "could not read transfer from")
			}
			to, err := transfer.To()
			if err != nil {
				return nil, errors.Wrap(err, "could not read transfer to")
			}
			amount := transfer.Amount()
			item := types.Transfer{
				From:   from,
				To:     to,
				Amount: amount,
			}
			v.Transfers = append(v.Transfers, item)
		}
		for i := 0; i < fees.Len(); i++ {
			fee := fees.At(i)
			from, err := fee.From()
			if err != nil {
				return nil, errors.Wrap(err, "could not read fee from")
			}
			amount := fee.Amount()
			item := types.Fee{
				From:   from,
				Amount: amount,
			}
			v.Fees = append(v.Fees, item)
		}
		for i := 0; i < sigs.Len(); i++ {
			sig, err := sigs.At(i)
			if err != nil {
				return nil, errors.Wrap(err, "could not read signature")
			}
			v.Signatures = append(v.Signatures, sig)
		}
		return &v, nil
	case Z_Which_mempool:
		mempool, err := z.Mempool()
		if err != nil {
			return nil, errors.Wrap(err, "could not read proto mempool")
		}
		data, err := mempool.Bloom()
		if err != nil {
			return nil, errors.Wrap(err, "could not read mempool bloom")
		}
		buf := bytes.NewBuffer(data)
		bloom := &bloom.BloomFilter{}
		_, err = bloom.ReadFrom(buf)
		if err != nil {
			return nil, errors.Wrap(err, "could not decode bloom filter")
		}
		v := node.Mempool{
			Bloom: bloom,
		}
		return &v, nil
	case Z_Which_inventory:
		inventory, err := z.Inventory()
		if err != nil {
			return nil, errors.Wrap(err, "could not read proto inventory")
		}
		ids, err := inventory.Ids()
		if err != nil {
			return nil, errors.Wrap(err, "could not read inventory IDs")
		}
		v := node.Inventory{
			IDs: make([][]byte, 0, ids.Len()),
		}
		for i := 0; i < ids.Len(); i++ {
			id, err := ids.At(i)
			if err != nil {
				return nil, errors.Wrap(err, "could not read ID")
			}
			v.IDs = append(v.IDs, id)
		}
		return &v, nil
	case Z_Which_request:
		request, err := z.Request()
		if err != nil {
			return nil, errors.Wrap(err, "could not read proto request")
		}
		ids, err := request.Ids()
		if err != nil {
			return nil, errors.Wrap(err, "could not read request IDs")
		}
		v := node.Request{
			IDs: make([][]byte, 0, ids.Len()),
		}
		for i := 0; i < ids.Len(); i++ {
			id, err := ids.At(i)
			if err != nil {
				return nil, errors.Wrap(err, "could not read ID")
			}
			v.IDs = append(v.IDs, id)
		}
		return &v, nil
	default:
		return nil, errors.Errorf("invalid proto code (%v)", z.Which())
	}
}
