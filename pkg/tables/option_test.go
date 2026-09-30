package tables

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client options", func() {
	It("configures credentials and app passwords", func() {
		client, err := NewClient("https://example.test", WithCredentials("alice", "pw"))
		Expect(err).NotTo(HaveOccurred())
		Expect(client.auth).To(Equal(BasicAuth{Username: "alice", Password: "pw"}))

		client, err = NewClient("https://example.test", WithAppPassword("alice", "app"))
		Expect(err).NotTo(HaveOccurred())
		Expect(client.auth).To(Equal(BasicAuth{Username: "alice", Password: "app"}))
	})

	It("rejects invalid credentials", func() {
		_, err := NewClient("https://example.test", WithCredentials("", "pw"))
		Expect(IsCode(err, CodeInvalidConfiguration)).To(BeTrue())
		_, err = NewClient("https://example.test", WithCredentials("alice", ""))
		Expect(IsCode(err, CodeInvalidConfiguration)).To(BeTrue())
	})

	It("configures a custom HTTP client", func() {
		httpClient := &http.Client{}
		client, err := NewClient("https://example.test", WithHTTPClient(httpClient))
		Expect(err).NotTo(HaveOccurred())
		Expect(client.httpClient).To(BeIdenticalTo(httpClient))

		_, err = NewClient("https://example.test", WithHTTPClient(nil))
		Expect(IsCode(err, CodeInvalidConfiguration)).To(BeTrue())
	})
})
