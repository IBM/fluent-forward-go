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
	URL      string
	AuthInfo *IAMAuthInfo
}

func (wcf *DefaultWSConnectionFactory) New() (ext.Conn, error) {
	var (
		dialer websocket.Dialer
		header = http.Header{}
	)

	if wcf.AuthInfo != nil && len(wcf.AuthInfo.IAMToken()) > 0 {
		header.Add(AuthorizationHeader, wcf.AuthInfo.IAMToken())
	}

	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //#nosec

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
	Factory  WSConnectionFactory
	AuthInfo *IAMAuthInfo
	URL      string
}

// WSClient manages the lifetime of a single websocket connection.
type WSClient struct {
	ConnectionFactory WSConnectionFactory
	URL               string
	AuthInfo          *IAMAuthInfo
	ConnectionOptions ws.ConnectionOptions
	session           *WSSession
	errLock           sync.RWMutex
	sessionLock       sync.RWMutex
	err               error
}

func NewWS(opts WSConnectionOptions) *WSClient {
	if opts.AuthInfo == nil {
		opts.AuthInfo = &IAMAuthInfo{}
	}

	return &WSClient{
		ConnectionOptions: opts.ConnectionOptions,
		ConnectionFactory: opts.Factory,
		URL:               opts.URL,
		AuthInfo:          opts.AuthInfo,
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

// connect is for internal use and should be called within
// the scope of an acquired 'c.sessionLock.Lock()'
//
// extracted for internal re-use.
func (c *WSClient) connect() error {
	if c.ConnectionFactory == nil {
		c.ConnectionFactory = &DefaultWSConnectionFactory{
			URL:      c.URL,
			AuthInfo: c.AuthInfo,
		}
	}

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
		// it is returned the next time SendMessage is called. That should be
		// sufficient for most cases where the client cares only about sending.
		// If the client really cares about handling reads, they will define a
		// custom ReadHandler that will receive the error synchronously.
		if err := c.session.Connection.Listen(); err != nil {
			c.setErr(err)
		}
	}()

	return nil
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

// SendMessage sends a single msgp.Encodable across the wire.
func (c *WSClient) SendMessage(e protocol.ChunkEncoder) error {
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
