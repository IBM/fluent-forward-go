package client

import (
	"errors"
	"fmt"
	"sync"

	"crypto/rand"
	"net"
	"time"

	"github.com/IBM/fluent-forward-go/fluent/protocol"

	"github.com/tinylib/msgp/msgp"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

const (
	DefaultConnectionTimeout time.Duration = 60 * time.Second
)

// MessageClient implementations send MessagePack messages to a peer
//counterfeiter:generate . MessageClient
type MessageClient interface {
	Connect() error
	Disconnect() (err error)
	Reconnect() error
	SendMessage(e protocol.ChunkEncoder) error
	SendRaw(raw []byte) error
}

// ConnectionFactory implementations create new connections
//counterfeiter:generate . ConnectionFactory
type ConnectionFactory interface {
	New() (net.Conn, error)
}

type Client struct {
	ConnectionFactory
	RequireAck bool
	Timeout    time.Duration
	Session    *Session
	AuthInfo   AuthInfo
	Hostname   string
	ackMutex   sync.Mutex
}

type ConnectionOptions struct {
	Factory           ConnectionFactory
	RequireAck        bool
	ConnectionTimeout time.Duration
	// TODO:
	// ReadTimeout       time.Duration
	// WriteTimeout      time.Duration
	AuthInfo AuthInfo
	Hostname string
	Port     int
}

type ServerAddress struct {
	Hostname string
	Port     int
}

func (sa ServerAddress) String() string {
	return fmt.Sprintf("%s:%d", sa.Hostname, sa.Port)
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

func New(opts ConnectionOptions) *Client {
	factory := opts.Factory
	if factory == nil {
		if opts.Hostname == "" {
			opts.Hostname = "localhost"
		}

		if opts.Port == 0 {
			opts.Port = 24224
		}

		factory = &TCPConnectionFactory{
			Target: ServerAddress{
				Hostname: opts.Hostname,
				Port:     opts.Port,
			},
		}
	}

	if opts.ConnectionTimeout == 0 {
		opts.ConnectionTimeout = DefaultConnectionTimeout
	}

	return &Client{
		ConnectionFactory: factory,
		AuthInfo:          opts.AuthInfo,
		Hostname:          opts.Hostname,
		RequireAck:        opts.RequireAck,
		Timeout:           opts.ConnectionTimeout,
	}
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
	if c.Session != nil {
		_ = c.Disconnect()
	}

	return c.Connect()
}

// Disconnect terminates a client connection
func (c *Client) Disconnect() (err error) {
	if c.Session != nil {
		if c.Session.Connection != nil {
			err = c.Session.Connection.Close()
		}
	}

	c.Session = nil

	return
}

func (c *Client) checkAck(chunk string) error {
	if c.Timeout != 0 {
		if err := c.Session.Connection.SetReadDeadline(time.Now().Add(c.Timeout)); err != nil {
			return err
		}
	}

	var ack protocol.AckMessage
	if err := msgp.Decode(c.Session.Connection, &ack); err != nil {
		return err
	}

	if ack.Ack != chunk {
		return fmt.Errorf("Expected chunk %s, but got %s", chunk, ack.Ack)
	}

	return nil
}

// SendMessage sends a single protocol.ChunkEncoder across the wire.  If the session
// is not yet in transport phase, an error is returned, and no message is sent.
func (c *Client) SendMessage(e protocol.ChunkEncoder) error {
	if c.Session == nil {
		return errors.New("no active session")
	}

	if !c.Session.TransportPhase {
		return errors.New("session handshake not completed")
	}

	var (
		chunk string
		err   error
	)

	if c.RequireAck {
		if chunk, err = e.Chunk(); err != nil {
			return err
		}

		c.ackMutex.Lock()
		defer c.ackMutex.Unlock()
	}

	err = msgp.Encode(c.Session.Connection, e)
	if err != nil || !c.RequireAck {
		return err
	}

	return c.checkAck(chunk)
}

// SendRaw sends bytes across the wire. If the session
// is not yet in transport phase, an error is returned,
// and no message is sent.
func (c *Client) SendRaw(m []byte) error {
	return c.SendMessage(protocol.RawMessage(m))
}

// Handshake initiates handshake mode.  Users must call this before attempting
// to send any messages when the server is configured with a shared key, otherwise
// the server will reject any message events.  Successful completion of the
// handshake puts the connection into message (or forward) mode, at which time
// the client is free to send event messages.
func (c *Client) Handshake() error {
	if c.Session == nil || c.Session.Connection == nil {
		return errors.New("not connected")
	}

	var helo protocol.Helo

	r := msgp.NewReader(c.Session.Connection)
	err := helo.DecodeMsg(r)

	if err != nil {
		return err
	}

	salt := make([]byte, 16)

	_, err = rand.Read(salt)
	if err != nil {
		return err
	}

	ping, err := protocol.NewPing(c.Hostname, c.AuthInfo.SharedKey, salt, helo.Options.Nonce)
	if err != nil {
		return err
	}

	err = msgp.Encode(c.Session.Connection, ping)
	if err != nil {
		return err
	}

	var pong protocol.Pong

	err = pong.DecodeMsg(r)
	if err != nil {
		return err
	}

	if err := protocol.ValidatePongDigest(&pong, c.AuthInfo.SharedKey,
		helo.Options.Nonce, salt); err != nil {
		return err
	}

	c.Session.TransportPhase = true

	return nil
}
