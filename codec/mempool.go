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

	"github.com/alvalor/alvalor-go/node"
	"github.com/pkg/errors"
	capnp "zombiezen.com/go/capnproto2"
)

func mempoolToMessage(entity *node.Mempool) (*capnp.Message, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize message")
	}
	z, err := NewRootZ(seg)
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize wrapper")
	}
	buf := &bytes.Buffer{}
	_, err = entity.Bloom.WriteTo(buf)
	if err != nil {
		return nil, errors.Wrap(err, "could not encode bloom filter")
	}
	mempool, err := z.NewMempool()
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize mempool")
	}
	err = mempool.SetBloom(buf.Bytes())
	if err != nil {
		return nil, errors.Wrap(err, "could not set bloom")
	}
	return msg, nil
}
