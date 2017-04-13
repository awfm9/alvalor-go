package network

import (
	"time"
)

// DefaultConfig variable.
var DefaultConfig = Config{
	log:        DefaultLog,
	book:       DefaultBook,
	codec:      DefaultCodec,
	subscriber: nil,
	listen:     false,
	address:    "",
	minPeers:   3,
	maxPeers:   10,
	check:      time.Second * 2,
	heartbeat:  time.Second * 1,
	timeout:    time.Second * 5,
}

// Config struct.
type Config struct {
	log        Log
	book       Book
	codec      Codec
	subscriber chan<- interface{}
	listen     bool
	address    string
	minPeers   uint
	maxPeers   uint
	check      time.Duration
	heartbeat  time.Duration
	timeout    time.Duration
}

// SetLog function.
func SetLog(log Log) func(*Config) {
	return func(cfg *Config) {
		cfg.log = log
	}
}

// SetBook function.
func SetBook(book Book) func(*Config) {
	return func(cfg *Config) {
		cfg.book = book
	}
}

// SetCodec function.
func SetCodec(codec Codec) func(*Config) {
	return func(cfg *Config) {
		cfg.codec = codec
	}
}

// SetSubscriber function.
func SetSubscriber(sub chan<- interface{}) func(*Config) {
	return func(cfg *Config) {
		cfg.subscriber = sub
	}
}

// SetListen function.
func SetListen(listen bool) func(*Config) {
	return func(cfg *Config) {
		cfg.listen = listen
	}
}

// SetAddress function.
func SetAddress(address string) func(*Config) {
	return func(cfg *Config) {
		cfg.address = address
	}
}

// SetMinPeers function.
func SetMinPeers(minPeers uint) func(*Config) {
	return func(cfg *Config) {
		cfg.minPeers = minPeers
	}
}

// SetMaxPeers function.
func SetMaxPeers(maxPeers uint) func(*Config) {
	return func(cfg *Config) {
		cfg.maxPeers = maxPeers
	}
}

// SetCheck function.
func SetCheck(check time.Duration) func(*Config) {
	return func(cfg *Config) {
		cfg.check = check
	}
}

// SetHeartbeat function.
func SetHeartbeat(heartbeat time.Duration) func(*Config) {
	return func(cfg *Config) {
		cfg.heartbeat = heartbeat
	}
}

// SetTimeout function.
func SetTimeout(timeout time.Duration) func(*Config) {
	return func(cfg *Config) {
		cfg.timeout = timeout
	}
}
