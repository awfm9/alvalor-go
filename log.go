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
