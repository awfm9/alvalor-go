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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alvalor/alvalor-go/node"
	"github.com/alvalor/alvalor-go/types"
)

func TestBatch(t *testing.T) {
	proto := &Proto{}
	batch := &node.Batch{
		Transactions: []*types.Transaction{
			{
				Transfers:  []*types.Transfer{{From: []byte{10}, To: []byte{11}, Amount: 1000}},
				Fees:       []*types.Fee{{From: []byte{13}, Amount: 1300}},
				Data:       []byte{14, 15, 16},
				Signatures: [][]byte{{17, 18, 19}},
			},
			{
				Transfers:  []*types.Transfer{{From: []byte{20}, To: []byte{21}, Amount: 2000}},
				Fees:       []*types.Fee{{From: []byte{23}, Amount: 2300}},
				Data:       []byte{24, 25, 26},
				Signatures: [][]byte{{27, 28, 29}},
			},
			{
				Transfers:  []*types.Transfer{{From: []byte{31}, To: []byte{32}, Amount: 3000}},
				Fees:       []*types.Fee{{From: []byte{33}, Amount: 3300}},
				Data:       []byte{34, 35, 36},
				Signatures: [][]byte{{37, 38, 39}},
			},
		},
	}

	buf := &bytes.Buffer{}
	err := proto.Encode(buf, batch)
	assert.Nil(t, err)

	msg, err := proto.Decode(buf)
	assert.Nil(t, err)
	assert.Equal(t, batch, msg)
}
