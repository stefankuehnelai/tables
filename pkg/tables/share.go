package tables

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
)

// ShareToken identifies a public Tables share.
type ShareToken string

// SharePassword is the optional password for a public Tables share.
type SharePassword string

// Share contains credentials for a public Tables share.
type Share struct {
	Token    ShareToken
	Password SharePassword
}

// ShareScope identifies one public share and owns isolated session cookies.
type ShareScope struct {
	client     *Client
	share      Share
	httpClient *http.Client
	authOnce   sync.Once
	authErr    error
}

// Share returns an immutable public-share scope with an isolated in-memory cookie jar.
func (c *Client) Share(share Share) *ShareScope {
	jar, _ := cookiejar.New(nil)
	copyClient := *c.httpClient
	copyClient.Jar = jar
	return &ShareScope{client: c, share: share, httpClient: &copyClient}
}

// Rows returns row operations scoped to the public share.
func (s *ShareScope) Rows() Rows {
	if s == nil {
		return Rows{}
	}
	return Rows{client: s.client, share: s}
}

// ensureAuthenticated authenticates a password-protected share at most once per scope.
func (s *ShareScope) ensureAuthenticated(ctx context.Context) error {
	if s == nil {
		return &Error{ErrorCode: CodeInvalidArgument, Message: "share scope must not be nil"}
	}
	if s.share.Token == "" {
		return &Error{ErrorCode: CodeInvalidArgument, Message: "share token must not be empty"}
	}
	if s.share.Password == "" {
		return nil
	}
	s.authOnce.Do(func() {
		s.authErr = s.authenticate(ctx)
	})
	return s.authErr
}

// authenticate establishes the isolated Nextcloud public-share session cookie.
func (s *ShareScope) authenticate(ctx context.Context) error {
	form := url.Values{"password": {string(s.share.Password)}, "passwordRequest": {"no"}}
	apiPath := fmt.Sprintf("/index.php/apps/tables/s/%s/authenticate", url.PathEscape(string(s.share.Token)))
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, s.client.endpoint(apiPath, nil), strings.NewReader(form.Encode()))
	if err != nil {
		return &Error{ErrorCode: CodeInvalidArgument, Message: "create share authentication request", Cause: err}
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := *s.httpClient
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	response, err := client.Do(request)
	if err != nil {
		return &Error{ErrorCode: CodeTransport, Message: "authenticate public share", Cause: err}
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusMultipleChoices && response.StatusCode < http.StatusBadRequest {
		return nil
	}
	if response.StatusCode == http.StatusOK {
		return &Error{ErrorCode: CodeAuthenticationFailed, Message: "public share password authentication failed", Status: response.StatusCode}
	}
	return responseError(response.StatusCode, nil)
}
