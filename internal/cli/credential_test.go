package cli

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/zalando/go-keyring"
)

var _ = Describe("keyring credential store", func() {
	BeforeEach(func() { keyring.MockInit() })

	It("stores, reads, and deletes credentials", func() {
		store := keyringCredentialStore{}
		Expect(store.Set("cloud.example.net", "alice", "app-password", "secret")).To(Succeed())
		value, err := store.Get("cloud.example.net", "alice", "app-password")
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(Equal("secret"))
		Expect(store.Delete("cloud.example.net", "alice", "app-password")).To(Succeed())
		_, err = store.Get("cloud.example.net", "alice", "app-password")
		Expect(err).To(MatchError(errCredentialNotFound))
		Expect(store.Delete("cloud.example.net", "alice", "app-password")).To(Succeed())
	})
})
