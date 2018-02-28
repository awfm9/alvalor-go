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
	"github.com/pkg/errors"
	capnp "zombiezen.com/go/capnproto2"

	"github.com/alvalor/alvalor-go/node"
)

type initRequest func() (Request, error)

func createRootRequest(z Z) initRequest {
	return z.NewRequest
}

func readRootRequest(z Z) initRequest {
	return z.Request
}

func encodeRequest(seg *capnp.Segment, create initRequest, e *node.Request) (Request, error) {
	request, err := create()
	if err != nil {
		return Request{}, errors.Wrap(err, "could not create request")
	}
	ids, err := request.NewIds(int32(len(e.IDs)))
	if err != nil {
		return Request{}, errors.Wrap(err, "could not create ID list")
	}
	for i, id := range e.IDs {
		err = ids.Set(i, id)
		if err != nil {
			return Request{}, errors.Wrap(err, "could not set ID")
		}
	}
	return request, nil
}

func decodeRequest(read initRequest) (*node.Request, error) {
	request, err := read()
	if err != nil {
		return nil, errors.Wrap(err, "could not read request")
	}
	ids, err := request.Ids()
	if err != nil {
		return nil, errors.Wrap(err, "could not read ID list")
	}
	e := &node.Request{
		IDs: make([][]byte, 0, ids.Len()),
	}
	for i := 0; i < ids.Len(); i++ {
		id, err := ids.At(i)
		if err != nil {
			return nil, errors.Wrap(err, "could not get ID")
		}
		e.IDs = append(e.IDs, id)
	}
	return e, nil
}
