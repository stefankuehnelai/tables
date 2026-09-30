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

// ShareToken identifies a public Nextcloud Tables share.
type ShareToken string

// SharePassword is an optional password protecting a public share.
type SharePassword string

// Share identifies a public row scope.
type Share struct {
	Token    ShareToken
	Password SharePassword
}

// ShareScope binds operations to one isolated public-share session.
type ShareScope struct {
	client *Client
	state  *shareState
}

// shareState owns one share's in-memory cookie jar and lazy authentication state.
type shareState struct {
	share      Share
	httpClient *http.Client
	once       sync.Once
	err        error
}

// Share creates a lightweight public-share scope with an isolated in-memory cookie jar.
func (c *Client) Share(share Share) ShareScope {
	jar, _ := cookiejar.New(nil)
	httpClient := *c.httpClient
	httpClient.Jar = jar
	return ShareScope{
		client: c,
		state: &shareState{
			share:      share,
			httpClient: &httpClient,
		},
	}
}

// Rows returns the common row CRUD API scoped to this public share.
func (s ShareScope) Rows() Rows {
	return Rows{client: s.client, httpClient: s.state.httpClient, share: s.state}
}

// authenticate lazily authenticates a protected share exactly once per scope.
func (s *shareState) authenticate(ctx context.Context, client *Client) error {
	if strings.TrimSpace(string(s.share.Token)) == "" {
		return NewError(CodeInvalidArgument, "share token must not be empty")
	}
	if s.share.Password == "" {
		return nil
	}
	s.once.Do(func() {
		s.err = s.login(ctx, client)
	})
	return s.err
}

// login posts the share password and retains only in-memory session cookies.
func (s *shareState) login(ctx context.Context, client *Client) error {
	requestURL := *client.baseURL
	requestURL.Path = joinURLPath(client.baseURL.Path, "/index.php/apps/tables/s/"+url.PathEscape(string(s.share.Token))+"/authenticate")
	form := url.Values{"password": []string{string(s.share.Password)}}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return &Error{code: CodeInvalidArgument, message: "create share authentication request", cause: err}
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpClient := *s.httpClient
	httpClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return &Error{code: CodeTransport, message: "authenticate public share", cause: err}
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusFound || response.StatusCode == http.StatusSeeOther {
		return nil
	}
	if response.StatusCode >= http.StatusBadRequest {
		return errorForStatus(response.StatusCode, "public share authentication failed")
	}
	return NewError(CodeAuthenticationFailed, fmt.Sprintf("public share authentication failed with HTTP %d", response.StatusCode))
}
