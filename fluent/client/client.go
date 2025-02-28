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
	DefaultDialTimeout  time.Duration = 10 * time.Second
	DefaultReadTimeout  time.Duration = 30 * time.Second
	DefaultWriteTimeout time.Duration = 30 * time.Second
)

// MessageClient implementations send MessagePack messages to a peer
//
//counterfeiter:generate . MessageClient
type MessageClient interface {
	Connect() error
	Disconnect() (err error)
	Reconnect() error
	Send(e protocol.ChunkEncoder) error
	SendCompressed(tag string, entries protocol.EntryList) error
	SendCompressedFromBytes(tag string, entries []byte) error
	SendForward(tag string, entries protocol.EntryList) error
	SendMessage(tag string, record interface{}) error
	SendMessageExt(tag string, record interface{}) error
	SendPacked(tag string, entries protocol.EntryList) error
	SendPackedFromBytes(tag string, entries []byte) error
	SendRaw(raw []byte) error
}

// ConnectionFactory implementations create new connections
//
//counterfeiter:generate . ConnectionFactory
type ConnectionFactory interface {
	New() (net.Conn, error)
}

type Client struct {
	ConnectionFactory
	RequireAck   bool
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	AuthInfo     AuthInfo
	Hostname     string
	session      *Session
	ackLock      sync.Mutex
	sessionLock  sync.RWMutex
}

type ConnectionOptions struct {
	Factory      ConnectionFactory
	RequireAck   bool
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	AuthInfo     AuthInfo
}

type AuthInfo struct {
	SharedKey []byte
	Username  string
	Password  string
}

type Session struct {
	Connection     net.Conn
	TransportPhase bool
}

func New(opts ConnectionOptions) *Client {
	factory := opts.Factory
	if factory == nil {
		factory = &ConnFactory{
			Network: "tcp",
			Address: "localhost:24224",
		}
	}

	// Set default timeouts if not provided
	if opts.DialTimeout == 0 {
		opts.DialTimeout = DefaultDialTimeout
	}

	if opts.ReadTimeout == 0 {
		opts.ReadTimeout = DefaultReadTimeout
	}

	if opts.WriteTimeout == 0 {
		opts.WriteTimeout = DefaultWriteTimeout
	}

	return &Client{
		ConnectionFactory: factory,
		AuthInfo:          opts.AuthInfo,
		RequireAck:        opts.RequireAck,
		DialTimeout:       opts.DialTimeout,
		ReadTimeout:       opts.ReadTimeout,
		WriteTimeout:      opts.WriteTimeout,
	}
}

// TransportPhase indicates if the client has completed the
// initial connection handshake.
func (c *Client) TransportPhase() bool {
	c.sessionLock.RLock()
	defer c.sessionLock.RUnlock()

	return c.session != nil && c.session.TransportPhase
}

func (c *Client) connect() error {
	conn, err := c.New()
	if err != nil {
		return err
	}

	c.session = &Session{
		Connection: conn,
	}

	// If no shared key, handshake mode is not required
	if c.AuthInfo.SharedKey == nil {
		c.session.TransportPhase = true
	}

	return nil
}

// Handshake initiates handshake mode.  Users must call this before attempting
// to send any messages when the server is configured with a shared key, otherwise
// the server will reject any message events.  Successful completion of the
// handshake puts the connection into message (or forward) mode, at which time
// the client is free to send event messages.
func (c *Client) Handshake() error {
	c.sessionLock.RLock()
	defer c.sessionLock.RUnlock()

	if c.session == nil {
		return errors.New("not connected")
	}

	var helo protocol.Helo

	// apply read timeout for reading helo message
	if c.ReadTimeout != 0 {
		if err := c.session.Connection.SetReadDeadline(time.Now().Add(c.ReadTimeout)); err != nil {
			return err
		}
	}

	r := msgp.NewReader(c.session.Connection)
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

	// apply write timeout for sending the ping message
	if c.WriteTimeout != 0 {
		if err := c.session.Connection.SetWriteDeadline(time.Now().Add(c.WriteTimeout)); err != nil {
			return err
		}
	}

	err = msgp.Encode(c.session.Connection, ping)
	if err != nil {
		return err
	}

	var pong protocol.Pong

	// apply read timeout for receiving the pong message
	if c.ReadTimeout != 0 {
		if err := c.session.Connection.SetReadDeadline(time.Now().Add(c.ReadTimeout)); err != nil {
			return err
		}
	}

	err = pong.DecodeMsg(r)
	if err != nil {
		return err
	}

	if err := protocol.ValidatePongDigest(&pong, c.AuthInfo.SharedKey,
		helo.Options.Nonce, salt); err != nil {
		return err
	}

	c.session.TransportPhase = true

	return nil
}

// Connect initializes the Session and Connection objects by opening
// a client connect to the target configured in the ConnectionFactory
func (c *Client) Connect() error {
	c.sessionLock.Lock()
	defer c.sessionLock.Unlock()

	if c.session != nil {
		return errors.New("a session is already active")
	}

	return c.connect()
}

func (c *Client) disconnect() (err error) {
	if c.session != nil {
		err = c.session.Connection.Close()
	}

	c.session = nil

	return
}

// Disconnect terminates a client connection
func (c *Client) Disconnect() error {
	c.sessionLock.Lock()
	defer c.sessionLock.Unlock()

	return c.disconnect()
}

func (c *Client) Reconnect() error {
	c.sessionLock.Lock()
	defer c.sessionLock.Unlock()

	_ = c.disconnect()

	return c.connect()
}

func (c *Client) checkAck(chunk string) error {
	if c.ReadTimeout != 0 {
		if err := c.session.Connection.SetReadDeadline(time.Now().Add(c.ReadTimeout)); err != nil {
			return err
		}
	}

	var ack protocol.AckMessage
	if err := msgp.Decode(c.session.Connection, &ack); err != nil {
		return err
	}

	if ack.Ack != chunk {
		return fmt.Errorf("Expected chunk %s, but got %s", chunk, ack.Ack)
	}

	return nil
}

// Send sends a single protocol.ChunkEncoder across the wire.  If the session
// is not yet in transport phase, an error is returned, and no message is sent.
func (c *Client) Send(e protocol.ChunkEncoder) error {
	c.sessionLock.RLock()
	defer c.sessionLock.RUnlock()

	if c.session == nil {
		return errors.New("no active session")
	}

	if !c.session.TransportPhase {
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

		c.ackLock.Lock()
		defer c.ackLock.Unlock()
	}

	// apply write timeout for sending the message
	if c.WriteTimeout != 0 {
		if err := c.session.Connection.SetWriteDeadline(time.Now().Add(c.WriteTimeout)); err != nil {
			return err
		}
	}

	err = msgp.Encode(c.session.Connection, e)
	if err != nil || !c.RequireAck {
		return err
	}

	return c.checkAck(chunk)
}

// SendRaw sends bytes across the wire. If the session
// is not yet in transport phase, an error is returned,
// and no message is sent.
func (c *Client) SendRaw(m []byte) error {
	c.sessionLock.RLock()
	defer c.sessionLock.RUnlock()

	if c.session == nil {
		return errors.New("no active session")
	}

	if !c.session.TransportPhase {
		return errors.New("session handshake not completed")
	}

	// apply write timeout for write
	if c.WriteTimeout != 0 {
		if err := c.session.Connection.SetWriteDeadline(time.Now().Add(c.WriteTimeout)); err != nil {
			return err
		}
	}

	_, err := c.session.Connection.Write(m)

	return err
}

func (c *Client) SendPacked(tag string, entries protocol.EntryList) error {
	msg, err := protocol.NewPackedForwardMessage(tag, entries)
	if err == nil {
		err = c.Send(msg)
	}

	return err
}

func (c *Client) SendPackedFromBytes(tag string, entries []byte) error {
	msg := protocol.NewPackedForwardMessageFromBytes(tag, entries)

	return c.Send(msg)
}

func (c *Client) SendMessage(tag string, record interface{}) error {
	msg := protocol.NewMessage(tag, record)

	return c.Send(msg)
}

func (c *Client) SendMessageExt(tag string, record interface{}) error {
	msg := protocol.NewMessageExt(tag, record)

	return c.Send(msg)
}

func (c *Client) SendForward(tag string, entries protocol.EntryList) error {
	msg := protocol.NewForwardMessage(tag, entries)

	return c.Send(msg)
}

func (c *Client) SendCompressed(tag string, entries protocol.EntryList) error {
	msg, err := protocol.NewCompressedPackedForwardMessage(tag, entries)
	if err == nil {
		err = c.Send(msg)
	}

	return err
}

func (c *Client) SendCompressedFromBytes(tag string, entries []byte) error {
	msg, err := protocol.NewCompressedPackedForwardMessageFromBytes(tag, entries)
	if err == nil {
		err = c.Send(msg)
	}

	return err
}
