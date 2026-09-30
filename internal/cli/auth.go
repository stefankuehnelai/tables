package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

type loginOptions struct {
	passwordStdin    bool
	appPasswordStdin bool
}

func (a *application) newAuthCommand() *cobra.Command {
	command := &cobra.Command{Use: "auth", Short: "Manage Nextcloud authentication"}
	command.AddCommand(a.newAuthLoginCommand())
	command.AddCommand(a.newAuthStatusCommand())
	command.AddCommand(a.newAuthSwitchCommand())
	command.AddCommand(a.newAuthLogoutCommand())
	return command
}

func (a *application) newAuthLoginCommand() *cobra.Command {
	options := loginOptions{}
	command := &cobra.Command{
		Use:   "login",
		Short: "Log in to a Nextcloud account",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return a.login(command, options)
		},
	}
	flags := command.Flags()
	flags.BoolVar(&options.passwordStdin, "password-stdin", false, "Read password from stdin")
	flags.BoolVar(&options.appPasswordStdin, "app-password-stdin", false, "Read app password from stdin")
	return command
}

func (a *application) login(command *cobra.Command, options loginOptions) error {
	if options.passwordStdin && options.appPasswordStdin {
		return tables.NewError(tables.CodeInvalidArgument, "--password-stdin and --app-password-stdin are mutually exclusive")
	}
	hostname := firstNonEmpty(a.root.hostname, a.deps.getenv(envHostname))
	if hostname == "" {
		return tables.NewError(tables.CodeInvalidArgument, "--hostname is required")
	}
	username := firstNonEmpty(a.root.username, a.deps.getenv(envUsername))
	if username == "" {
		return tables.NewError(tables.CodeInvalidArgument, "--username is required")
	}
	server := firstNonEmpty(a.root.server, a.deps.getenv(envServer), "https://"+hostname)
	if err := validateServer(server); err != nil {
		return err
	}

	kind, secret, err := a.loginSecret(command, options)
	if err != nil {
		return err
	}
	if secret == "" {
		return tables.NewError(tables.CodeInvalidArgument, "credential must not be empty")
	}

	if err := a.deps.credentials.Set(hostname, username, kind, secret); err != nil {
		return fmt.Errorf("store account credential: %w", err)
	}
	cfg, err := a.deps.config.Load()
	if err != nil {
		_ = a.deps.credentials.Delete(hostname, username, kind)
		return err
	}
	host := cfg.Hosts[hostname]
	if host.Accounts == nil {
		host.Accounts = map[string]accountConfig{}
	}
	host.Server = server
	host.ActiveUsername = username
	host.Accounts[username] = accountConfig{CredentialKind: kind}
	cfg.Hosts[hostname] = host
	cfg.ActiveHostname = hostname
	if err := a.deps.config.Save(cfg); err != nil {
		_ = a.deps.credentials.Delete(hostname, username, kind)
		return err
	}
	_, err = fmt.Fprintf(a.deps.stdout, "Logged in to %s as %s\n", hostname, username)
	return err
}

func (a *application) loginSecret(command *cobra.Command, options loginOptions) (string, string, error) {
	if value := a.deps.getenv(envAppPassword); value != "" {
		return "app-password", value, nil
	}
	if value := a.deps.getenv(envPassword); value != "" {
		return "password", value, nil
	}
	if options.appPasswordStdin {
		value, err := readOneLine(command.InOrStdin())
		return "app-password", value, err
	}
	if options.passwordStdin {
		value, err := readOneLine(command.InOrStdin())
		return "password", value, err
	}
	value, err := a.deps.readSecret("Password or app password: ")
	return "password", value, err
}

func (a *application) newAuthStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show configured accounts",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := a.deps.config.Load()
			if err != nil {
				return err
			}
			if len(cfg.Hosts) == 0 {
				_, err = fmt.Fprintln(a.deps.stdout, "No accounts configured")
				return err
			}
			for hostname, host := range cfg.Hosts {
				for username := range host.Accounts {
					marker := " "
					if cfg.ActiveHostname == hostname && host.ActiveUsername == username {
						marker = "*"
					}
					if _, err := fmt.Fprintf(a.deps.stdout, "%s %s %s %s\n", marker, hostname, username, host.Server); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}
}

func (a *application) newAuthSwitchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "switch",
		Short: "Switch the active account",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			targetHost := firstNonEmpty(a.root.hostname, a.deps.getenv(envHostname))
			targetUser := firstNonEmpty(a.root.username, a.deps.getenv(envUsername))
			if targetHost == "" || targetUser == "" {
				return tables.NewError(tables.CodeInvalidArgument, "--hostname and --username are required")
			}
			cfg, err := a.deps.config.Load()
			if err != nil {
				return err
			}
			host, ok := cfg.Hosts[targetHost]
			if !ok {
				return tables.NewError(tables.CodeNotFound, "hostname is not configured")
			}
			if _, ok := host.Accounts[targetUser]; !ok {
				return tables.NewError(tables.CodeNotFound, "username is not configured for hostname")
			}
			host.ActiveUsername = targetUser
			cfg.Hosts[targetHost] = host
			cfg.ActiveHostname = targetHost
			if err := a.deps.config.Save(cfg); err != nil {
				return err
			}
			_, err = fmt.Fprintf(a.deps.stdout, "Switched to %s as %s\n", targetHost, targetUser)
			return err
		},
	}
}

func (a *application) newAuthLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Log out from a Nextcloud account",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := a.deps.config.Load()
			if err != nil {
				return err
			}
			targetHost := firstNonEmpty(a.root.hostname, a.deps.getenv(envHostname), cfg.ActiveHostname)
			host, ok := cfg.Hosts[targetHost]
			if !ok {
				return tables.NewError(tables.CodeNotFound, "hostname is not configured")
			}
			targetUser := firstNonEmpty(a.root.username, a.deps.getenv(envUsername), host.ActiveUsername)
			account, ok := host.Accounts[targetUser]
			if !ok {
				return tables.NewError(tables.CodeNotFound, "username is not configured for hostname")
			}
			if err := a.deps.credentials.Delete(targetHost, targetUser, account.CredentialKind); err != nil {
				return fmt.Errorf("delete account credential: %w", err)
			}
			delete(host.Accounts, targetUser)
			if host.ActiveUsername == targetUser {
				host.ActiveUsername = firstAccount(host.Accounts)
			}
			if len(host.Accounts) == 0 {
				delete(cfg.Hosts, targetHost)
				if cfg.ActiveHostname == targetHost {
					cfg.ActiveHostname = firstHost(cfg.Hosts)
				}
			} else {
				cfg.Hosts[targetHost] = host
			}
			if err := a.deps.config.Save(cfg); err != nil {
				return err
			}
			_, err = fmt.Fprintf(a.deps.stdout, "Logged out from %s as %s\n", targetHost, targetUser)
			return err
		},
	}
}

func readOneLine(reader io.Reader) (string, error) {
	scanner := bufio.NewScanner(reader)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return strings.TrimSpace(scanner.Text()), nil
}

func firstAccount(accounts map[string]accountConfig) string {
	for username := range accounts {
		return username
	}
	return ""
}

func firstHost(hosts map[string]hostConfig) string {
	for hostname := range hosts {
		return hostname
	}
	return ""
}
