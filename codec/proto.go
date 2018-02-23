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
func (p Proto) Encode(w io.Writer, i interface{}) error {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return errors.Wrap(err, "could not create proto message")
	}
	z, err := NewRootZ(seg)
	if err != nil {
		return errors.Wrap(err, "could not create proto wrapper")
	}
	switch v := i.(type) {
	case *network.Ping:
		var ping Ping
		ping, err = z.NewPing()
		if err != nil {
			return errors.Wrap(err, "could not create proto ping")
		}
		ping.SetNonce(v.Nonce)
	case *network.Pong:
		var pong Pong
		pong, err = z.NewPong()
		if err != nil {
			return errors.Wrap(err, "could not create proto pong")
		}
		pong.SetNonce(v.Nonce)
	case *network.Discover:
		_, err = z.NewDiscover()
		if err != nil {
			return errors.Wrap(err, "could not create proto discover")
		}
	case *network.Peers:
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
			err = addrs.Set(i, address)
			if err != nil {
				return errors.Wrap(err, "could not set address")
			}
		}
	case *types.Transaction:
		var transaction Transaction
		transaction, err = z.NewTransaction()
		if err != nil {
			return errors.Wrap(err, "could not create proto transaction")
		}
		var transfers Transfer_List
		transfers, err = transaction.NewTransfers(int32(len(v.Transfers)))
		if err != nil {
			return errors.Wrap(err, "could not create transaction list")
		}
		for i, item := range v.Transfers {
			var transfer Transfer
			transfer, err = NewTransfer(seg)
			if err != nil {
				return errors.Wrap(err, "could not create proto transfer")
			}
			transfer.SetFrom(item.From)
			transfer.SetTo(item.To)
			transfer.SetAmount(item.Amount)
			err = transfers.Set(i, transfer)
			if err != nil {
				return errors.Wrap(err, "could not set transfer")
			}
		}
		var fees Fee_List
		fees, err = transaction.NewFees(int32(len(v.Fees)))
		if err != nil {
			return errors.Wrap(err, "could not create fee list")
		}
		for i, item := range v.Fees {
			var fee Fee
			fee, err = NewFee(seg)
			if err != nil {
				return errors.Wrap(err, "could not create proto fee")
			}
			fee.SetFrom(item.From)
			fee.SetAmount(item.Amount)
			err = fees.Set(i, fee)
			if err != nil {
				return errors.Wrap(err, "could not set fee")
			}
		}
		err = transaction.SetData(v.Data)
		if err != nil {
			return errors.Wrap(err, "could not set transaction data")
		}
		var sigs capnp.DataList
		sigs, err = transaction.NewSignatures(int32(len(v.Signatures)))
		if err != nil {
			return errors.Wrap(err, "could not create signature list")
		}
		for i, sig := range v.Signatures {
			err = sigs.Set(i, sig)
			if err != nil {
				return errors.Wrap(err, "could not set signature")
			}
		}
	case *node.Mempool:
		buf := &bytes.Buffer{}
		_, err = v.Bloom.WriteTo(buf)
		if err != nil {
			return errors.Wrap(err, "could not encode bloom filter")
		}
		var mempool Mempool
		mempool, err = z.NewMempool()
		if err != nil {
			return errors.Wrap(err, "could not create proto mempool")
		}
		err = mempool.SetBloom(buf.Bytes())
		if err != nil {
			return errors.Wrap(err, "could not set mempool bloom")
		}
	case *node.Inventory:
		var inventory Inventory
		inventory, err = z.NewInventory()
		if err != nil {
			return errors.Wrap(err, "could not create proto inventory")
		}
		var ids capnp.DataList
		ids, err = inventory.NewIds(int32(len(v.IDs)))
		if err != nil {
			return errors.Wrap(err, "could not create id list")
		}
		for i, id := range v.IDs {
			err = ids.Set(i, id)
			if err != nil {
				return errors.Wrap(err, "could not set ID")
			}
		}
	case *node.Request:
		var request Request
		request, err = z.NewRequest()
		if err != nil {
			return errors.Wrap(err, "could not create proto request")
		}
		var ids capnp.DataList
		ids, err = request.NewIds(int32(len(v.IDs)))
		if err != nil {
			return errors.Wrap(err, "could not create id list")
		}
		for i, id := range v.IDs {
			err = ids.Set(i, id)
			if err != nil {
				return errors.Wrap(err, "could not set ID")
			}
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
		v := &node.Mempool{
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
		v := &node.Inventory{
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
		v := &node.Request{
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
