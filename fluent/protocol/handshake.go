package protocol

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"io"

	"github.com/vmihailenco/msgpack/v5"
)

// =========
// HANDSHAKE
// =========

const (
	MSGTYPE_HELO = "HELO"
	MSGTYPE_PING = "PING"
	MSGTYPE_PONG = "PONG"
)

// Remember that the handshake flow is like this:
// 1. Server -> HELO -> Client
// 2. Client -> PING -> Server
// 3. Server -> PONG -> Client

// PackedHelo returns a msgpack-encoded Helo message with the specified options.
// if opts is nil, then a nonce is generated, auth is left empty, and
// keepalive is true.
func PackedHelo(opts *HeloOpts) ([]byte, error) {
	h := Helo{MessageType: MSGTYPE_HELO}
	if opts == nil {
		h.Options = &HeloOpts{
			Keepalive: true,
		}
	}
	return msgpack.Marshal(h)
}

// Helo is the initial handshake message, sent by the server and received
// by the client.  Client will respond with a Ping.
type Helo struct {
	_msgpack    struct{} `msgpack:",as_array"`
	MessageType string
	Options     *HeloOpts
}

type HeloOpts struct {
	Nonce     []byte `msgpack:"nonce"`
	Auth      []byte `msgpack:"auth"`
	Keepalive bool   `msgpack:"keepalive"`
}

// PackedPing returns a msgpack-encoded PING message.  The digest is computed
// from the hostname, key, salt, and nonce using SHA512.
func PackedPing(hostname string, sharedKey, salt, nonce []byte, username, password string) ([]byte, error) {
	p := Ping{
		MessageType:        MSGTYPE_PING,
		ClientHostname:     hostname,
		SharedKeySalt:      salt,
		SharedKeyHexDigest: computeHexDigest(salt, hostname, nonce, sharedKey),
		Username:           username,
		Password:           password,
	}
	return msgpack.Marshal(p)
}

// Ping is the response message sent by the client after receiving a
// Helo from the server.  Server will respond with a Pong.
type Ping struct {
	_msgpack           struct{} `msgpack:",as_array"`
	MessageType        string
	ClientHostname     string
	SharedKeySalt      []byte
	SharedKeyHexDigest []byte
	Username           string
	Password           string
}

// PackedPong returns a msgpack-encoded PONG message.  AuthResult indicates
// whether the credentials presented by the client were accepted and therefore
// whether the client can continue using the connection, switching from
// handshake mode to sending events.
// As with the PING, the digest is computed
// from the hostname, key, salt, and nonce using SHA512.
// Server implementations must use the nonce created for the initial Helo and
// the salt sent by the client in the Ping.
func PackedPong(authResult bool, reason string, hostname string, sharedKey, salt, nonce []byte) ([]byte, error) {
	p := Pong{
		MessageType:        MSGTYPE_PONG,
		AuthResult:         authResult,
		Reason:             reason,
		ServerHostname:     hostname,
		SharedKeyHexDigest: computeHexDigest(salt, hostname, nonce, sharedKey),
	}
	return msgpack.Marshal(p)
}

// Pong is the response message sent by the server after receiving a
// Ping from the client.  A Pong concludes the handshake.
type Pong struct {
	_msgpack           struct{} `msgpack:",as_array"`
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
	expected := computeHexDigest(salt, hostname, nonce, key)
	if bytes.Compare(received, expected) != 0 {
		return errors.New("No match")
	}
	return nil
}

func computeHexDigest(salt []byte, hostname string, nonce, sharedKey []byte) []byte {
	h := sha512.New()
	h.Write(salt)
	io.WriteString(h, hostname)
	h.Write(nonce)
	h.Write(sharedKey)
	sum := h.Sum(nil)
	hexOut := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(hexOut, sum)
	return hexOut
}
