package tables_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

var _ = Describe("Client options", func() {
	It("rejects empty credentials", func() {
		_, err := tables.NewClient("https://cloud.example.net", tables.WithCredentials("", "secret"))
		Expect(err).To(MatchError(ContainSubstring("username and password")))
	})

	It("accepts credentials and app passwords", func() {
		_, err := tables.NewClient("https://cloud.example.net", tables.WithCredentials("alice", "secret"))
		Expect(err).NotTo(HaveOccurred())
		_, err = tables.NewClient("https://cloud.example.net", tables.WithAppPassword("alice", "app-secret"))
		Expect(err).NotTo(HaveOccurred())
	})

	It("accepts a custom HTTP client", func() {
		_, err := tables.NewClient("https://cloud.example.net", tables.WithHTTPClient(&http.Client{}))
		Expect(err).NotTo(HaveOccurred())
	})

	It("rejects a nil HTTP client and nil option", func() {
		_, err := tables.NewClient("https://cloud.example.net", tables.WithHTTPClient(nil))
		Expect(err).To(HaveOccurred())
		_, err = tables.NewClient("https://cloud.example.net", nil)
		Expect(err).To(HaveOccurred())
	})
})
