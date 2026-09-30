package cli

import (
	"bytes"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type memoryConfigStore struct {
	value config
	err   error
}

func (s *memoryConfigStore) Load() (config, error)   { return s.value, s.err }
func (s *memoryConfigStore) Save(value config) error { s.value = value; return s.err }

type memoryCredentialStore struct {
	values map[string]string
	err    error
}

func (s *memoryCredentialStore) Set(hostname, username, kind, secret string) error {
	if s.err != nil {
		return s.err
	}
	s.values[credentialAccount(hostname, username, kind)] = secret
	return nil
}
func (s *memoryCredentialStore) Get(hostname, username, kind string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	value, ok := s.values[credentialAccount(hostname, username, kind)]
	if !ok {
		return "", errCredentialNotFound
	}
	return value, nil
}
func (s *memoryCredentialStore) Delete(hostname, username, kind string) error {
	if s.err != nil {
		return s.err
	}
	delete(s.values, credentialAccount(hostname, username, kind))
	return nil
}

func testDependencies() (dependencies, *memoryConfigStore, *memoryCredentialStore, *bytes.Buffer) {
	cfg := &memoryConfigStore{value: newConfig()}
	creds := &memoryCredentialStore{values: map[string]string{}}
	output := &bytes.Buffer{}
	env := map[string]string{}
	return dependencies{
		config: cfg, credentials: creds, stdin: bytes.NewBufferString(""), stdout: output, stderr: &bytes.Buffer{},
		getenv:     func(key string) string { return env[key] },
		readSecret: func(string) (string, error) { return "prompt-secret", nil },
	}, cfg, creds, output
}

var _ = Describe("root command", func() {
	It("resolves stored and environment connections in precedence order", func() {
		deps, cfg, creds, _ := testDependencies()
		cfg.value = config{ActiveHostname: "stored.example", Hosts: map[string]hostConfig{
			"stored.example": {Server: "https://stored.example/nc", ActiveUsername: "stored", Accounts: map[string]accountConfig{"stored": {CredentialKind: "password"}}},
		}}
		creds.values[credentialAccount("stored.example", "stored", "password")] = "stored-secret"
		env := map[string]string{envHostname: "stored.example", envUsername: "stored", envAppPassword: "env-secret"}
		deps.getenv = func(key string) string { return env[key] }
		app := &application{deps: deps}

		connection, err := app.resolveConnection()
		Expect(err).NotTo(HaveOccurred())
		Expect(connection.server).To(Equal("https://stored.example/nc"))
		Expect(connection.kind).To(Equal("app-password"))
		Expect(connection.secret).To(Equal("env-secret"))
	})

	It("reports missing connection configuration", func() {
		deps, _, _, _ := testDependencies()
		app := &application{deps: deps}
		_, err := app.resolveConnection()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("auth login"))
	})

	It("validates server URLs", func() {
		Expect(validateServer("https://example.net/nextcloud")).To(Succeed())
		Expect(validateServer("not-a-url")).To(HaveOccurred())
	})

	It("selects the first non-empty value", func() {
		Expect(firstNonEmpty(" ", "value", "other")).To(Equal("value"))
		Expect(firstNonEmpty("", " ")).To(BeEmpty())
	})

	It("creates the production command", func() {
		old := fmt.Sprintf("%s", GinkgoT().TempDir())
		DeferCleanup(func() {})
		_ = old
		command, err := NewRootCommand()
		Expect(err).NotTo(HaveOccurred())
		Expect(command.Use).To(Equal("tables"))
	})
})
