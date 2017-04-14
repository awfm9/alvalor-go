// Copyright (c) 2017 The Veltor Authors
//
// This file is part of Veltor.
//
// Veltor Network is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Veltor Network is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Veltor Network.  If not, see <http://www.gnu.org/licenses/>.

package network

import (
	"bytes"
	"math/rand"
	"net"
	"reflect"
	"time"

	"github.com/pierrec/lz4"
	"github.com/pkg/errors"

	"github.com/veltor/veltor-network/message"
)

// Node struct.
type Node struct {
	log        Log
	book       Book
	codec      Codec
	subscriber chan<- interface{}
	check      time.Duration
	heartbeat  time.Duration
	timeout    time.Duration
	peers      map[string]*peer
}

// NewNode function.
func NewNode(options ...func(*Config)) *Node {
	cfg := DefaultConfig
	for _, option := range options {
		option(&cfg)
	}
	node := &Node{
		log:        cfg.log,
		book:       cfg.book,
		codec:      cfg.codec,
		subscriber: cfg.subscriber,
		check:      cfg.check,
		heartbeat:  cfg.heartbeat,
		timeout:    cfg.timeout,
		peers:      make(map[string]*peer),
	}
	node.book.Blacklist(cfg.address)
	if cfg.listen {
		go node.listen(cfg.address, cfg.maxPeers)
	}
	go node.balance(cfg.minPeers, cfg.maxPeers)
	go node.manage()
	return node
}

// manage method.
func (node *Node) manage() {
Outer:
	for {
		var peers []*peer
		var cases []reflect.SelectCase
		for _, peer := range node.peers {
			peers = append(peers, peer)
			submitter := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(peer.out)}
			cases = append(cases, submitter)
		}
		for _, peer := range node.peers {
			heartbeater := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(peer.hb.C)}
			cases = append(cases, heartbeater)
		}
		node.log.Debugf("rebuilt select cases for %v peers", len(peers))
		if len(cases) == 0 {
			time.Sleep(time.Millisecond * 100)
			continue
		}
		for {
			i, val, ok := reflect.Select(cases)
			if !ok {
				peer := peers[i]
				delete(node.peers, peer.addr)
				peer.close()
				node.book.Dropped(peer.addr)
				node.log.Warningf("dropped failed peer on %v: %v", peer.addr, peer.err)
				continue Outer
			}
			_, ok = val.Interface().(time.Time)
			if ok {
				i = i % len(peers)
				peer := peers[i]
				ping := message.Ping{
					Nonce: rand.Uint32(),
				}
				node.log.Debugf("pinging peer %v", peer.addr)
				err := node.Send(peer.addr, &ping)
				if err != nil {
					node.log.Errorf("could not send ping to %v: %v", peer.addr, err)
				}
				continue
			}
			packet, ok := val.Interface().(*Packet)
			if ok {
				node.log.Debugf("processing packet from peer %v", packet.Address)
				err := node.process(val.Interface().(*Packet))
				if err != nil {
					node.log.Errorf("could not process packet from %v: %v", packet.Address, err)
				}
				continue
			}
		}
	}
}

// process method.
func (node *Node) process(pk *Packet) error {
	switch pk.Message.(type) {
	case *message.Ping:
		return node.processPing(pk)
	case *message.Pong:
		return node.processPong(pk)
	case *message.Discover:
		return node.processDiscover(pk)
	case *message.Peers:
		return node.processPeers(pk)
	default:
		return node.processCustom(pk)
	}
}

// processPing method.
func (node *Node) processPing(pk *Packet) error {
	node.log.Debugf("processing ping from %v", pk.Address)
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
	node.log.Debugf("processing pong from %v", pk.Address)
	node.book.Connected(pk.Address)
	return nil
}

// processDiscover method.
func (node *Node) processDiscover(pk *Packet) error {
	node.log.Debugf("processing discover from %v", pk.Address)
	addrs, err := node.book.Sample()
	if err != nil {
		return errors.Wrap(err, "could not get address sample")
	}
	peers := message.Peers{
		Addresses: addrs,
	}
	err = node.Send(pk.Address, &peers)
	if err != nil {
		return errors.Wrap(err, "could not send peers message")
	}
	return nil
}

// processPeers method.
func (node *Node) processPeers(pk *Packet) error {
	node.log.Debugf("processing peers from %v", pk.Address)
	peers := pk.Message.(*message.Peers)
	for _, addr := range peers.Addresses {
		node.book.Add(addr)
	}
	return nil
}

// processCustom method.
func (node *Node) processCustom(pk *Packet) error {
	node.log.Debugf("processing custom from %v", pk.Address)
	select {
	case node.subscriber <- pk:
		return nil
	default:
		return errors.New("subscriber stalling")
	}
}

// listen method.
func (node *Node) listen(localAddr string, maxPeers uint) {
	_, _, err := net.SplitHostPort(localAddr)
	if err != nil {
		node.log.Errorf("invalid listen address: %v", err)
		return
	}
	ln, err := net.Listen("tcp", localAddr)
	if err != nil {
		node.log.Errorf("could not create listener on %v: %v", localAddr, err)
		return
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			node.log.Errorf("could not accept connection: %v", err)
			break
		}
		remoteAddr := conn.RemoteAddr().String()
		if localAddr == remoteAddr {
			node.log.Warningf("attempted connectiong to self, dropping")
			conn.Close()
			return
		}
		if len(node.peers) > int(maxPeers) {
			node.log.Infof("too many peers, not accepting %v", remoteAddr)
			conn.Close()
			return
		}
		_, ok := node.peers[remoteAddr]
		if ok {
			node.log.Warningf("refusing duplicate incoming peer on %v", remoteAddr)
			conn.Close()
			return
		}
		go node.welcome(conn)
	}
}

// balance method.
func (node *Node) balance(minPeers uint, maxPeers uint) {
	for {
		if uint(len(node.peers)) < minPeers {
			node.log.Infof("adding peer to rebalance")
			node.add()
		}
		if uint(len(node.peers)) > maxPeers {
			node.log.Infof("removing peer to bebalance")
			node.remove()
		}
		time.Sleep(node.check)
	}
}

// add method.
func (node *Node) add() {
	addr, err := node.book.Get()
	if err != nil {
		node.log.Infof("could not get address from book, launching discovery")
		discover := message.Discover{}
		err = node.Broadcast(&discover)
		if err != nil {
			node.log.Errorf("could not broadcast discover: %v", err)
		}
		return
	}
	_, ok := node.peers[addr]
	if ok {
		node.log.Warningf("already connected to peer %v", addr)
		return
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		node.log.Errorf("could not dial address to add peer: %v", err)
		return
	}
	go node.handshake(conn)
}

// remove method.
func (node *Node) remove() {
	index := 0
	goal := rand.Int() % len(node.peers)
	for addr, peer := range node.peers {
		if index != goal {
			index++
			continue
		}
		delete(node.peers, addr)
		peer.close()
		node.book.Disconnected(addr)
		return
	}
	node.log.Errorf("could not find node to remove")
}

// handshake method.
func (node *Node) handshake(conn net.Conn) {
	node.log.Infof("adding outgoing peer on %v", conn.RemoteAddr())
	_, err := conn.Write(MsgSyn)
	if err != nil {
		node.drop(conn)
		return
	}
	ack := make([]byte, 4)
	_, err = conn.Read(ack)
	if err != nil {
		node.drop(conn)
		return
	}
	if !bytes.Equal(ack, MsgAck) {
		node.drop(conn)
		return
	}
	node.init(conn)
}

// welcome method.
func (node *Node) welcome(conn net.Conn) {
	node.log.Infof("adding incoming peer on %v", conn.RemoteAddr())
	syn := make([]byte, 4)
	_, err := conn.Read(syn)
	if err != nil {
		node.drop(conn)
		return
	}
	if !bytes.Equal(syn, MsgSyn) {
		node.drop(conn)
		return
	}
	_, err = conn.Write(MsgAck)
	if err != nil {
		node.drop(conn)
		return
	}
	node.init(conn)
}

// init method.
func (node *Node) init(conn net.Conn) {
	node.log.Infof("finished handshake with %v", conn.RemoteAddr())
	addr := conn.RemoteAddr().String()
	r := lz4.NewReader(conn)
	w := lz4.NewWriter(conn)
	out := make(chan *Packet)
	p := peer{
		conn:      conn,
		addr:      addr,
		r:         r,
		w:         w,
		out:       out,
		codec:     node.codec,
		heartbeat: node.heartbeat,
		timeout:   node.timeout,
		hb:        time.NewTimer(node.heartbeat),
	}
	node.peers[addr] = &p
	node.book.Connected(addr)
	go p.receive()
}

// drop method.
func (node *Node) drop(conn net.Conn) {
	addr := conn.RemoteAddr().String()
	err := conn.Close()
	if err != nil {
		node.log.Errorf("could not close dropped connection: %v", err)
	}
	node.book.Dropped(addr)
}

// Send method.
func (node *Node) Send(addr string, msg interface{}) error {
	peer, ok := node.peers[addr]
	if !ok {
		return errors.New("could not find peer with given address")
	}
	err := peer.send(msg)
	if err != nil {
		node.book.Failed(addr)
		return errors.Wrap(err, "could not send message to peer")
	}
	return nil
}

// Broadcast method.
func (node *Node) Broadcast(msg interface{}) error {
	node.log.Infof("broadcasting message")
	for _, peer := range node.peers {
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
	addrs := make([]string, 0, len(node.peers))
	for _, peer := range node.peers {
		addrs = append(addrs, peer.addr)
	}
	return addrs
}
