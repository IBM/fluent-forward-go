/*
MIT License

Copyright contributors to the fluent-forward-go project

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

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
