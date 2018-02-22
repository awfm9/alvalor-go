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
	"io"
	"net"
	"time"

	"github.com/stretchr/testify/mock"
)

type ErrorMock struct {
	mock.Mock
}

func (em *ErrorMock) Error() string {
	args := em.Called()
	return args.String(0)
}

func (em *ErrorMock) Timeout() bool {
	args := em.Called()
	return args.Bool(0)
}

func (em *ErrorMock) Temporary() bool {
	args := em.Called()
	return args.Bool(0)
}

type AddrMock struct {
	mock.Mock
}

func (am *AddrMock) Network() string {
	args := am.Called()
	return args.String(0)
}

func (am *AddrMock) String() string {
	args := am.Called()
	return args.String(0)
}

type ConnMock struct {
	mock.Mock
}

func (cm *ConnMock) Read(b []byte) (int, error) {
	args := cm.Called(b)
	return args.Int(0), args.Error(1)
}

func (cm *ConnMock) Write(b []byte) (int, error) {
	args := cm.Called(b)
	return args.Int(0), args.Error(1)
}

func (cm *ConnMock) Close() error {
	args := cm.Called()
	return args.Error(0)
}

func (cm *ConnMock) LocalAddr() net.Addr {
	args := cm.Called()
	return args.Get(0).(*AddrMock)
}

func (cm *ConnMock) RemoteAddr() net.Addr {
	args := cm.Called()
	return args.Get(0).(*AddrMock)
}

func (cm *ConnMock) SetDeadline(t time.Time) error {
	args := cm.Called(t)
	return args.Error(0)
}

func (cm *ConnMock) SetReadDeadline(t time.Time) error {
	args := cm.Called(t)
	return args.Error(0)
}

func (cm *ConnMock) SetWriteDeadline(t time.Time) error {
	args := cm.Called(t)
	return args.Error(0)
}

type ListenerMock struct {
	mock.Mock
}

func (lm *ListenerMock) Accept() (net.Conn, error) {
	args := lm.Called()
	var conn net.Conn
	if args.Get(0) != nil {
		conn = args.Get(0).(*ConnMock)
	}
	return conn, args.Error(1)
}

func (lm *ListenerMock) Close() error {
	args := lm.Called()
	return args.Error(0)
}

func (lm *ListenerMock) SetDeadline(t time.Time) error {
	args := lm.Called(t)
	return args.Error(0)
}

type CodecMock struct {
	mock.Mock
}

func (cm *CodecMock) Encode(w io.Writer, i interface{}) error {
	args := cm.Called(w, i)
	return args.Error(0)
}

func (cm *CodecMock) Decode(r io.Reader) (interface{}, error) {
	args := cm.Called(r)
	return args.Get(0), args.Error(1)
}

type DialManagerMock struct {
	mock.Mock
}

func (dm *DialManagerMock) Dial(address string) (net.Conn, error) {
	args := dm.Called(address)
	var conn net.Conn
	if args.Get(0) != nil {
		conn = args.Get(0).(*ConnMock)
	}
	return conn, args.Error(1)
}

type ListenManagerMock struct {
	mock.Mock
}

func (lm *ListenManagerMock) Listen(address string) (Listener, error) {
	args := lm.Called(address)
	var ln Listener
	if args.Get(0) != nil {
		ln = args.Get(0).(*ListenerMock)
	}
	return ln, args.Error(1)
}

type PendingManagerMock struct {
	mock.Mock
}

func (pm *PendingManagerMock) Claim(address string) error {
	args := pm.Called(address)
	return args.Error(0)
}

func (pm *PendingManagerMock) Release(address string) error {
	args := pm.Called(address)
	return args.Error(0)
}

func (pm *PendingManagerMock) Count() uint {
	args := pm.Called()
	return uint(args.Int(0))
}

func (pm *PendingManagerMock) Addresses() []string {
	args := pm.Called()
	var addresses []string
	if args.Get(0) != nil {
		addresses = args.Get(0).([]string)
	}
	return addresses
}

type PeerManagerMock struct {
	mock.Mock
}

func (pm *PeerManagerMock) Add(conn net.Conn, nonce []byte) error {
	args := pm.Called(conn, nonce)
	return args.Error(0)
}

func (pm *PeerManagerMock) Send(address string, msg interface{}) error {
	args := pm.Called(address, msg)
	return args.Error(0)
}

func (pm *PeerManagerMock) Drop(address string) error {
	args := pm.Called(address)
	return args.Error(0)
}

func (pm *PeerManagerMock) Known(nonce []byte) bool {
	args := pm.Called(nonce)
	return args.Bool(0)
}

func (pm *PeerManagerMock) Count() uint {
	args := pm.Called()
	return uint(args.Int(0))
}

func (pm *PeerManagerMock) Addresses() []string {
	args := pm.Called()
	var addresses []string
	if args.Get(0) != nil {
		addresses = args.Get(0).([]string)
	}
	return addresses
}

type ReputationManagerMock struct {
	mock.Mock
}

func (rm *ReputationManagerMock) Failure(address string) {
	_ = rm.Called(address)
}

func (rm *ReputationManagerMock) Success(address string) {
	_ = rm.Called(address)
}

func (rm *ReputationManagerMock) Score(address string) float32 {
	args := rm.Called(address)
	return float32(args.Get(0).(float64))
}

func (rm *ReputationManagerMock) Fail(address string) time.Time {
	args := rm.Called(address)
	return args.Get(0).(time.Time)
}

type HandlerManagerMock struct {
	mock.Mock
}

func (hm *HandlerManagerMock) Dropper() {
	_ = hm.Called()
}

func (hm *HandlerManagerMock) Server() {
	_ = hm.Called()
}

func (hm *HandlerManagerMock) Dialer() {
	_ = hm.Called()
}

func (hm *HandlerManagerMock) Listener() {
	_ = hm.Called()
}

func (hm *HandlerManagerMock) Discoverer() {
	_ = hm.Called()
}

func (hm *HandlerManagerMock) Acceptor(conn net.Conn) {
	_ = hm.Called(conn)
}

func (hm *HandlerManagerMock) Connector(address string) {
	_ = hm.Called(address)
}

func (hm *HandlerManagerMock) Sender(address string, output <-chan interface{}, w io.Writer) {
	_ = hm.Called(address, output, w)
}

func (hm *HandlerManagerMock) Processor(address string, input <-chan interface{}, output chan<- interface{}) {
	_ = hm.Called(address, input, output)
}

func (hm *HandlerManagerMock) Receiver(address string, r io.Reader, input chan<- interface{}) {
	_ = hm.Called(address, r, input)
}

type AddressManagerMock struct {
	mock.Mock
}

func (am *AddressManagerMock) Add(address string) {
	_ = am.Called(address)
}

func (am *AddressManagerMock) Remove(address string) {
	_ = am.Called(address)
}

func (am *AddressManagerMock) Block(address string) {
	_ = am.Called(address)
}

func (am *AddressManagerMock) Unblock(address string) {
	_ = am.Called(address)
}

func (am *AddressManagerMock) Sample(count uint, params ...interface{}) []string {
	args := am.Called(count, params)
	var sample []string
	if args.Get(0) != nil {
		sample = args.Get(0).([]string)
	}
	return sample
}

type EventManagerMock struct {
	mock.Mock
}

func (events *EventManagerMock) Disconnected(addr string) error {
	args := events.Called(addr)
	return args.Error(0)
}
func (events *EventManagerMock) Connected(addr string) error {
	args := events.Called(addr)
	return args.Error(0)
}
func (events *EventManagerMock) Received(addr string, msg interface{}) error {
	args := events.Called(addr, msg)
	return args.Error(0)
}
