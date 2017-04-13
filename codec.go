package network

import (
	"encoding/json"
	"io"

	"github.com/pkg/errors"

	"github.com/veltor/veltor-network/message"
)

// Codec interface.
type Codec interface {
	Encode(w io.Writer, i interface{}) error
	Decode(r io.Reader) (interface{}, error)
}

// DefaultCodec variable.
var DefaultCodec = &SimpleCodec{}

// Msg enum.
const (
	MsgPing = iota
	MsgPong
	MsgDiscover
	MsgPeers
	MsgString
	MsgBytes
)

// SimpleCodec struct.
type SimpleCodec struct{}

// Encode method.
func (s SimpleCodec) Encode(w io.Writer, i interface{}) error {
	code := make([]byte, 1)
	switch i.(type) {
	case *message.Ping:
		code[0] = MsgPing
	case *message.Pong:
		code[0] = MsgPong
	case *message.Discover:
		code[0] = MsgDiscover
	case *message.Peers:
		code[0] = MsgPeers
	case string:
		code[0] = MsgString
	case []byte:
		code[0] = MsgBytes
	default:
		return errors.Errorf("unknown json type (%T)", i)
	}
	_, err := w.Write(code)
	if err != nil {
		return errors.Wrap(err, "could not write json code")
	}
	enc := json.NewEncoder(w)
	err = enc.Encode(i)
	if err != nil {
		return errors.Wrap(err, "could not write json data")
	}
	return nil
}

// Decode method.
func (s SimpleCodec) Decode(r io.Reader) (interface{}, error) {
	code := make([]byte, 1)
	_, err := r.Read(code)
	if err != nil {
		return nil, errors.Wrap(err, "could not read json code")
	}
	var i interface{}
	dec := json.NewDecoder(r)
	switch code[0] {
	case MsgPing:
		var ping message.Ping
		err = dec.Decode(&ping)
		i = &ping
	case MsgPong:
		var pong message.Pong
		err = dec.Decode(&pong)
		i = &pong
	case MsgDiscover:
		var discover message.Discover
		err = dec.Decode(&discover)
		i = &discover
	case MsgPeers:
		var peers message.Peers
		err = dec.Decode(&peers)
		i = &peers
	case MsgString:
		var str string
		err = dec.Decode(&str)
		i = str
	case MsgBytes:
		var bytes []byte
		err = dec.Decode(&bytes)
		i = bytes
	default:
		return nil, errors.Errorf("invalid json code (%T)", code)
	}
	if err != nil {
		return nil, errors.Wrap(err, "could not decode json data")
	}
	return i, nil
}
