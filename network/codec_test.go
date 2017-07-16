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

    buf := &bytes.Buffer{}   
    msg := 1
    err := codec.Encode(buf, msg)

    assert.NotNil(t, err, "Expected to receive error in case message type is unknown")
}

func TestDecodeUnknownCode(t *testing.T) {
    codec := SimpleCodec{}

    buf := make([]byte, 1)
    buf[0] = 77
    reader := bytes.NewReader(buf)
    _, err := codec.Decode(reader)

    assert.NotNil(t, err, "Expected to receive error in case message type is unknown")
}

func TestEncodePing(t *testing.T) {
    codec := SimpleCodec{}

    buf := &bytes.Buffer{}   
    msg := &Ping{ Nonce : rand.Uint32()}
    codec.Encode(buf, msg)

    decoded, _ := codec.Decode(buf)

    decodedMsg, ok := decoded.(*Ping)

    assert.True(t, ok, "Decoded type is other than expected")
    assert.Equal(t, msg.Nonce, decodedMsg.Nonce)
}

func TestEncodePong(t *testing.T) {
    codec := SimpleCodec{}

    buf := &bytes.Buffer{}   
    msg := &Pong{ Nonce : rand.Uint32()}
    codec.Encode(buf, msg)

    decoded, _ := codec.Decode(buf)

    decodedMsg, ok := decoded.(*Pong)

    assert.True(t, ok, "Decoded type is other than expected")
    assert.Equal(t, msg.Nonce, decodedMsg.Nonce)
}

func TestEncodeDiscover(t *testing.T) {
    codec := SimpleCodec{}

    buf := &bytes.Buffer{}   
    msg := &Discover{}
    codec.Encode(buf, msg)

    decoded, _ := codec.Decode(buf)

    _, ok := decoded.(*Discover)

    assert.True(t, ok, "Decoded type is other than expected")
}

func TestEncodePeers(t *testing.T) {
    codec := SimpleCodec{}

    buf := bytes.NewBuffer(make([]byte, 0))
    addrs := make([]string, 2)
    addrs[0] = "127.0.0.1"
    addrs[1] = "192.168.4.62"
    msg := &Peers{ Addresses: addrs}
    codec.Encode(buf, msg)

    decoded, _ := codec.Decode(buf)

    decodedMsg, ok := decoded.(*Peers)

    assert.True(t, ok, "Decoded type is other than expected")
    assert.EqualValues(t, msg.Addresses, decodedMsg.Addresses)
}

func TestEncodeString(t *testing.T) {
    codec := SimpleCodec{}

    buf := &bytes.Buffer{}   
    msg := "hello"
    codec.Encode(buf, msg)

    decoded, _ := codec.Decode(buf)

    decodedMsg, ok := decoded.(string)

    assert.True(t, ok, "Decoded type is other than expected")
    assert.Equal(t, msg, decodedMsg)
}

func TestEncodeBytes(t *testing.T) {
    codec := SimpleCodec{}

    buf := &bytes.Buffer{}   
    msg := make([]byte, 2)
    msg[0] = 1
    msg[1] = 55
    codec.Encode(buf, msg)

    decoded, _ := codec.Decode(buf)

    decodedMsg, ok := decoded.([]byte)

    assert.True(t, ok, "Decoded type is other than expected")
    assert.EqualValues(t, msg, decodedMsg)
}