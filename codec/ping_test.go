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
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alvalor/alvalor-go/network"
)

func TestPing(t *testing.T) {
	proto := &Proto{}
	ping := &network.Ping{
		Nonce: rand.Uint32(),
	}

	buf := &bytes.Buffer{}
	err := proto.Encode(buf, ping)
	assert.Nil(t, err)

	msg, err := proto.Decode(buf)
	assert.Nil(t, err)
	assert.Equal(t, ping, msg)
}
