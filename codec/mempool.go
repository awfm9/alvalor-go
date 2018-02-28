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

	"github.com/pkg/errors"
	"github.com/willf/bloom"
	capnp "zombiezen.com/go/capnproto2"

	"github.com/alvalor/alvalor-go/node"
)

type initMempool func() (Mempool, error)

func createRootMempool(z Z) initMempool {
	return z.NewMempool
}

func readRootMempool(z Z) initMempool {
	return z.Mempool
}

func encodeMempool(seg *capnp.Segment, create initMempool, e *node.Mempool) (Mempool, error) {
	mempool, err := create()
	if err != nil {
		return Mempool{}, errors.Wrap(err, "could not create mempool")
	}
	buf := &bytes.Buffer{}
	_, err = e.Bloom.WriteTo(buf)
	if err != nil {
		return Mempool{}, errors.Wrap(err, "could not encode bloom")
	}
	err = mempool.SetBloom(buf.Bytes())
	if err != nil {
		return Mempool{}, errors.Wrap(err, "could not set bloom")
	}
	return mempool, nil
}

func decodeMempool(read initMempool) (*node.Mempool, error) {
	mempool, err := read()
	if err != nil {
		return nil, errors.Wrap(err, "could not read mempool")
	}
	data, err := mempool.Bloom()
	if err != nil {
		return nil, errors.Wrap(err, "could not get bloom")
	}
	buf := bytes.NewBuffer(data)
	bloom := &bloom.BloomFilter{}
	_, err = bloom.ReadFrom(buf)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode bloom")
	}
	e := &node.Mempool{
		Bloom: bloom,
	}
	return e, nil
}
