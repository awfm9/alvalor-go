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

// Log interface.
type Log interface {
	Tracef(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warningf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Criticalf(format string, v ...interface{})
}

// DefaultLog variable.
var DefaultLog = &NoopLog{}

// NoopLog struct.
type NoopLog struct{}

// Tracef method.
func (s *NoopLog) Tracef(format string, v ...interface{}) {}

// Debugf method.
func (s *NoopLog) Debugf(format string, v ...interface{}) {}

// Infof method.
func (s *NoopLog) Infof(format string, v ...interface{}) {}

// Warningf method.
func (s *NoopLog) Warningf(format string, v ...interface{}) {}

// Errorf method.
func (s *NoopLog) Errorf(format string, v ...interface{}) {}

// Criticalf method.
func (s *NoopLog) Criticalf(format string, v ...interface{}) {}
