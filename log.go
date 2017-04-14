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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

// Log interface.
type Log interface {
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warningf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

// DefaultLog variable.
var DefaultLog = &SimpleLog{w: os.Stderr, prefix: "node"}

// NoopLog variable.
var NoopLog = &SimpleLog{w: ioutil.Discard}

// SimpleLog struct.
type SimpleLog struct {
	lock   sync.Mutex
	prefix string
	w      io.Writer
}

// NewSimpleLog function.
func NewSimpleLog(prefix string) *SimpleLog {
	return &SimpleLog{
		prefix: prefix,
		w:      os.Stderr,
	}
}

// Debugf method.
func (s *SimpleLog) Debugf(format string, v ...interface{}) {
	s.printf("DEBUG", format, v...)
}

// Infof method.
func (s *SimpleLog) Infof(format string, v ...interface{}) {
	s.printf("INFO", format, v...)
}

// Warningf method.
func (s *SimpleLog) Warningf(format string, v ...interface{}) {
	s.printf("WARNING", format, v...)
}

// Errorf method.
func (s *SimpleLog) Errorf(format string, v ...interface{}) {
	s.printf("ERROR", format, v...)
}

// Printf method.
func (s *SimpleLog) printf(level string, format string, v ...interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()
	message := fmt.Sprintf(format, v...)
	fmt.Fprintf(s.w, "%v - (%v) - [%v] %v\n", time.Now().UTC().Format(time.RFC3339), s.prefix, level, message)
}
