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

package ext

import (
	"io"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// Conn is an interface for websocket.Conn. Generated with
// ifacemaker (https://github.com/vburenin/ifacemaker).
//
//counterfeiter:generate . Conn
type Conn interface {
	// Subprotocol returns the negotiated protocol for the connection.
	Subprotocol() string
	// Close closes the underlying network connection without sending or waiting
	// for a close message.
	Close() error
	// LocalAddr returns the local network address.
	LocalAddr() net.Addr
	// RemoteAddr returns the remote network address.
	RemoteAddr() net.Addr
	// WriteControl writes a control message with the given deadline. The allowed
	// message types are CloseMessage, PingMessage and PongMessage.
	WriteControl(messageType int, data []byte, deadline time.Time) error
	// NextWriter returns a writer for the next message to send. The writer's Close
	// method flushes the complete message to the network.
	//
	// There can be at most one open writer on a connection. NextWriter closes the
	// previous writer if the application has not already done so.
	//
	// All message types (TextMessage, BinaryMessage, CloseMessage, PingMessage and
	// PongMessage) are supported.
	NextWriter(messageType int) (io.WriteCloser, error)
	// WritePreparedMessage writes prepared message into connection.
	WritePreparedMessage(pm *websocket.PreparedMessage) error
	// WriteMessage is a helper method for getting a writer using NextWriter,
	// writing the message and closing the writer.
	WriteMessage(messageType int, data []byte) error
	// SetWriteDeadline sets the write deadline on the underlying network
	// connection. After a write has timed out, the websocket state is corrupt and
	// all future writes will return an error. A zero value for t means writes will
	// not time out.
	SetWriteDeadline(t time.Time) error
	// NextReader returns the next data message received from the peer. The
	// returned messageType is either TextMessage or BinaryMessage.
	//
	// There can be at most one open reader on a connection. NextReader discards
	// the previous message if the application has not already consumed it.
	//
	// Applications must break out of the application's read loop when this method
	// returns a non-nil error value. Errors returned from this method are
	// permanent. Once this method returns a non-nil error, all subsequent calls to
	// this method return the same error.
	NextReader() (messageType int, r io.Reader, err error)
	// ReadMessage is a helper method for getting a reader using NextReader and
	// reading from that reader to a buffer.
	ReadMessage() (messageType int, p []byte, err error)
	// SetReadDeadline sets the read deadline on the underlying network connection.
	// After a read has timed out, the websocket connection state is corrupt and
	// all future reads will return an error. A zero value for t means reads will
	// not time out.
	SetReadDeadline(t time.Time) error
	// SetReadLimit sets the maximum size in bytes for a message read from the peer. If a
	// message exceeds the limit, the connection sends a close message to the peer
	// and returns ErrReadLimit to the application.
	SetReadLimit(limit int64)
	// CloseHandler returns the current close handler
	CloseHandler() func(code int, text string) error
	// SetCloseHandler sets the handler for close messages received from the peer.
	// The code argument to h is the received close code or CloseNoStatusReceived
	// if the close message is empty. The default close handler sends a close
	// message back to the peer.
	//
	// The handler function is called from the NextReader, ReadMessage and message
	// reader Read methods. The application must read the connection to process
	// close messages as described in the section on Control Messages above.
	//
	// The connection read methods return a CloseError when a close message is
	// received. Most applications should handle close messages as part of their
	// normal error handling. Applications should only set a close handler when the
	// application must perform some action before sending a close message back to
	// the peer.
	SetCloseHandler(h func(code int, text string) error)
	// PingHandler returns the current ping handler
	PingHandler() func(appData string) error
	// SetPingHandler sets the handler for ping messages received from the peer.
	// The appData argument to h is the PING message application data. The default
	// ping handler sends a pong to the peer.
	//
	// The handler function is called from the NextReader, ReadMessage and message
	// reader Read methods. The application must read the connection to process
	// ping messages as described in the section on Control Messages above.
	SetPingHandler(h func(appData string) error)
	// PongHandler returns the current pong handler
	PongHandler() func(appData string) error
	// SetPongHandler sets the handler for pong messages received from the peer.
	// The appData argument to h is the PONG message application data. The default
	// pong handler does nothing.
	//
	// The handler function is called from the NextReader, ReadMessage and message
	// reader Read methods. The application must read the connection to process
	// pong messages as described in the section on Control Messages above.
	SetPongHandler(h func(appData string) error)
	// UnderlyingConn returns the internal net.Conn. This can be used to further
	// modifications to connection specific flags.
	UnderlyingConn() net.Conn
	// EnableWriteCompression enables and disables write compression of
	// subsequent text and binary messages. This function is a noop if
	// compression was not negotiated with the peer.
	EnableWriteCompression(enable bool)
	// SetCompressionLevel sets the flate compression level for subsequent text and
	// binary messages. This function is a noop if compression was not negotiated
	// with the peer. See the compress/flate package for a description of
	// compression levels.
	SetCompressionLevel(level int) error
}
