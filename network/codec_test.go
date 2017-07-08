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
)

func TestEncodeUnknownType(t *testing.T) {
    codec := SimpleCodec{}

    buf := bytes.NewBuffer(make([]byte, 0))    
    msg := 1
    err := codec.Encode(buf, msg)

    if err == nil {
        t.Fatal("Expected to receive error in case message type is unknown")
    }
}

func TestDecodeUnknownCode(t *testing.T) {
    codec := SimpleCodec{}

    buf := make([]byte, 1)
    buf[0] = 77
    reader := bytes.NewReader(buf)
    _, err := codec.Decode(reader)

    if err == nil {
        t.Fatal("Expected to receive error in case message type is unknown")
    }
}

func TestEncodePing(t *testing.T) {
    codec := SimpleCodec{}

    buf := bytes.NewBuffer(make([]byte, 0))    
    msg := &Ping{ Nonce : rand.Uint32()}
    codec.Encode(buf, msg)

    reader := bytes.NewReader(buf.Bytes())
    decoded, _ := codec.Decode(reader)

    decodedMsg, ok := decoded.(*Ping)

    if !ok {
        t.Fatal("Decoded type is other than expected")
    }

    if msg.Nonce != decodedMsg.Nonce {
        t.Fatalf("Expected: %+v. Actual: %+v.", msg, decoded)
    }
}

func TestEncodePong(t *testing.T) {
    codec := SimpleCodec{}

    buf := bytes.NewBuffer(make([]byte, 0))    
    msg := &Pong{ Nonce : rand.Uint32()}
    codec.Encode(buf, msg)

    reader := bytes.NewReader(buf.Bytes())
    decoded, _ := codec.Decode(reader)

    decodedMsg, ok := decoded.(*Pong)

    if !ok {
        t.Fatal("Decoded type is other than expected")
    }

    if msg.Nonce != decodedMsg.Nonce {
        t.Fatalf("Expected: %+v. Actual: %+v.", msg, decoded)
    }
}

func TestEncodeDiscover(t *testing.T) {
    codec := SimpleCodec{}

    buf := bytes.NewBuffer(make([]byte, 0))    
    msg := &Discover{}
    codec.Encode(buf, msg)

    reader := bytes.NewReader(buf.Bytes())
    decoded, _ := codec.Decode(reader)

    _, ok := decoded.(*Discover)

    if !ok {
        t.Fatal("Decoded type is other than expected")
    }
}

func TestEncodePeers(t *testing.T) {
    codec := SimpleCodec{}

    buf := bytes.NewBuffer(make([]byte, 0))
    addrs := make([]string, 2)
    addrs[0] = "127.0.0.1"
    addrs[1] = "192.168.4.62"
    msg := &Peers{ Addresses: addrs}
    codec.Encode(buf, msg)

    reader := bytes.NewReader(buf.Bytes())
    decoded, _ := codec.Decode(reader)

    decodedMsg, ok := decoded.(*Peers)

    if !ok {
        t.Fatal("Decoded type is other than expected")
    }

    for i := 0; i < len(msg.Addresses); i++ {
         if msg.Addresses[i] != decodedMsg.Addresses[i] {
             t.Fatalf("Expected: %+v. Actual: %+v.", msg, decoded)
        }
    }
}

func TestEncodeString(t *testing.T) {
    codec := SimpleCodec{}

    buf := bytes.NewBuffer(make([]byte, 0))    
    msg := "hello"
    codec.Encode(buf, msg)

    reader := bytes.NewReader(buf.Bytes())
    decoded, _ := codec.Decode(reader)

    decodedMsg, ok := decoded.(string)

    if !ok {
        t.Fatal("Decoded type is other than expected")
    }

    if msg != decodedMsg {
        t.Fatalf("Expected: %+v. Actual: %+v.", msg, decodedMsg)
    }
}

func TestEncodeBytes(t *testing.T) {
    codec := SimpleCodec{}

    buf := bytes.NewBuffer(make([]byte, 0))    
    msg := make([]byte, 2)
    msg[0] = 1
    msg[1] = 55
    codec.Encode(buf, msg)

    reader := bytes.NewReader(buf.Bytes())
    decoded, _ := codec.Decode(reader)

    decodedMsg, ok := decoded.([]byte)

    if !ok {
        t.Fatal("Decoded type is other than expected")
    }

    for i := 0; i < len(msg); i++ {
         if msg[i] != decodedMsg[i] {
             t.Fatalf("Expected: %+v. Actual: %+v.", msg, decodedMsg)
        }
    }
}