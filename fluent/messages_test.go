package fluent_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.ibm.com/Observability/fluent-forward-go/fluent"

	"github.com/vmihailenco/msgpack/v5"
)

var _ = Describe("Messages", func() {
	Describe("NewHelo", func() {
		var (
		// opts HeloOpts
		)

		It("Returns a properly-structured, msgpack-encoded HELO message", func() {
			b, _ := NewHelo(nil)
			var h Helo
			msgpack.Unmarshal(b, &h)
			fmt.Fprintf(os.Stderr, "VALUE IS \n%v\n", h)
			fmt.Fprintf(os.Stderr, "ENCODED IS \n%s\n", string(b))
			Expect(h.Options.Keepalive).To(BeTrue())
		})
	})
})
