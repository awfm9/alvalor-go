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
	"github.com/alvalor/alvalor-go/node"
	"github.com/pkg/errors"
	capnp "zombiezen.com/go/capnproto2"
)

type initRequest func() (Request, error)

func rootRequest(z Z) initRequest {
	return z.NewRequest
}

func encodeRequest(seg *capnp.Segment, init initRequest, e *node.Request) (Request, error) {
	request, err := init()
	if err != nil {
		return Request{}, errors.Wrap(err, "could not initialize request")
	}
	ids, err := request.NewIds(int32(len(e.IDs)))
	if err != nil {
		return Request{}, errors.Wrap(err, "could not initialize ID list")
	}
	for i, id := range e.IDs {
		err = ids.Set(i, id)
		if err != nil {
			return Request{}, errors.Wrap(err, "could not set ID")
		}
	}
	return request, nil
}
