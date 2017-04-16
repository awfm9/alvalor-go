// Copyright (c) 2017 The Veltor Authors
//
// This file is part of Veltor.
//
// Veltor is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Veltor is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Veltor.  If not, see <http://www.gnu.org/licenses/>.

package network

import (
	"bytes"
	"math/rand"
	"net"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/pierrec/lz4"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"

	"github.com/veltor/veltor-go/message"
)

// Node struct.
type Node struct {
	nonce      []byte
	network    []byte
	log        Log
	book       Book
	codec      Codec
	subscriber chan<- interface{}
	server     bool
	address    string
	minPeers   uint
	maxPeers   uint
	balance    time.Duration
	heartbeat  time.Duration
	timeout    time.Duration
	discovery  *time.Ticker
	count      int32
	peers      *registry
}

// NewNode function.
func NewNode(options ...func(*Config)) *Node {
	cfg := DefaultConfig
	for _, option := range options {
		option(&cfg)
	}
	node := &Node{
		nonce:      uuid.NewV4().Bytes(),
		network:    cfg.network,
		log:        cfg.log,
		book:       cfg.book,
		codec:      cfg.codec,
		subscriber: cfg.subscriber,
		server:     cfg.server,
		address:    cfg.address,
		minPeers:   cfg.minPeers,
		maxPeers:   cfg.maxPeers,
		balance:    cfg.balance,
		heartbeat:  cfg.heartbeat,
		timeout:    cfg.timeout,
		discovery:  time.NewTicker(cfg.discovery),
		peers:      &registry{peers: make(map[string]*peer)},
	}
	node.book.Blacklist(cfg.address)
	if cfg.server {
		go node.listen()
	}
	go node.check()
	go node.manage()
	return node
}

// manage method.
func (node *Node) manage() {
Outer:
	for {
		peers := node.peers.slice()
		var cases []reflect.SelectCase
		for _, peer := range peers {
			peers = append(peers, peer)
			submitter := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(peer.out)}
			cases = append(cases, submitter)
		}
		for _, peer := range peers {
			heartbeater := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(peer.hb.C)}
			cases = append(cases, heartbeater)
		}
		if len(cases) == 0 {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		for {
			i, val, ok := reflect.Select(cases)
			if !ok {
				node.disconnect(peers[i])
				continue Outer
			}
			_, ok = val.Interface().(time.Time)
			if ok {
				node.ping(peers[i%len(peers)])
				continue
			}
			pk, ok := val.Interface().(*Packet)
			if ok {
				node.process(pk)
				continue
			}
		}
	}
}

// ping method.
func (node *Node) ping(peer *peer) {
	node.log.Debugf("pinging peer on %v", peer.addr)
	ping := message.Ping{
		Nonce: rand.Uint32(),
	}
	err := node.Send(peer.addr, &ping)
	if err != nil {
		node.log.Errorf("could not send ping to %v: %v", peer.addr, err)
	}
}

// disconnect method.
func (node *Node) disconnect(peer *peer) {
	node.log.Infof("disconnecting peer on %v", peer.addr)
	node.peers.remove(peer.addr)
	peer.close()
	node.book.Dropped(peer.addr)
	atomic.AddInt32(&node.count, -1)
}

// process method.
func (node *Node) process(pk *Packet) {
	node.log.Debugf("processing %T message from %v", pk.Message, pk.Address)
	var err error
	switch pk.Message.(type) {
	case *message.Ping:
		err = node.processPing(pk)
	case *message.Pong:
		err = node.processPong(pk)
	case *message.Discover:
		err = node.processDiscover(pk)
	case *message.Peers:
		err = node.processPeers(pk)
	default:
		err = node.processCustom(pk)
	}
	if err != nil {
		node.log.Errorf("could not process %T message: %v", pk.Message, err)
	}
}

// processPing method.
func (node *Node) processPing(pk *Packet) error {
	ping := pk.Message.(*message.Ping)
	pong := message.Pong{
		Nonce: ping.Nonce,
	}
	err := node.Send(pk.Address, &pong)
	if err != nil {
		return errors.Wrap(err, "could not send ping reply")
	}
	return nil
}

// processPong method.
func (node *Node) processPong(pk *Packet) error {
	return nil
}

// processDiscover method.
func (node *Node) processDiscover(pk *Packet) error {
	addrs, err := node.book.Sample()
	if err != nil {
		return errors.Wrap(err, "could not get address sample")
	}
	err = node.share(pk.Address, addrs)
	if err != nil {
		return errors.Wrap(err, "could not share address sample")
	}
	return nil
}

// processPeers method.
func (node *Node) processPeers(pk *Packet) error {
	peers := pk.Message.(*message.Peers)
	for _, addr := range peers.Addresses {
		node.book.Add(addr)
		if node.peers.has(addr) {
			node.book.Connected(addr)
		}
	}
	return nil
}

// processCustom method.
func (node *Node) processCustom(pk *Packet) error {
	select {
	case node.subscriber <- pk:
		return nil
	default:
		return errors.New("subscriber stalling")
	}
}

// listen method.
func (node *Node) listen() {
	_, _, err := net.SplitHostPort(node.address)
	if err != nil {
		node.log.Errorf("invalid listen address: %v", err)
		return
	}
	ln, err := net.Listen("tcp", node.address)
	if err != nil {
		node.log.Errorf("could not create listener on %v: %v", node.address, err)
		return
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			node.log.Errorf("could not accept connection: %v", err)
			break
		}
		if node.peers.count() > int(node.maxPeers) {
			node.log.Debugf("too many peers, not accepting %v", conn.RemoteAddr())
			conn.Close()
			return
		}
		go node.welcome(conn)
	}
}

// check method.
func (node *Node) check() {
	for {
		count := uint(atomic.LoadInt32(&node.count))
		if count < node.minPeers {
			node.add()
		}
		if count > node.maxPeers {
			node.remove()
		}
		time.Sleep(node.balance)
	}
}

// add method.
func (node *Node) add() {
	addr, err := node.book.Get()
	if err != nil {
		node.discover()
		return
	}
	if node.peers.has(addr) {
		node.log.Errorf("already connected to peer %v", addr)
		return
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		node.log.Errorf("could not dial peer %v: %v", addr, err)
		return
	}
	go node.handshake(conn)
}

// discover method.
func (node *Node) discover() {
	select {
	case <-node.discovery.C:
	default:
		return
	}
	node.log.Infof("launching peer discovery")
	discover := message.Discover{}
	err := node.Broadcast(&discover)
	if err != nil {
		node.log.Errorf("could not launch discovery: %v", err)
	}
}

// remove method.
func (node *Node) remove() {
	index := 0
	goal := rand.Int() % node.peers.count()
	for _, peer := range node.peers.slice() {
		if index != goal {
			index++
			continue
		}
		node.disconnect(peer)
		return
	}
}

// known method.
func (node *Node) known(nonce []byte) bool {
	for _, peer := range node.peers.slice() {
		if bytes.Equal(nonce, peer.nonce) {
			return true
		}
	}
	return false
}

// handshake method.
func (node *Node) handshake(conn net.Conn) {
	addr := conn.RemoteAddr().String()
	node.log.Infof("adding outgoing peer on %v", addr)
	atomic.AddInt32(&node.count, 1)
	syn := append(node.network, node.nonce...)
	_, err := conn.Write(syn)
	if err != nil {
		node.drop(conn)
		return
	}
	ack := make([]byte, len(syn))
	_, err = conn.Read(ack)
	if err != nil {
		node.drop(conn)
		return
	}
	code := ack[:len(node.network)]
	nonce := ack[len(node.network):]
	if !bytes.Equal(code, node.network) || bytes.Equal(nonce, node.nonce) || node.known(nonce) {
		node.log.Warningf("dropping invalid outgoing connection to %v", addr)
		node.drop(conn)
		return
	}
	node.init(conn, nonce)
}

// welcome method.
func (node *Node) welcome(conn net.Conn) {
	addr := conn.RemoteAddr().String()
	node.log.Infof("adding incoming peer on %v", addr)
	atomic.AddInt32(&node.count, 1)
	ack := append(node.network, node.nonce...)
	syn := make([]byte, len(ack))
	_, err := conn.Read(syn)
	if err != nil {
		node.drop(conn)
		return
	}
	code := syn[:len(node.network)]
	nonce := syn[len(node.network):]
	if !bytes.Equal(code, node.network) || bytes.Equal(nonce, node.nonce) || node.known(nonce) {
		node.log.Warningf("dropping invalid incoming connection from %v", addr)
		node.drop(conn)
		return
	}
	_, err = conn.Write(ack)
	if err != nil {
		node.drop(conn)
		return
	}
	node.init(conn, nonce)
}

// init method.
func (node *Node) init(conn net.Conn, nonce []byte) {
	addr := conn.RemoteAddr().String()
	node.log.Infof("finalizing handshake with %v", addr)
	r := lz4.NewReader(conn)
	w := lz4.NewWriter(conn)
	out := make(chan *Packet)
	p := peer{
		conn:      conn,
		addr:      addr,
		nonce:     nonce,
		r:         r,
		w:         w,
		out:       out,
		codec:     node.codec,
		heartbeat: node.heartbeat,
		timeout:   node.timeout,
		hb:        time.NewTimer(node.heartbeat),
	}
	node.peers.add(addr, &p)
	node.book.Connected(addr)
	go p.receive()
	if node.server {
		err := node.share(addr, []string{node.address})
		if err != nil {
			node.log.Errorf("could not share initial address: %v", err)
		}
	}
}

// share method.
func (node *Node) share(addr string, addrs []string) error {
	peers := message.Peers{
		Addresses: addrs,
	}
	err := node.Send(addr, &peers)
	if err != nil {
		return errors.Wrap(err, "could not send peers message")
	}
	return nil
}

// drop method.
func (node *Node) drop(conn net.Conn) {
	addr := conn.RemoteAddr().String()
	node.log.Infof("dropping connection to %v", addr)
	atomic.AddInt32(&node.count, -1)
	err := conn.Close()
	if err != nil {
		node.log.Errorf("could not close dropped connection: %v", err)
	}
	node.book.Dropped(addr)
	node.book.Blacklist(addr)
}

// Send method.
func (node *Node) Send(addr string, msg interface{}) error {
	node.log.Debugf("sending %T message to %v", msg, addr)
	if !node.peers.has(addr) {
		return errors.New("could not find peer with given address")
	}
	peer := node.peers.get(addr)
	err := peer.send(msg)
	if err != nil {
		node.book.Failed(addr)
		return errors.Wrap(err, "could not send message to peer")
	}
	return nil
}

// Broadcast method.
func (node *Node) Broadcast(msg interface{}) error {
	node.log.Debugf("broadcasting %T message", msg)
	for _, peer := range node.peers.slice() {
		err := peer.send(msg)
		if err != nil {
			node.book.Failed(peer.addr)
			return errors.Wrapf(err, "could not broadcast message to peer %v", peer.addr)
		}
	}
	return nil
}

// Peers method.
func (node *Node) Peers() []string {
	peers := node.peers.slice()
	addrs := make([]string, 0, len(peers))
	for _, peer := range peers {
		addrs = append(addrs, peer.addr)
	}
	return addrs
}
