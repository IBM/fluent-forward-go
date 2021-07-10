package client

import (
	"fmt"
	"net"
	"time"
)

const (
	DEFAULT_CONNECTION_TIMEOUT time.Duration = 60 * time.Second
)

type Client struct {
	Timeout time.Duration
	Session *Session
}

type ServerAddress struct {
	Hostname string
	Port     int
}

type AuthInfo struct {
	SharedKey []byte
	Username  string
	Password  string
}

type Session struct {
	ServerAddress
	SharedKey  []byte
	Connection net.Conn
	AuthInfo   AuthInfo
}

func (c *Client) Connect(hostname string, port int, auth AuthInfo) error {
	var t time.Duration
	if c.Timeout != 0 {
		t = c.Timeout
	} else {
		t = DEFAULT_CONNECTION_TIMEOUT
	}
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", hostname, port), t)
	if err != nil {
		return err
	}

	// Store the session information for (re)use
	c.Session = &Session{
		ServerAddress: ServerAddress{
			Hostname: hostname,
			Port:     port,
		},
		Connection: conn,
	}

	return nil
}

func (c *Client) Reconnect() error {
	// var t time.Duration
	// if c.Timeout != 0 {
	// 	t = c.Timeout
	// } else {
	// 	t = DEFAULT_CONNECTION_TIMEOUT
	// }
	//
	// if c.Session != nil {
	// 	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.Session.Hostname,
	// 		c.Session.Port), t)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	c.Session.Connection = conn
	// }

	return nil
}

func (c *Client) Disconnect() {
	if c.Session != nil {
		if c.Session.Connection != nil {
			c.Session.Connection.Close()
		}
	}
}

func (c *Client) Handshake() error {

	return nil
}
