package fluent

import (
	"github.com/vmihailenco/msgpack/v5"
)

const (
	MSGTYPE_HELO = "HELO"
	MSGTYPE_PING = "PING"
	MSGTYPE_PONG = "PONG"
)

// NewHelo returns a msgpack-encoded Helo message with the specified options.
// if opts is nil, then a nonce is generated, auth is left empty, and
// keepalive is true.
func NewHelo(opts *HeloOpts) ([]byte, error) {
	h := Helo{MessageType: MSGTYPE_HELO}
	if opts == nil {
		h.Options = &HeloOpts{
			Keepalive: true,
		}
	}
	return msgpack.Marshal(h)
}

type Helo struct {
	_msgpack    struct{} `msgpack:",as_array"`
	MessageType string
	Options     *HeloOpts
}

type HeloOpts struct {
	Nonce     string `msgpack:"nonce"`
	Auth      string `msgpack:"auth"`
	Keepalive bool   `msgpack:"keepalive"`
}
