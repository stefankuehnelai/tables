package tables

import "net/http"

// Auth applies authentication to an HTTP request.
type Auth interface {
	Apply(*http.Request) error
}

// BasicAuth applies Nextcloud username/password Basic authentication.
type BasicAuth struct {
	Username string
	Password string
}

// Apply adds the Basic authentication header to request.
func (a BasicAuth) Apply(request *http.Request) error {
	request.SetBasicAuth(a.Username, a.Password)
	return nil
}
