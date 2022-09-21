package ws_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ws Suite")
}
