package tables

import (
	"net/http"
	"strings"
)

// Option configures a Client.
type Option func(*Client) error

// WithCredentials configures username/password authentication.
func WithCredentials(username string, password string) Option {
	return withBasicAuth(username, password)
}

// WithAppPassword configures username/app-password authentication.
func WithAppPassword(username string, appPassword string) Option {
	return withBasicAuth(username, appPassword)
}

// WithHTTPClient replaces the default HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) error {
		if httpClient == nil {
			return NewError(CodeInvalidConfiguration, "HTTP client must not be nil")
		}
		client.httpClient = httpClient
		return nil
	}
}

// withBasicAuth validates and installs Basic authentication.
func withBasicAuth(username string, password string) Option {
	return func(client *Client) error {
		if strings.TrimSpace(username) == "" {
			return NewError(CodeInvalidConfiguration, "username must not be empty")
		}
		if password == "" {
			return NewError(CodeInvalidConfiguration, "password must not be empty")
		}
		client.auth = BasicAuth{Username: username, Password: password}
		return nil
	}
}
