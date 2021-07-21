package client

import (
	"fmt"
	"net"
)

type TCPConnectionFactory struct {
	Target ServerAddress
}

func (f *TCPConnectionFactory) New() (net.Conn, error) {
	return net.Dial("tcp", fmt.Sprintf("%s:%d", f.Target.Hostname, f.Target.Port))
}
