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
	"fmt"
	"bytes"
	"math/rand"
	"net"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/pierrec/lz4"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

// Node struct.
type Node struct {
	nonce      []byte
	network    []byte
	log        *zap.Logger
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
			msg, ok := val.Interface().(*Message)
			if ok {
				node.process(msg)
				continue
			}
		}
	}
}

// ping method.
func (node *Node) ping(peer *peer) {
	node.log.Debug("pinging peer on address", zap.String("addr", peer.addr))
	ping := Ping{
		Nonce: rand.Uint32(),
	}
	err := node.Send(peer.addr, &ping)
	if err != nil {
		node.log.Error("could not send ping to address", zap.String("addr", peer.addr), zap.Error(err))
	}
}

// disconnect method.
func (node *Node) disconnect(peer *peer) {
	node.log.Info("disconnecting peer on address", zap.String("addr", peer.addr))
	node.peers.remove(peer.addr)
	peer.close()
	node.book.Dropped(peer.addr)
	atomic.AddInt32(&node.count, -1)
	e := Disconnected{
		Address: peer.addr,
	}
	node.event(&e)
}

// process method.
func (node *Node) process(msg *Message) {
	node.log.Debug("processing message from address", zap.String("message", fmt.Sprintf("%T", msg.Value)), zap.String("addr", msg.Address))
	var err error
	switch msg.Value.(type) {
	case *Ping:
		err = node.processPing(msg)
	case *Pong:
		err = node.processPong(msg)
	case *Discover:
		err = node.processDiscover(msg)
	case *Peers:
		err = node.processPeers(msg)
	default:
		node.event(msg)
	}
	if err != nil {
		node.log.Error("could not process message", zap.String("message", fmt.Sprintf("%T", msg.Value)), zap.Error(err))
	}
}

// processPing method.
func (node *Node) processPing(msg *Message) error {
	ping := msg.Value.(*Ping)
	pong := Pong{
		Nonce: ping.Nonce,
	}
	err := node.Send(msg.Address, &pong)
	if err != nil {
		return errors.Wrap(err, "could not send ping reply")
	}
	return nil
}

// processPong method.
func (node *Node) processPong(msg *Message) error {
	return nil
}

// processDiscover method.
func (node *Node) processDiscover(msg *Message) error {
	addrs, err := node.book.Sample()
	if err != nil {
		return errors.Wrap(err, "could not get address sample")
	}
	err = node.share(msg.Address, addrs)
	if err != nil {
		return errors.Wrap(err, "could not share address sample")
	}
	return nil
}

// processPeers method.
func (node *Node) processPeers(msg *Message) error {
	peers := msg.Value.(*Peers)
	for _, addr := range peers.Addresses {
		node.book.Add(addr)
		if node.peers.has(addr) {
			node.book.Connected(addr)
		}
	}
	return nil
}

// event method.
func (node *Node) event(e interface{}) {
	select {
	case node.subscriber <- e:
	default:
		node.log.Error("subscriber stalling on event", zap.String("event", fmt.Sprintf("%T", e)))
	}
}

// listen method.
func (node *Node) listen() {
	_, _, err := net.SplitHostPort(node.address)
	if err != nil {
		node.log.Error("invalid listen address", zap.Error(err))
		return
	}
	ln, err := net.Listen("tcp", node.address)
	if err != nil {
		node.log.Error("could not create listener on address", zap.String("addr", node.address), zap.Error(err))
		return
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			node.log.Error("could not accept connection", zap.Error(err))
			break
		}
		if node.peers.count() > int(node.maxPeers) {
			node.log.Debug("too many peers, not accepting remote address", zap.String("addr", conn.RemoteAddr().String()))
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
		node.log.Error("already connected to peer on address", zap.String("addr", addr))
		return
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		node.log.Error("could not dial peer on address", zap.String("addr", addr), zap.Error(err))
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
	node.log.Info("launching peer discovery")
	discover := Discover{}
	err := node.Broadcast(&discover)
	if err != nil {
		node.log.Error("could not launch discovery", zap.Error(err))
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
	node.log.Info("adding outgoing peer on address", zap.String("addr", addr))
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
		node.log.Warn("dropping invalid outgoing connection to address", zap.String("addr", addr))
		node.drop(conn)
		return
	}
	node.init(conn, nonce)
}

// welcome method.
func (node *Node) welcome(conn net.Conn) {
	addr := conn.RemoteAddr().String()
	node.log.Info("adding incoming peer on address", zap.String("addr", addr))
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
		node.log.Warn("dropping invalid incoming connection from address", zap.String("addr", addr))
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
	node.log.Info("finalizing handshake with address", zap.String("addr", addr))
	r := lz4.NewReader(conn)
	w := lz4.NewWriter(conn)
	out := make(chan *Message)
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
			node.log.Error("could not share initial address", zap.Error(err))
		}
	}
	e := Connected{
		Address: p.addr,
	}
	node.event(&e)
}

// share method.
func (node *Node) share(addr string, addrs []string) error {
	peers := Peers{
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
	node.log.Info("dropping connection to address", zap.String("addr", addr))
	atomic.AddInt32(&node.count, -1)
	err := conn.Close()
	if err != nil {
		node.log.Error("could not close dropped connection", zap.Error(err))
	}
	node.book.Dropped(addr)
	node.book.Blacklist(addr)
}

// Send method.
func (node *Node) Send(addr string, msg interface{}) error {
	node.log.Debug("sending message to address", zap.String("message", fmt.Sprintf("%T", msg)), zap.String("addr", addr))
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
	node.log.Debug("broadcasting message", zap.String("message", fmt.Sprintf("%T", msg)))
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
