package ws

import (
	"io"
	"log"
	"runtime/debug"
	"sync"
	"time"

	ext "github.com/IBM/fluent-forward-go/fluent/client/ws/ext"
	"github.com/gorilla/websocket"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type ReadHandler func(conn Connection, messageType int, p []byte, err error) error

type ConnectionOptions struct {
	CloseHandler func(conn Connection, code int, text string) error
	PingHandler  func(conn Connection, appData string) error
	PongHandler  func(conn Connection, appData string) error
	ReadDeadline time.Time
	// ReadHandler handles new messages received on the websocket. If the handler receives
	// an error, the client MUST close the connection and return an error.
	ReadHandler ReadHandler
	// A zero value for means writes will not time out.
	WriteDeadline time.Time
}

//counterfeiter:generate . Connection
type Connection interface {
	ext.Conn
	CloseWithMsg(closeCode int, msg string) error
	Closed() bool
	Listen() error
	ReadHandler() ReadHandler
	SetReadHandler(rh ReadHandler)
	Write(data []byte) (int, error)
}

type connection struct {
	ext.Conn
	writeLock   sync.Mutex
	closedLock  sync.RWMutex
	readHandler ReadHandler
	connClosed  bool
}

func NewConnection(conn ext.Conn, opts ConnectionOptions) (Connection, error) {
	wsc := &connection{
		Conn: conn,
	}

	if opts.CloseHandler != nil {
		wsc.SetCloseHandler(func(code int, text string) error {
			return opts.CloseHandler(wsc, code, text)
		})
	}

	if opts.PingHandler != nil {
		wsc.SetPingHandler(func(appData string) error {
			return opts.PingHandler(wsc, appData)
		})
	}

	if opts.PongHandler != nil {
		wsc.SetPongHandler(func(appData string) error {
			return opts.PongHandler(wsc, appData)
		})
	}

	if opts.ReadHandler == nil {
		opts.ReadHandler = func(c Connection, _ int, _ []byte, err error) error {
			if err != nil {
				_ = c.Close()
			}

			return err
		}
	}

	wsc.SetReadHandler(opts.ReadHandler)

	if err := wsc.SetReadDeadline(opts.ReadDeadline); err != nil {
		return nil, err
	}

	if err := wsc.SetWriteDeadline(opts.WriteDeadline); err != nil {
		return nil, err
	}

	return wsc, nil
}

// CloseConn return true
func (wsc *connection) CloseWithMsg(closeCode int, msg string) error {
	wsc.closedLock.Lock()
	defer wsc.closedLock.Unlock()

	if wsc.connClosed {
		return nil
	}

	wsc.connClosed = true

	err := wsc.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(
			closeCode, msg,
		),
	)

	if err != nil {
		log.Println("write close failed", err)
		return wsc.Conn.Close()
	}

	return nil
}

func (wsc *connection) Close() error {
	return wsc.CloseWithMsg(websocket.CloseNormalClosure, "so long and thanks for all the fish")
}

func (wsc *connection) Closed() bool {
	wsc.closedLock.RLock()
	defer wsc.closedLock.RUnlock()

	return wsc.connClosed
}

func (wsc *connection) Listen() error {
	// TODO prevent call if already listening
	errChan := make(chan error, 1)

	go func() {
		defer func() {
			log.Println("Connection exiting read loop")

			if r := recover(); r != nil {
				debug.PrintStack()
				log.Println("panic:", r)
			}
		}()

		defer func() { errChan <- nil }()

		for {
			mt, message, err := wsc.Conn.ReadMessage()
			if err = wsc.readHandler(wsc, mt, message, err); err != nil {
				log.Println("readhandler error:", err)

				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
					errChan <- err
					return
				}

				return
			}
		}
	}()

	return <-errChan
}

func (wsc *connection) NextReader() (messageType int, r io.Reader, err error) {
	panic("use ReadHandler instead")
}

func (wsc *connection) ReadMessage() (messageType int, p []byte, err error) {
	panic("use ReadHandler instead")
}

func (wsc *connection) SetReadHandler(rh ReadHandler) {
	wsc.readHandler = rh
}

func (wsc *connection) ReadHandler() ReadHandler {
	return wsc.readHandler
}

func (wsc *connection) WriteMessage(messageType int, data []byte) error {
	wsc.writeLock.Lock()
	defer wsc.writeLock.Unlock()

	return wsc.Conn.WriteMessage(messageType, data)
}

func (wsc *connection) Write(data []byte) (int, error) {
	if err := wsc.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return 0, err
	}

	return len(data), nil
}
