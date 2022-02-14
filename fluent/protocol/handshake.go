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

package protocol

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"io"
)

//go:generate msgp

// =========
// HANDSHAKE
// =========

const (
	MsgTypeHelo = "HELO"
	MsgTypePing = "PING"
	MsgTypePong = "PONG"
)

// Remember that the handshake flow is like this:
// 1. Server -> HELO -> Client
// 2. Client -> PING -> Server
// 3. Server -> PONG -> Client

// NewHelo returns a Helo message with the specified options.
// if opts is nil, then a nonce is generated, auth is left empty, and
// keepalive is true.
func NewHelo(opts *HeloOpts) *Helo {
	h := Helo{MessageType: MsgTypeHelo}
	if opts == nil {
		h.Options = &HeloOpts{
			Keepalive: true,
		}
	} else {
		h.Options = opts
	}

	return &h
}

// Helo is the initial handshake message, sent by the server and received
// by the client.  Client will respond with a Ping.
//msgp:tuple Helo
type Helo struct {
	MessageType string
	Options     *HeloOpts
}

type HeloOpts struct {
	Nonce     []byte `msg:"nonce"`
	Auth      []byte `msg:"auth"`
	Keepalive bool   `msg:"keepalive"`
}

// NewPing returns a PING message.  The digest is computed
// from the hostname, key, salt, and nonce using SHA512.
func NewPing(hostname string, sharedKey, salt, nonce []byte) (*Ping, error) {
	return makePing(hostname, sharedKey, salt, nonce)
}

// NewPingWithAuth returns a PING message containing the username and password
// to be used for authentication.  The digest is computed
// from the hostname, key, salt, and nonce using SHA512.
func NewPingWithAuth(hostname string, sharedKey, salt, nonce []byte, username, password string) (*Ping, error) {
	return makePing(hostname, sharedKey, salt, nonce, username, password)
}

func makePing(hostname string, sharedKey, salt, nonce []byte, creds ...string) (*Ping, error) {
	bytes, err := computeHexDigest(salt, hostname, nonce, sharedKey)

	p := Ping{
		MessageType:        MsgTypePing,
		ClientHostname:     hostname,
		SharedKeySalt:      salt,
		SharedKeyHexDigest: bytes,
	}

	if len(creds) >= 2 {
		p.Username = creds[0]
		p.Password = creds[1]
	}

	return &p, err
}

// Ping is the response message sent by the client after receiving a
// Helo from the server.  Server will respond with a Pong.
//msgp:tuple Ping
type Ping struct {
	MessageType        string
	ClientHostname     string
	SharedKeySalt      []byte
	SharedKeyHexDigest []byte
	Username           string
	Password           string
}

// NewPong returns a PONG message.  AuthResult indicates
// whether the credentials presented by the client were accepted and therefore
// whether the client can continue using the connection, switching from
// handshake mode to sending events.
// As with the PING, the digest is computed
// from the hostname, key, salt, and nonce using SHA512.
// Server implementations must use the nonce created for the initial Helo and
// the salt sent by the client in the Ping.
func NewPong(authResult bool, reason string, hostname string, sharedKey []byte,
	helo *Helo, ping *Ping) (*Pong, error) {
	if helo == nil || ping == nil {
		return nil, errors.New("Either helo or ping is nil")
	}

	if helo.Options == nil {
		return nil, errors.New("Helo has a nil options field")
	}

	bytes, err := computeHexDigest(ping.SharedKeySalt, hostname, helo.Options.Nonce, sharedKey)

	p := Pong{
		MessageType:        MsgTypePong,
		AuthResult:         authResult,
		Reason:             reason,
		ServerHostname:     hostname,
		SharedKeyHexDigest: bytes,
	}

	return &p, err
}

// Pong is the response message sent by the server after receiving a
// Ping from the client.  A Pong concludes the handshake.
//msgp:tuple Pong
type Pong struct {
	MessageType        string
	AuthResult         bool
	Reason             string
	ServerHostname     string
	SharedKeyHexDigest []byte
}

// ValidatePingDigest validates that the digest contained in the PING message
// is valid for the client hostname (as contained in the PING).
// Returns a non-nil error if validation fails, nil otherwise.
func ValidatePingDigest(p *Ping, key, nonce []byte) error {
	return validateDigest(p.SharedKeyHexDigest, key, nonce, p.SharedKeySalt, p.ClientHostname)
}

// ValidatePongDigest validates that the digest contained in the PONG message
// is valid for the server hostname (as contained in the PONG).
// Returns a non-nil error if validation fails, nil otherwise.
func ValidatePongDigest(p *Pong, key, nonce, salt []byte) error {
	return validateDigest(p.SharedKeyHexDigest, key, nonce, salt, p.ServerHostname)
}

func validateDigest(received, key, nonce, salt []byte, hostname string) error {
	expected, err := computeHexDigest(salt, hostname, nonce, key)
	if err != nil {
		return err
	}

	if !bytes.Equal(received, expected) {
		return errors.New("No match")
	}

	return nil
}

func computeHexDigest(salt []byte, hostname string, nonce, sharedKey []byte) ([]byte, error) {
	h := sha512.New()
	h.Write(salt)

	_, err := io.WriteString(h, hostname)
	if err != nil {
		return nil, err
	}

	h.Write(nonce)
	h.Write(sharedKey)
	sum := h.Sum(nil)
	hexOut := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(hexOut, sum)

	return hexOut, err
}
