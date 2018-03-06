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

	"github.com/alvalor/alvalor-go/types"
)

func TestTransaction(t *testing.T) {
	proto := &Proto{}
	tx := &types.Transaction{
		Transfers: []*types.Transfer{
			{From: []byte{1}, To: []byte{2}, Amount: 1000},
			{From: []byte{2}, To: []byte{3}, Amount: 2000},
			{From: []byte{4}, To: []byte{5}, Amount: 3000},
		},
		Fees: []*types.Fee{
			{From: []byte{15}, Amount: 19},
			{From: []byte{25}, Amount: 29},
			{From: []byte{35}, Amount: 39},
		},
		Data:  []byte{10, 20, 30},
		Nonce: 1337,
		Signatures: [][]byte{
			{1, 2, 3},
			{4, 5, 6},
			{7, 8, 9},
		},
	}

	buf := &bytes.Buffer{}
	err := proto.Encode(buf, tx)
	assert.Nil(t, err)

	msg, err := proto.Decode(buf)
	assert.Nil(t, err)
	assert.Equal(t, tx, msg)
}
