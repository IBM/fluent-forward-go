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
	// Network indicates the type of connection. The default value is "tcp".
	Network   string
	Address   string
	TLSConfig *tls.Config
	Timeout   time.Duration
}

func (f *ConnFactory) New() (net.Conn, error) {
	if len(f.Network) == 0 {
		f.Network = "tcp"
	}

	dialer := &net.Dialer{Timeout: f.Timeout}

	if f.TLSConfig != nil {
		return tls.DialWithDialer(dialer, f.Network, f.Address, f.TLSConfig)
	}

	return dialer.Dial(f.Network, f.Address)
}
