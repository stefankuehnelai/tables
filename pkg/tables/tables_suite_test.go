package tables

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTables(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tables Suite")
}
