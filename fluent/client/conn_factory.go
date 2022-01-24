package client

import (
	"crypto/tls"
	"net"
	"time"
)

// ConnFactory is a light wrapper for net.Dial and tls.Dial. When TLSConfig
// is not nil, tls.Dial is called. Otherwise, net.Dial is used. See Go's
// net.Dial documentation for more information.
type ConnFactory struct {
	Network   string
	Address   string
	TLSConfig *tls.Config
	Timeout   time.Duration
}

func (f *ConnFactory) New() (net.Conn, error) {
	dialer := &net.Dialer{Timeout: f.Timeout}

	if f.TLSConfig != nil {
		return tls.DialWithDialer(dialer, f.Network, f.Address, f.TLSConfig)
	}

	return dialer.Dial(f.Network, f.Address)
}
