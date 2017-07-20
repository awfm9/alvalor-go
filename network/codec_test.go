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

package network

import (
    "testing"
    "bytes"
    "math/rand"
    "github.com/stretchr/testify/assert"
)

func TestEncodeUnknownType(t *testing.T) {
    codec := SimpleCodec{}

    data := &bytes.Buffer{}   
    msg := 1
    err := codec.Encode(data, msg)

    assert.NotNil(t, err, "Expected to receive error in case message type is unknown")
}

func TestDecodeUnknownCode(t *testing.T) {
    codec := SimpleCodec{}

    data := []byte{77}
    reader := bytes.NewReader(data)
    _, err := codec.Decode(reader)

    assert.NotNil(t, err, "Expected to receive error in case message type is unknown")
}

func TestEncodePing(t *testing.T) {
    codec := SimpleCodec{}

    data := &bytes.Buffer{}   
    msg := &Ping{ Nonce : rand.Uint32()}
    codec.Encode(data, msg)

    decoded, _ := codec.Decode(data)

    decodedMsg, ok := decoded.(*Ping)

    assert.True(t, ok, "Decoded type is other than expected")
    assert.Equal(t, msg.Nonce, decodedMsg.Nonce)
}

func TestEncodePong(t *testing.T) {
    codec := SimpleCodec{}

    data := &bytes.Buffer{}   
    msg := &Pong{ Nonce : rand.Uint32()}
    codec.Encode(data, msg)

    decoded, _ := codec.Decode(data)

    decodedMsg, ok := decoded.(*Pong)

    assert.True(t, ok, "Decoded type is other than expected")
    assert.Equal(t, msg.Nonce, decodedMsg.Nonce)
}

func TestEncodeDiscover(t *testing.T) {
    codec := SimpleCodec{}

    data := &bytes.Buffer{}   
    msg := &Discover{}
    codec.Encode(data, msg)

    decoded, _ := codec.Decode(data)

    _, ok := decoded.(*Discover)

    assert.True(t, ok, "Decoded type is other than expected")
}

func TestEncodePeers(t *testing.T) {
    codec := SimpleCodec{}

    data := bytes.NewBuffer(make([]byte, 0))
    addrs := []string{"127.0.0.1", "192.168.4.62"}
    msg := &Peers{ Addresses: addrs}
    codec.Encode(data, msg)

    decoded, _ := codec.Decode(data)

    decodedMsg, ok := decoded.(*Peers)

    assert.True(t, ok, "Decoded type is other than expected")
    assert.EqualValues(t, msg.Addresses, decodedMsg.Addresses)
}

func TestEncodeString(t *testing.T) {
    codec := SimpleCodec{}

    data := &bytes.Buffer{}   
    msg := "hello"
    codec.Encode(data, msg)

    decoded, _ := codec.Decode(data)

    decodedMsg, ok := decoded.(string)

    assert.True(t, ok, "Decoded type is other than expected")
    assert.Equal(t, msg, decodedMsg)
}

func TestEncodeBytes(t *testing.T) {
    codec := SimpleCodec{}

    data := &bytes.Buffer{}   
    msg := []byte{1,55}
    codec.Encode(data, msg)

    decoded, _ := codec.Decode(data)

    decodedMsg, ok := decoded.([]byte)

    assert.True(t, ok, "Decoded type is other than expected")
    assert.EqualValues(t, msg, decodedMsg)
}
