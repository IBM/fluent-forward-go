package client

import (
	"net"
)

// SocketFactory is a light wrapper for net.Dial. It can be used to
// connect to network types described in the net.Dial documentation.
type SocketFactory struct {
	Network string
	Address string
}

func (f *SocketFactory) New() (net.Conn, error) {
	return net.Dial(f.Network, f.Address)
}

func (f *SocketFactory) Session() (*Session, error) {
	conn, _ := f.New()

	return &Session{
		Connection: conn,
	}, nil
}
