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
    "time"
    "go.uber.org/zap"
    "io"
)

func TestSetAddress(t *testing.T) {
    config := DefaultConfig
    addr := "192.168.4.62"
    
    setFunc := SetAddress("192.168.4.62")
    setFunc(&config)

    if config.address != addr {
        t.Fatalf("Expected to set address to %s. Actual: %s", addr, config.address)
    }
}

func TestSetBalance(t *testing.T) {
    config := DefaultConfig
    balance := time.Duration(5)
    
    setFunc := SetBalance(balance)
    setFunc(&config)

    if config.balance != balance {
        t.Fatalf("Expected to set balance to %s. Actual: %s", balance, config.balance)
    }
}

func TestSetBook(t *testing.T) {
    config := DefaultConfig
    book := NewSimpleBook()
    
    setFunc := SetBook(book)
    setFunc(&config)

    if config.book != book {
        t.Fatalf("Expected to set book to %v. Actual: %v", book, config.book)
    }
}

func TestSetCodec(t *testing.T) {
    config := DefaultConfig
    codec := DummyCodec{}
    
    setFunc := SetCodec(codec)
    setFunc(&config)

    if config.codec != codec {
        t.Fatalf("Expected to set codec to %v. Actual: %v", codec, config.codec)
    }
}

func TestSetDiscovery(t *testing.T) {
    config := DefaultConfig
    discovery := time.Duration(5)
    
    setFunc := SetDiscovery(discovery)
    setFunc(&config)

    if config.discovery != discovery {
        t.Fatalf("Expected to set discovery to %s. Actual: %s", discovery, config.discovery)
    }
}

func TestSetHeartbeat(t *testing.T) {
    config := DefaultConfig
    heartbeat := time.Duration(5)
    
    setFunc := SetHeartbeat(heartbeat)
    setFunc(&config)

    if config.heartbeat != heartbeat {
        t.Fatalf("Expected to set heartbeat to %s. Actual: %s", heartbeat, config.heartbeat)
    }
}

func TestSetLog(t *testing.T) {
    config := DefaultConfig
    log, _ := zap.NewDevelopment()
    
    setFunc := SetLog(log)
    setFunc(&config)

    if config.log != log {
        t.Fatalf("Expected to set log to %v. Actual: %v", log, config.log)
    }
}

func TestSetMaxPeers(t *testing.T) {
    config := DefaultConfig
    maxPeers := uint(15)
    
    setFunc := SetMaxPeers(maxPeers)
    setFunc(&config)

    if config.maxPeers != maxPeers {
        t.Fatalf("Expected to set maxPeers to %d. Actual: %d", maxPeers, config.maxPeers)
    }
}

func TestSetMinPeers(t *testing.T) {
    config := DefaultConfig
    minPeers := uint(5)
    
    setFunc := SetMinPeers(minPeers)
    setFunc(&config)

    if config.minPeers != minPeers {
        t.Fatalf("Expected to set minPeers to %d. Actual: %d", minPeers, config.minPeers)
    }
}

func TestSetNetwork(t *testing.T) {
    config := DefaultConfig
    network := make([]byte, 2)
    network[0] = 5
    network[1] = 10
    
    setFunc := SetNetwork(network)
    setFunc(&config)

    if config.network[0] != network[0] || config.network[1] != network[1] {
        t.Fatalf("Expected to set network to %d. Actual: %d", network, config.network)
    }
}

func TestSetServer(t *testing.T) {
    config := DefaultConfig
    server := true
    
    setFunc := SetServer(server)
    setFunc(&config)

    if config.server != server {
        t.Fatalf("Expected to set server to %v. Actual: %v", server, config.server)
    }
}

func TestSetSubscriber(t *testing.T) {
    config := DefaultConfig
    subscriber := make(chan interface{})
    
    setFunc := SetSubscriber(subscriber)
    setFunc(&config)

    if config.subscriber != subscriber {
        t.Fatalf("Expected to set subscriber to %v. Actual: %v", subscriber, config.subscriber)
    }
}

func TestSetTimeout(t *testing.T) {
    config := DefaultConfig
    timeout := time.Duration(5)
    
    setFunc := SetTimeout(timeout)
    setFunc(&config)

    if config.timeout != timeout {
        t.Fatalf("Expected to set timeout to %v. Actual: %v", timeout, config.timeout)
    }
}

type DummyCodec struct{}

// Encode will write the type byte of the given entity to the writer, followed by its JSON encoding.
// It will fail for unknown entities.
func (s DummyCodec) Encode(w io.Writer, i interface{}) error {
	return nil
}

// Decode will use the type byte to initialize the correct entity and then decode the JSON into it.
func (s DummyCodec) Decode(r io.Reader) (interface{}, error) {
	return 1, nil
}