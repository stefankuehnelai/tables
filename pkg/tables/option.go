package tables

import "net/http"

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

// WithHTTPClient configures the HTTP client used for requests.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(client *Client) error {
		if httpClient == nil {
			return &Error{ErrorCode: CodeInvalidArgument, Message: "HTTP client must not be nil"}
		}
		client.httpClient = httpClient
		return nil
	}
}

// withBasicAuth configures the shared HTTP Basic authentication mechanism.
func withBasicAuth(username string, password string) Option {
	return func(client *Client) error {
		if username == "" || password == "" {
			return &Error{ErrorCode: CodeInvalidArgument, Message: "username and password must not be empty"}
		}
		client.auth = BasicAuth{Username: username, Password: password}
		return nil
	}
}
