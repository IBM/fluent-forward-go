package client

import (
	"crypto/tls"
	"net"
)

// SocketFactory is a light wrapper for net.Dial and tls.Dial. When TLSConfig
// is not nil, tls.Dial is called. Otherwise, net.Dial is used. See Go's
// net.Dial documentation for more information.
type SocketFactory struct {
	Network   string
	Address   string
	TLSConfig *tls.Config
}

func (f *SocketFactory) New() (net.Conn, error) {
	if f.TLSConfig != nil {
		return tls.Dial(f.Network, f.Address, f.TLSConfig)
	}

	return net.Dial(f.Network, f.Address)
}

func (f *SocketFactory) Session() (*Session, error) {
	conn, _ := f.New()

	return &Session{
		Connection: conn,
	}, nil
}
