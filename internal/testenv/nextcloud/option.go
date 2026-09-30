package nextcloud

import "time"

const (
	defaultNextcloudImage = "nextcloud:35.0.1-apache"
	defaultTablesVersion  = "2.3.1"
)

type settings struct {
	image          string
	tablesVersion  string
	adminUsername  string
	adminPassword  string
	startupTimeout time.Duration
}

// Option configures a Nextcloud test environment.
type Option func(*settings)

// WithImage overrides the pinned Nextcloud container image.
func WithImage(image string) Option { return func(s *settings) { s.image = image } }

// WithTablesVersion overrides the pinned Tables release version.
func WithTablesVersion(version string) Option { return func(s *settings) { s.tablesVersion = version } }

// WithAdminCredentials overrides the test administrator credentials.
func WithAdminCredentials(username, password string) Option {
	return func(s *settings) { s.adminUsername, s.adminPassword = username, password }
}

// WithStartupTimeout overrides the container readiness timeout.
func WithStartupTimeout(timeout time.Duration) Option {
	return func(s *settings) { s.startupTimeout = timeout }
}

func defaultSettings() settings {
	return settings{
		image: defaultNextcloudImage, tablesVersion: defaultTablesVersion,
		adminUsername: "admin", adminPassword: "nextcloud-tables-admin",
		startupTimeout: 3 * time.Minute,
	}
}
