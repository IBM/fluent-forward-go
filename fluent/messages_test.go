package fluent_test

import (
	"crypto/sha512"
	"encoding/hex"
	"io"
	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.ibm.com/Observability/fluent-forward-go/fluent"

	"github.com/vmihailenco/msgpack/v5"
)

var _ = Describe("Messages", func() {
	var (
		hostname, username, password string
		salt, nonce, key             []byte
		authResult                   bool
		reason                       string
	)

	BeforeEach(func() {
		hostname = "client.host.domain"
		username = "someuser"
		password = "xyz123"

		salt = make([]byte, 16)
		rand.Read(salt)

		nonce = make([]byte, 16)
		rand.Read(nonce)

		key = []byte(`thisisasharedkeyandmaybeitshouldbereallylong?`)

		authResult = true
	})

	digest := func(salt []byte, hostname string, nonce, sharedKey []byte) []byte {
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

	Describe("PackedHelo", func() {
		var (
		// opts HeloOpts
		)

		It("Returns a properly-structured, msgpack-encoded HELO message", func() {
			b, _ := PackedHelo(nil)
			var h Helo
			msgpack.Unmarshal(b, &h)
			Expect(h.Options.Keepalive).To(BeTrue())
		})
	})

	Describe("PackedPing", func() {
		It("Returns a correctly structured and msgpack-encoded PING message", func() {
			b, _ := PackedPing(hostname, key, salt, nonce, username, password)
			var p Ping
			msgpack.Unmarshal(b, &p)
			Expect(p.MessageType).To(Equal(MSGTYPE_PING))
			Expect(p.ClientHostname).To(Equal(hostname))
			Expect(p.SharedKeySalt).To(Equal(salt))
			Expect(p.SharedKeyHexDigest).To(Equal(digest(salt, hostname, nonce, key)))
			Expect(p.Username).To(Equal(username))
			Expect(p.Password).To(Equal(password))
		})
	})

	Describe("PackedPong", func() {
		It("Returns a correctly structured and msgpack-encoded PONG message", func() {
			b, _ := PackedPong(authResult, reason, hostname, key, salt, nonce)
			var p Pong
			msgpack.Unmarshal(b, &p)
			Expect(p.MessageType).To(Equal(MSGTYPE_PONG))
			Expect(p.AuthResult).To(Equal(authResult))
			Expect(p.Reason).To(Equal(reason))
			Expect(p.ServerHostname).To(Equal(hostname))
			Expect(p.SharedKeyHexDigest).To(Equal(digest(salt, hostname, nonce, key)))
		})
	})

	Describe("ValidatePingDigest", func() {
		var (
			ping                               *Ping
			digestHostname                     string
			digestSalt, digestNonce, digestKey []byte
		)

		BeforeEach(func() {
			digestHostname = hostname
			digestSalt = salt
			digestNonce = nonce
			digestKey = key
		})

		JustBeforeEach(func() {
			ping = &Ping{
				MessageType:        MSGTYPE_PING,
				ClientHostname:     hostname,
				SharedKeySalt:      salt,
				SharedKeyHexDigest: digest(digestSalt, digestHostname, digestNonce, digestKey),
				Username:           username,
				Password:           password,
			}
		})

		It("Does not return an error", func() {
			Expect(ValidatePingDigest(ping, key, nonce)).NotTo(HaveOccurred())
		})

		Context("When the hostname does not match the digest", func() {
			BeforeEach(func() {
				digestHostname = "not.the.correct.hostname"
			})

			It("Returns an error", func() {
				Expect(ValidatePingDigest(ping, key, nonce)).To(HaveOccurred())
			})
		})

		Context("When the salt does not match the digest", func() {
			BeforeEach(func() {
				digestSalt = []byte(`someothersalt`)
			})

			It("Returns an error", func() {
				Expect(ValidatePingDigest(ping, key, nonce)).To(HaveOccurred())
			})
		})

		Context("When the nonce does not match the digest", func() {
			BeforeEach(func() {
				digestNonce = []byte(`someothernonce`)
			})

			It("Returns an error", func() {
				Expect(ValidatePingDigest(ping, key, nonce)).To(HaveOccurred())
			})
		})

		Context("When the key does not match the digest", func() {
			BeforeEach(func() {
				digestKey = []byte(`someotherkey`)
			})

			It("Returns an error", func() {
				Expect(ValidatePingDigest(ping, key, nonce)).To(HaveOccurred())
			})
		})
	})

	Describe("ValidatePongDigest", func() {
		var (
			pong                               *Pong
			digestHostname                     string
			digestSalt, digestNonce, digestKey []byte
		)

		BeforeEach(func() {
			digestHostname = hostname
			digestSalt = salt
			digestNonce = nonce
			digestKey = key
		})

		JustBeforeEach(func() {
			pong = &Pong{
				MessageType:        MSGTYPE_PONG,
				AuthResult:         authResult,
				Reason:             reason,
				ServerHostname:     hostname,
				SharedKeyHexDigest: digest(digestSalt, digestHostname, digestNonce, digestKey),
			}
		})

		It("Does not return an error", func() {
			Expect(ValidatePongDigest(pong, key, nonce, salt)).NotTo(HaveOccurred())
		})

		Context("When the hostname does not match the digest", func() {
			BeforeEach(func() {
				digestHostname = "not.the.correct.hostname"
			})

			It("Returns an error", func() {
				Expect(ValidatePongDigest(pong, key, nonce, salt)).To(HaveOccurred())
			})
		})

		Context("When the salt does not match the digest", func() {
			BeforeEach(func() {
				digestSalt = []byte(`someothersalt`)
			})

			It("Returns an error", func() {
				Expect(ValidatePongDigest(pong, key, nonce, salt)).To(HaveOccurred())
			})
		})

		Context("When the nonce does not match the digest", func() {
			BeforeEach(func() {
				digestNonce = []byte(`someothernonce`)
			})

			It("Returns an error", func() {
				Expect(ValidatePongDigest(pong, key, nonce, salt)).To(HaveOccurred())
			})
		})

		Context("When the key does not match the digest", func() {
			BeforeEach(func() {
				digestKey = []byte(`someotherkey`)
			})

			It("Returns an error", func() {
				Expect(ValidatePongDigest(pong, key, nonce, salt)).To(HaveOccurred())
			})
		})
	})
})
