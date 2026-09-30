// Package nextcloud provisions ephemeral Nextcloud instances for integration tests.
package nextcloud

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/stefankuehnelai/tables/pkg/tables"
)

// Nextcloud owns an ephemeral Nextcloud server and its administrator account.
type Nextcloud struct {
	URL string

	AdminUsername string
	AdminPassword string

	container testcontainers.Container
}

// New starts Nextcloud, installs the requested Tables release, and returns a reachable server.
func New(ctx context.Context, options ...Option) (*Nextcloud, error) {
	cfg := defaultSettings()
	for _, option := range options {
		option(&cfg)
	}
	if cfg.image == "" || cfg.tablesVersion == "" || cfg.adminUsername == "" || cfg.adminPassword == "" {
		return nil, fmt.Errorf("nextcloud test environment options must not be empty")
	}

	request := testcontainers.ContainerRequest{
		Image:        cfg.image,
		ExposedPorts: []string{"80/tcp"},
		Env: map[string]string{
			"NEXTCLOUD_ADMIN_USER":     cfg.adminUsername,
			"NEXTCLOUD_ADMIN_PASSWORD": cfg.adminPassword,
			"SQLITE_DATABASE":          "nextcloud",
		},
		WaitingFor: wait.ForHTTP("/status.php").WithPort("80/tcp").WithStartupTimeout(cfg.startupTimeout),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: request, Started: true})
	if err != nil {
		return nil, fmt.Errorf("start Nextcloud container: %w", err)
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = container.Terminate(context.Background())
		}
	}()

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve Nextcloud host: %w", err)
	}
	port, err := container.MappedPort(ctx, "80/tcp")
	if err != nil {
		return nil, fmt.Errorf("resolve Nextcloud port: %w", err)
	}
	instance := &Nextcloud{
		URL: "http://" + host + ":" + port.Port(), AdminUsername: cfg.adminUsername,
		AdminPassword: cfg.adminPassword, container: container,
	}
	if _, err := instance.OCC(ctx, "config:system:set", "trusted_domains", "1", "--value="+host); err != nil {
		return nil, fmt.Errorf("trust mapped Nextcloud host: %w", err)
	}
	if err := instance.installTables(ctx, cfg.tablesVersion); err != nil {
		return nil, err
	}
	if err := instance.waitForTables(ctx, cfg.startupTimeout); err != nil {
		return nil, err
	}
	cleanup = false
	return instance, nil
}

// Close stops and removes the owned container.
func (n *Nextcloud) Close() error {
	if n == nil || n.container == nil {
		return nil
	}
	return n.container.Terminate(context.Background())
}

// Client creates an authenticated Tables client for this server.
func (n *Nextcloud) Client(username, password string) (*tables.Client, error) {
	return tables.NewClient(n.URL, tables.WithCredentials(username, password))
}

func (n *Nextcloud) installTables(ctx context.Context, version string) error {
	releaseURL := fmt.Sprintf("https://github.com/nextcloud-releases/tables/releases/download/v%s/tables-v%s.tar.gz", version, version)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, releaseURL, nil)
	if err != nil {
		return fmt.Errorf("create Tables release request: %w", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("download Tables %s: %w", version, err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("download Tables %s: HTTP %d", version, response.StatusCode)
	}
	archive, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read Tables release: %w", err)
	}
	if err := n.container.CopyToContainer(ctx, archive, "/tmp/tables.tar.gz", 0o644); err != nil {
		return fmt.Errorf("copy Tables release to container: %w", err)
	}
	code, output, err := n.container.Exec(ctx, []string{"tar", "-xzf", "/tmp/tables.tar.gz", "-C", "/var/www/html/custom_apps"})
	if err != nil {
		return fmt.Errorf("extract Tables release: %w", err)
	}
	body, _ := io.ReadAll(output)
	if code != 0 {
		return fmt.Errorf("extract Tables release: exit %d: %s", code, body)
	}
	if _, err := n.OCC(ctx, "app:enable", "tables"); err != nil {
		return fmt.Errorf("enable Tables: %w", err)
	}
	return nil
}

func (n *Nextcloud) waitForTables(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		client, err := n.Client(n.AdminUsername, n.AdminPassword)
		if err == nil {
			if _, err = client.ListTables(ctx, tables.ListTablesOptions{}); err == nil {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return fmt.Errorf("tables did not become ready before timeout")
}
