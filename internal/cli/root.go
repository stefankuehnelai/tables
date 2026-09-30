// Package cli implements the tables command-line interface.
package cli

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	clioutput "github.com/stefankuehnelai/tables/internal/cli/output"
	"github.com/stefankuehnelai/tables/pkg/tables"
)

const (
	envHostname      = "NEXTCLOUD_TABLES_HOSTNAME"
	envServer        = "NEXTCLOUD_TABLES_SERVER"
	envUsername      = "NEXTCLOUD_TABLES_USERNAME"
	envPassword      = "NEXTCLOUD_TABLES_PASSWORD"
	envAppPassword   = "NEXTCLOUD_TABLES_APP_PASSWORD"
	envShareToken    = "NEXTCLOUD_TABLES_SHARE_TOKEN"
	envSharePassword = "NEXTCLOUD_TABLES_SHARE_PASSWORD"
)

type dependencies struct {
	config      configStore
	credentials credentialStore
	stdin       io.Reader
	stdout      io.Writer
	stderr      io.Writer
	getenv      func(string) string
	readSecret  func(string) (string, error)
}

type rootOptions struct {
	hostname string
	server   string
	username string
	json     string
	jq       string
	template string
}

type application struct {
	deps dependencies
	root rootOptions
}

// NewRootCommand constructs the tables CLI using production dependencies.
func NewRootCommand() (*cobra.Command, error) {
	configPath, err := defaultConfigPath()
	if err != nil {
		return nil, err
	}
	deps := dependencies{
		config:      newFileConfigStore(configPath),
		credentials: keyringCredentialStore{},
		stdin:       os.Stdin,
		stdout:      os.Stdout,
		stderr:      os.Stderr,
		getenv:      os.Getenv,
		readSecret:  terminalSecretReader(os.Stdin, os.Stderr),
	}
	return newRootCommand(deps), nil
}

func newRootCommand(deps dependencies) *cobra.Command {
	app := &application{deps: deps}
	command := &cobra.Command{Use: "tables", Short: "Work with Nextcloud Tables", SilenceUsage: true, SilenceErrors: true}
	command.SetIn(deps.stdin)
	command.SetOut(deps.stdout)
	command.SetErr(deps.stderr)
	flags := command.PersistentFlags()
	flags.StringVar(&app.root.hostname, "hostname", "", "Nextcloud hostname")
	flags.StringVar(&app.root.server, "server", "", "Nextcloud server URL")
	flags.StringVar(&app.root.username, "username", "", "Nextcloud username")
	flags.StringVar(&app.root.json, "json", "", "Output JSON with comma-separated fields")
	flags.StringVar(&app.root.jq, "jq", "", "Filter JSON output using a jq expression")
	flags.StringVar(&app.root.template, "template", "", "Format JSON output using a Go template")
	command.AddCommand(app.newAuthCommand())
	command.AddCommand(app.newListCommand())
	command.AddCommand(app.newGetCommand())
	command.AddCommand(app.newCreateCommand())
	command.AddCommand(app.newUpdateCommand())
	command.AddCommand(app.newDeleteCommand())
	command.AddCommand(app.newColumnsCommand())
	command.AddCommand(app.newRowsCommand())
	return command
}

func (a *application) formatter() (*clioutput.Formatter, error) {
	var fields []string
	if strings.TrimSpace(a.root.json) != "" {
		for _, field := range strings.Split(a.root.json, ",") {
			fields = append(fields, strings.TrimSpace(field))
		}
	}
	return clioutput.New(clioutput.Options{JSONFields: fields, JQ: a.root.jq, Template: a.root.template})
}

func (a *application) write(value any) error {
	formatter, err := a.formatter()
	if err != nil {
		return err
	}
	return formatter.Write(a.deps.stdout, value)
}

type connection struct {
	hostname string
	server   string
	username string
	kind     string
	secret   string
}

func (a *application) resolveConnection() (connection, error) {
	cfg, err := a.deps.config.Load()
	if err != nil {
		return connection{}, err
	}
	hostname := firstNonEmpty(a.root.hostname, a.deps.getenv(envHostname), cfg.ActiveHostname)
	if hostname == "" {
		return connection{}, tables.NewError(tables.CodeInvalidConfiguration, "no Nextcloud hostname configured; run tables auth login")
	}
	host := cfg.Hosts[hostname]
	username := firstNonEmpty(a.root.username, a.deps.getenv(envUsername), host.ActiveUsername)
	if username == "" {
		return connection{}, tables.NewError(tables.CodeInvalidConfiguration, "no Nextcloud username configured; run tables auth login")
	}
	server := firstNonEmpty(a.root.server, a.deps.getenv(envServer), host.Server)
	if server == "" {
		server = "https://" + hostname
	}
	if err := validateServer(server); err != nil {
		return connection{}, err
	}
	if secret := a.deps.getenv(envAppPassword); secret != "" {
		return connection{hostname: hostname, server: server, username: username, kind: "app-password", secret: secret}, nil
	}
	if secret := a.deps.getenv(envPassword); secret != "" {
		return connection{hostname: hostname, server: server, username: username, kind: "password", secret: secret}, nil
	}
	account, ok := host.Accounts[username]
	if !ok || account.CredentialKind == "" {
		return connection{}, tables.NewError(tables.CodeInvalidConfiguration, "no stored credential metadata for account")
	}
	secret, err := a.deps.credentials.Get(hostname, username, account.CredentialKind)
	if err != nil {
		return connection{}, fmt.Errorf("read account credential: %w", err)
	}
	return connection{hostname: hostname, server: server, username: username, kind: account.CredentialKind, secret: secret}, nil
}

func (a *application) client() (*tables.Client, error) {
	connection, err := a.resolveConnection()
	if err != nil {
		return nil, err
	}
	option := tables.WithCredentials(connection.username, connection.secret)
	if connection.kind == "app-password" {
		option = tables.WithAppPassword(connection.username, connection.secret)
	}
	return tables.NewClient(connection.server, option)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func validateServer(server string) error {
	parsed, err := url.Parse(server)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return tables.NewError(tables.CodeInvalidConfiguration, fmt.Sprintf("invalid Nextcloud server URL %q", server))
	}
	return nil
}

func terminalSecretReader(input *os.File, output io.Writer) func(string) (string, error) {
	return func(prompt string) (string, error) {
		if _, err := fmt.Fprint(output, prompt); err != nil {
			return "", err
		}
		if term.IsTerminal(int(input.Fd())) {
			value, err := term.ReadPassword(int(input.Fd()))
			if _, writeErr := fmt.Fprintln(output); writeErr != nil && err == nil {
				err = writeErr
			}
			return strings.TrimSpace(string(value)), err
		}
		scanner := bufio.NewScanner(input)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return "", err
			}
			return "", io.EOF
		}
		return strings.TrimSpace(scanner.Text()), nil
	}
}
