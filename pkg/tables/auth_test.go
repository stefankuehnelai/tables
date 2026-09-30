package tables_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

var _ = Describe("Authentication", func() {
	It("applies basic authentication", func() {
		By("Arrange")
		request, err := http.NewRequest(http.MethodGet, "https://cloud.example.net", nil)
		Expect(err).NotTo(HaveOccurred())
		auth := tables.BasicAuth{Username: "alice", Password: "secret"}

		By("Act")
		err = auth.Apply(request)

		By("Assert")
		Expect(err).NotTo(HaveOccurred())
		username, password, ok := request.BasicAuth()
		Expect(ok).To(BeTrue())
		Expect(username).To(Equal("alice"))
		Expect(password).To(Equal("secret"))
	})
})
