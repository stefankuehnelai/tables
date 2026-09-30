package cli

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("configuration", func() {
	It("round trips non-secret account data with restrictive permissions", func() {
		By("Arrange")
		path := filepath.Join(GinkgoT().TempDir(), "nested", "config.json")
		store := newFileConfigStore(path)
		expected := config{ActiveHostname: "cloud.example.net", Hosts: map[string]hostConfig{
			"cloud.example.net": {Server: "https://cloud.example.net/nextcloud", ActiveUsername: "alice", Accounts: map[string]accountConfig{"alice": {CredentialKind: "app-password"}}},
		}}

		By("Act")
		Expect(store.Save(expected)).To(Succeed())
		actual, err := store.Load()

		By("Assert")
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).To(Equal(expected))
		info, err := os.Stat(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(info.Mode().Perm()).To(Equal(os.FileMode(0o600)))
	})

	It("returns an empty config for a missing file", func() {
		store := newFileConfigStore(filepath.Join(GinkgoT().TempDir(), "missing.json"))
		value, err := store.Load()
		Expect(err).NotTo(HaveOccurred())
		Expect(value.Hosts).To(BeEmpty())
	})

	It("reports malformed configuration", func() {
		path := filepath.Join(GinkgoT().TempDir(), "config.json")
		Expect(os.WriteFile(path, []byte("{"), 0o600)).To(Succeed())
		_, err := newFileConfigStore(path).Load()
		Expect(err).To(MatchError(ContainSubstring("decode config")))
	})
})
