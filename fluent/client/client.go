package client

import (
	"errors"
	// "fmt"
	"math/rand"
	"net"
	"time"

	"github.ibm.com/Observability/fluent-forward-go/fluent/protocol"

	"github.com/tinylib/msgp/msgp"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

const (
	DEFAULT_CONNECTION_TIMEOUT time.Duration = 60 * time.Second
)

type Client struct {
	ConnectionFactory
	Timeout  time.Duration
	Session  *Session
	AuthInfo AuthInfo
	Hostname string
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
	Connection     net.Conn
	TransportPhase bool
}

// ConnectionFactory implementations create new connections
//counterfeiter:generate . ConnectionFactory
type ConnectionFactory interface {
	New() (net.Conn, error)
	Session() (*Session, error)
}

// Connect initializes the Session and Connection objects by opening
// a client connect to the target configured in the ConnectionFactory
func (c *Client) Connect() error {
	conn, err := c.New()
	if err != nil {
		return err
	}

	c.Session = &Session{
		Connection: conn,
	}

	// If no shared key, handshake mode is not required
	if c.AuthInfo.SharedKey == nil {
		c.Session.TransportPhase = true
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

// Disconnect terminates a client connection
func (c *Client) Disconnect() {
	if c.Session != nil {
		if c.Session.Connection != nil {
			c.Session.Connection.Close()
		}
	}

	c.Session = nil
}

// SendMessage sends a single msgp.Encodable across the wire.  If the session
// is not yet in transport phase, an error is returned, and no message is sent.
func (c *Client) SendMessage(e msgp.Encodable) error {
	if !c.Session.TransportPhase {
		return errors.New("Session handshake not completed")
	}

	return c.sendMessage(e)
}

// Private sender, bypasses transport phase checks.
func (c *Client) sendMessage(e msgp.Encodable) error {
	w := msgp.NewWriter(c.Session.Connection)
	e.EncodeMsg(w)
	w.Flush()
	return nil
}

// Handshake initiates handshake mode.  Users must call this before attempting
// to send any messages when the server is configured with a shared key, otherwise
// the server will reject any message events.  Successful completion of the
// handshake puts the connection into message (or forward) mode, at which time
// the client is free to send event messages.
func (c *Client) Handshake() error {
	if c.Session == nil || c.Session.Connection == nil {
		return errors.New("Not connected")
	}

	var helo protocol.Helo
	r := msgp.NewReader(c.Session.Connection)
	helo.DecodeMsg(r)

	salt := make([]byte, 16)
	rand.Read(salt)

	c.sendMessage(protocol.NewPing(c.Hostname, c.AuthInfo.SharedKey, salt, helo.Options.Nonce))
	var pong protocol.Pong
	pong.DecodeMsg(r)

	if err := protocol.ValidatePongDigest(&pong, c.AuthInfo.SharedKey,
		helo.Options.Nonce, salt); err != nil {
		return err
	}

	c.Session.TransportPhase = true
	return nil
}
