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

package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logger is a logger that wraps around a standard io.Writer and uses the
// standard library log.Logger to output log messages.
type Logger struct {
	level  Level
	stdlog *log.Logger
}

// Level defines the logging levels.
type Level uint8

// Level enum to define the different logging levels.
const (
	Trace Level = iota
	Debug
	Info
	Notice
	Warning
	Error
	Critical
)

// New creates a new logger with info level.
func New() *Logger {
	logger := &Logger{level: Info}
	logger.stdlog = log.New(logger, "", log.Lshortfile)
	return logger
}

// SetLevel allows us to adjust the log level at runtime.
func (log *Logger) SetLevel(level Level) {
	log.level = level
}

// Write provides the interface for the standard library logger.
func (log *Logger) Write(bytes []byte) (int, error) {
	return fmt.Fprintf(os.Stderr, "%v %v", time.Now().UTC().Format(time.RFC3339), string(bytes))
}

// Criticalf logs a critical message.
func (log *Logger) Criticalf(format string, v ...interface{}) {
	if log.level <= Critical {
		format = "[CRITICAL] " + format
		log.stdlog.Output(2, fmt.Sprintf(format, v...))
	}
}

// Errorf logs an error message.
func (log *Logger) Errorf(format string, v ...interface{}) {
	if log.level <= Error {
		format = "[ERROR] " + format
		log.stdlog.Output(2, fmt.Sprintf(format, v...))
	}
}

// Warningf logs a warning message.
func (log *Logger) Warningf(format string, v ...interface{}) {
	if log.level <= Warning {
		format = "[WARNING] " + format
		log.stdlog.Output(2, fmt.Sprintf(format, v...))
	}
}

// Noticef logs a notice message.
func (log *Logger) Noticef(format string, v ...interface{}) {
	if log.level <= Notice {
		format = "[NOTICE] " + format
		log.stdlog.Output(2, fmt.Sprintf(format, v...))
	}
}

// Infof logs an info message.
func (log *Logger) Infof(format string, v ...interface{}) {
	if log.level <= Info {
		format = "[INFO] " + format
		log.stdlog.Output(2, fmt.Sprintf(format, v...))
	}
}

// Debugf logs a debug message.
func (log *Logger) Debugf(format string, v ...interface{}) {
	if log.level <= Debug {
		format = "[DEBUG] " + format
		log.stdlog.Output(2, fmt.Sprintf(format, v...))
	}
}

// Tracef logs a trace message.
func (log *Logger) Tracef(format string, v ...interface{}) {
	if log.level <= Trace {
		format = "[TRACE] " + format
		log.stdlog.Output(2, fmt.Sprintf(format, v...))
	}
}
