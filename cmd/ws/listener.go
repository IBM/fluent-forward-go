// MIT License

// Copyright contributors to the fluent-forward-go project

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/IBM/fluent-forward-go/fluent/client/ws"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Listener struct {
	upgrader       *websocket.Upgrader
	server         *http.Server
	shutdown       chan struct{}
	exited         chan struct{}
	connectionLock sync.Mutex
	wsopts         ws.ConnectionOptions
}

func NewListener(server *http.Server, wsopts ws.ConnectionOptions) *Listener {
	return &Listener{
		&websocket.Upgrader{},
		server,
		make(chan struct{}, 1),
		make(chan struct{}, 1),
		sync.Mutex{},
		wsopts,
	}
}

func (s *Listener) Connect(w http.ResponseWriter, r *http.Request) {
	var (
		c   *websocket.Conn
		err error
	)

	if c, err = s.upgrader.Upgrade(w, r, nil); err == nil {
		connection, _ := ws.NewConnection(c, s.wsopts)

		s.server.RegisterOnShutdown(func() {
			if !connection.Closed() {
				if err := connection.Close(); err != nil && err != websocket.ErrCloseSent {
					log.Println("server conn close error:", err)
				}
			}

			log.Println("server conn closed")
		})

		if err = connection.Listen(); err != nil &&
			!websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			log.Println("server listen error:", err)
		}
	}
}

var (
	once   sync.Once
	router *mux.Router
)

func (s *Listener) ListenAndServe() error {
	defer func() { s.exited <- struct{}{} }()

	once.Do(func() {
		// TODO comment explaining this weirdness
		router = mux.NewRouter()
		http.Handle("/", router)
	})

	router.HandleFunc("/", s.Connect)

	if useTLS {
		config := &tls.Config{
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
			PreferServerCipherSuites: true,
			MinVersion:               tls.VersionTLS12,
		}

		s.server.TLSConfig = config

		go func() {
			if err := s.server.ListenAndServeTLS(
				"../../fluent/client/clientfakes/cert.pem", "../../fluent/client/clientfakes/key.pem",
			); err != nil && err != http.ErrServerClosed {
				log.Fatal("ListenAndServe error: " + err.Error())
			}
		}()
	} else {
		go func() {
			if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal("ListenAndServe error: " + err.Error())
			}
		}()
	}

	<-s.shutdown
	log.Println("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var err error
	if err = s.server.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		log.Println("shutdown server error:", err)

		if err = s.server.Close(); err != nil {
			log.Println("close server error:", err)
		}
	}

	if ctx.Err() != nil {
		log.Println("close context error:", ctx.Err())
	}

	s.exited <- struct{}{}

	return err
}

func (s *Listener) Shutdown() {
	s.shutdown <- struct{}{}
	<-s.exited
}
