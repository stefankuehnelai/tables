package nextcloud

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("test environment options", func() {
	It("uses deterministic pinned defaults", func() {
		value := defaultSettings()
		Expect(value.image).To(Equal("nextcloud:35.0.1-apache"))
		Expect(value.tablesVersion).To(Equal("2.3.1"))
	})

	It("applies explicit overrides", func() {
		value := defaultSettings()
		for _, option := range []Option{
			WithImage("nextcloud:35"), WithTablesVersion("2.3.1"),
			WithAdminCredentials("root", "secret"), WithStartupTimeout(time.Minute),
		} {
			option(&value)
		}
		Expect(value.image).To(Equal("nextcloud:35"))
		Expect(value.adminUsername).To(Equal("root"))
		Expect(value.adminPassword).To(Equal("secret"))
		Expect(value.startupTimeout).To(Equal(time.Minute))
	})
})
