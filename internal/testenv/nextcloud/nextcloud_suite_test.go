package nextcloud

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNextcloudTestEnvironment(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Nextcloud Test Environment Suite")
}
