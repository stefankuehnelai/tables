package cli

import (
	"bytes"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("auth commands", func() {
	It("logs in from stdin, switches accounts, shows status, and logs out", func() {
		By("Arrange")
		deps, cfg, creds, output := testDependencies()
		deps.stdin = bytes.NewBufferString("secret-one\n")
		command := newRootCommand(deps)

		By("Act")
		command.SetArgs([]string{"auth", "login", "--hostname", "cloud.example.net", "--server", "https://cloud.example.net/nextcloud", "--username", "alice", "--password-stdin"})
		Expect(command.Execute()).To(Succeed())

		By("Assert")
		Expect(cfg.value.ActiveHostname).To(Equal("cloud.example.net"))
		Expect(cfg.value.Hosts["cloud.example.net"].ActiveUsername).To(Equal("alice"))
		Expect(creds.values[credentialAccount("cloud.example.net", "alice", "password")]).To(Equal("secret-one"))

		By("Arrange a second user")
		deps.stdin = bytes.NewBufferString("secret-two\n")
		command = newRootCommand(deps)
		command.SetArgs([]string{"auth", "login", "--hostname", "cloud.example.net", "--username", "bob", "--app-password-stdin"})
		Expect(command.Execute()).To(Succeed())

		By("Switch")
		command = newRootCommand(deps)
		command.SetArgs([]string{"auth", "switch", "--hostname", "cloud.example.net", "--username", "alice"})
		Expect(command.Execute()).To(Succeed())
		Expect(cfg.value.Hosts["cloud.example.net"].ActiveUsername).To(Equal("alice"))

		By("Show status")
		output.Reset()
		command = newRootCommand(deps)
		command.SetArgs([]string{"auth", "status"})
		Expect(command.Execute()).To(Succeed())
		Expect(output.String()).To(ContainSubstring("alice"))
		Expect(output.String()).To(ContainSubstring("bob"))

		By("Logout")
		command = newRootCommand(deps)
		command.SetArgs([]string{"auth", "logout", "--hostname", "cloud.example.net", "--username", "alice"})
		Expect(command.Execute()).To(Succeed())
		_, exists := cfg.value.Hosts["cloud.example.net"].Accounts["alice"]
		Expect(exists).To(BeFalse())
	})

	It("uses environment secrets before prompting", func() {
		deps, _, creds, _ := testDependencies()
		deps.getenv = func(key string) string {
			values := map[string]string{envAppPassword: "from-env"}
			return values[key]
		}
		command := newRootCommand(deps)
		command.SetArgs([]string{"auth", "login", "--hostname", "cloud.example.net", "--username", "alice"})
		Expect(command.Execute()).To(Succeed())
		Expect(creds.values[credentialAccount("cloud.example.net", "alice", "app-password")]).To(Equal("from-env"))
	})

	It("rejects conflicting stdin flags", func() {
		deps, _, _, _ := testDependencies()
		command := newRootCommand(deps)
		command.SetArgs([]string{"auth", "login", "--hostname", "cloud.example.net", "--username", "alice", "--password-stdin", "--app-password-stdin"})
		Expect(command.Execute()).To(MatchError(ContainSubstring("mutually exclusive")))
	})

	It("surfaces secure storage errors", func() {
		deps, _, creds, _ := testDependencies()
		creds.err = errors.New("keyring unavailable")
		command := newRootCommand(deps)
		command.SetArgs([]string{"auth", "login", "--hostname", "cloud.example.net", "--username", "alice"})
		Expect(command.Execute()).To(MatchError(ContainSubstring("store account credential")))
	})

	It("reports empty status", func() {
		deps, _, _, output := testDependencies()
		command := newRootCommand(deps)
		command.SetArgs([]string{"auth", "status"})
		Expect(command.Execute()).To(Succeed())
		Expect(output.String()).To(ContainSubstring("No accounts configured"))
	})
})
