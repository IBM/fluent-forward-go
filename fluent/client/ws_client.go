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
	"errors"
	"net/http"
	"sync"

	"github.com/IBM/fluent-forward-go/fluent/client/ws"
	"github.com/IBM/fluent-forward-go/fluent/client/ws/ext"
	"github.com/IBM/fluent-forward-go/fluent/protocol"
	"github.com/gorilla/websocket"

	"github.com/tinylib/msgp/msgp"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

const (
	AuthorizationHeader = "Authorization"
)

// Expose message types as defined in underlying websocket library
const (
	TextMessage   = websocket.TextMessage
	BinaryMessage = websocket.BinaryMessage
)

//counterfeiter:generate . WSConnectionFactory
type WSConnectionFactory interface {
	New() (ext.Conn, error)
	NewSession(ws.Connection) *WSSession
}

type IAMAuthInfo struct {
	token string
	mutex sync.RWMutex
}

// IAMToken returns the current token value. It is thread safe.
func (ai *IAMAuthInfo) IAMToken() string {
	ai.mutex.RLock()
	defer ai.mutex.RUnlock()

	return ai.token
}

// SetIAMToken updates the token returned by IAMToken(). It is thread safe.
func (ai *IAMAuthInfo) SetIAMToken(token string) {
	ai.mutex.Lock()
	defer ai.mutex.Unlock()

	ai.token = token
}

func NewIAMAuthInfo(token string) *IAMAuthInfo {
	return &IAMAuthInfo{token: token}
}

// WSSession represents a single websocket connection.
type WSSession struct {
	URL        string
	Connection ws.Connection
}

// DefaultWSConnectionFactory is used by the client if no other
// ConnectionFactory is provided.
type DefaultWSConnectionFactory struct {
	URL       string
	AuthInfo  *IAMAuthInfo
	TLSConfig *tls.Config
}

func (wcf *DefaultWSConnectionFactory) New() (ext.Conn, error) {
	var (
		dialer websocket.Dialer
		header = http.Header{}
	)

	if wcf.AuthInfo != nil && len(wcf.AuthInfo.IAMToken()) > 0 {
		header.Add(AuthorizationHeader, wcf.AuthInfo.IAMToken())
	}

	if wcf.TLSConfig != nil {
		dialer.TLSClientConfig = wcf.TLSConfig
	}

	conn, resp, err := dialer.Dial(wcf.URL, header)
	if resp != nil && resp.Body != nil {
		// TODO: dump response, which is second return value from Dial
		resp.Body.Close()
	}

	return conn, err
}

func (wcf *DefaultWSConnectionFactory) NewSession(connection ws.Connection) *WSSession {
	return &WSSession{
		URL:        wcf.URL,
		Connection: connection,
	}
}

type WSConnectionOptions struct {
	ws.ConnectionOptions
	Factory WSConnectionFactory
}

// WSClient manages the lifetime of a single websocket connection.
type WSClient struct {
	ConnectionFactory WSConnectionFactory
	ConnectionOptions ws.ConnectionOptions
	session           *WSSession
	errLock           sync.RWMutex
	sessionLock       sync.RWMutex
	err               error
}

func NewWS(opts WSConnectionOptions) *WSClient {
	if opts.Factory == nil {
		opts.Factory = &DefaultWSConnectionFactory{
			URL: "127.0.0.1:8083",
		}
	}

	return &WSClient{
		ConnectionOptions: opts.ConnectionOptions,
		ConnectionFactory: opts.Factory,
	}
}

func (c *WSClient) setErr(err error) {
	c.errLock.Lock()
	defer c.errLock.Unlock()

	c.err = err
}

func (c *WSClient) getErr() error {
	c.errLock.RLock()
	defer c.errLock.RUnlock()

	return c.err
}

// Session provides the web socket session instance
func (c *WSClient) Session() *WSSession {
	c.sessionLock.RLock()
	defer c.sessionLock.RUnlock()

	return c.session
}

// connect is for internal use and should be called within
// the scope of an acquired 'c.sessionLock.Lock()'
//
// extracted for internal re-use.
func (c *WSClient) connect() error {
	conn, err := c.ConnectionFactory.New()
	if err != nil {
		return err
	}

	connection, err := ws.NewConnection(conn, c.ConnectionOptions)
	if err != nil {
		return err
	}

	c.session = c.ConnectionFactory.NewSession(connection)

	go func() {
		// There is a race condition where session is set to nil before
		// Listen is called. This check resolves segfaults during tests,
		// but there's still a gap where session can be nullified before
		// Listen is invoked. The odds of that happening outside of tests
		// is extremely small; e.g., who will call Dis/Reconnect immediately
		// after calling Connect?
		if c.Session() == nil {
			return
		}

		// Starts the async read. If there is a read error, it is set so that
		// it is returned the next time Send is called. That should be
		// sufficient for most cases where the client cares only about sending.
		// If the client really cares about handling reads, they will define a
		// custom ReadHandler that will receive the error synchronously.
		if err := c.session.Connection.Listen(); err != nil {
			c.setErr(err)
		}
	}()

	return nil
}

// Connect initializes the Session and Connection objects by opening
// a websocket connection. If AuthInfo is not nil, the token it returns
// will be passed via the "Authentication" header during the initial
// HTTP call.
func (c *WSClient) Connect() error {
	c.sessionLock.Lock()
	defer c.sessionLock.Unlock()

	if c.session != nil {
		return errors.New("a session is already active")
	}

	return c.connect()
}

// Disconnect ends the current Session and terminates its websocket connection.
func (c *WSClient) Disconnect() (err error) {
	c.sessionLock.Lock()
	defer c.sessionLock.Unlock()

	if c.session != nil && !c.session.Connection.Closed() {
		err = c.session.Connection.Close()
	}

	c.session = nil

	return
}

// Reconnect terminates the existing Session and creates a new one.
func (c *WSClient) Reconnect() (err error) {
	c.sessionLock.Lock()
	defer c.sessionLock.Unlock()

	if c.session != nil && !c.session.Connection.Closed() {
		_ = c.session.Connection.Close()
	}

	if err = c.connect(); err != nil {
		c.session = nil
	}

	c.setErr(err)

	return
}

// Send sends a single msgp.Encodable across the wire.
func (c *WSClient) Send(e protocol.ChunkEncoder) error {
	// Check for an async connection error and return it here.
	// In most cases, the client will not care about reading from
	// the connection, so checking for the error here is sufficient.
	if err := c.getErr(); err != nil {
		return err // TODO: wrap this
	}

	// prevent this from raise conditions by copy the session pointer
	session := c.Session()
	if session == nil || session.Connection.Closed() {
		return errors.New("no active session")
	}

	// msgp.Encode makes use of object pool to decrease allocations
	return msgp.Encode(session.Connection, e)
}

// SendRaw sends an array of bytes across the wire.
func (c *WSClient) SendRaw(m []byte) error {
	// Check for an async connection error and return it here.
	// In most cases, the client will not care about reading from
	// the connection, so checking for the error here is sufficient.
	if err := c.getErr(); err != nil {
		return err // TODO: wrap this
	}

	// prevent this from raise conditions by copy the session pointer
	session := c.Session()
	if session == nil || session.Connection.Closed() {
		return errors.New("no active session")
	}

	_, err := session.Connection.Write(m)

	return err
}
