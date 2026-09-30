package tables

import "net/http"

// Auth applies authentication to an HTTP request.
type Auth interface {
	Apply(*http.Request) error
}

// BasicAuth authenticates with a Nextcloud username and password or app password.
type BasicAuth struct {
	Username string
	Password string
}

// Apply adds HTTP Basic authentication to request.
func (a BasicAuth) Apply(request *http.Request) error {
	request.SetBasicAuth(a.Username, a.Password)
	return nil
}
