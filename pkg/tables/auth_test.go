package tables

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BasicAuth", func() {
	It("applies HTTP basic authentication", func() {
		By("Arrange")
		req, err := http.NewRequest(http.MethodGet, "https://example.test", nil)
		Expect(err).NotTo(HaveOccurred())

		By("Act")
		err = (BasicAuth{Username: "alice", Password: "secret"}).Apply(req)

		By("Assert")
		Expect(err).NotTo(HaveOccurred())
		username, password, ok := req.BasicAuth()
		Expect(ok).To(BeTrue())
		Expect(username).To(Equal("alice"))
		Expect(password).To(Equal("secret"))
	})
})
