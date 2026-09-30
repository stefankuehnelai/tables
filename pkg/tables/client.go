package tables

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const (
	ocsAPIPrefix  = "/ocs/v2.php/apps/tables/api/2"
	jsonAPIPrefix = "/index.php/apps/tables/api/1"
)

// Client is a reusable Nextcloud Tables API client.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	auth       Auth
}

// NewClient validates rawURL and constructs a reusable API client.
func NewClient(rawURL string, options ...Option) (*Client, error) {
	baseURL, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, &Error{code: CodeInvalidConfiguration, message: "invalid server URL", cause: err}
	}
	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return nil, NewError(CodeInvalidConfiguration, "server URL must use http or https")
	}
	if baseURL.Host == "" {
		return nil, NewError(CodeInvalidConfiguration, "server URL must include a host")
	}
	baseURL.RawQuery = ""
	baseURL.Fragment = ""
	baseURL.Path = strings.TrimRight(baseURL.Path, "/")

	client := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, option := range options {
		if option == nil {
			return nil, NewError(CodeInvalidConfiguration, "client option must not be nil")
		}
		if err := option(client); err != nil {
			return nil, err
		}
	}
	return client, nil
}

// BaseURL returns a copy of the normalized server URL.
func (c *Client) BaseURL() url.URL {
	return *c.baseURL
}

// doOCS performs an OCS API request and decodes the envelope data into output.
func (c *Client) doOCS(ctx context.Context, httpClient *http.Client, method string, route string, query url.Values, input any, output any) error {
	var envelope struct {
		OCS struct {
			Meta struct {
				Status     string `json:"status"`
				StatusCode int    `json:"statuscode"`
				Message    string `json:"message"`
			} `json:"meta"`
			Data json.RawMessage `json:"data"`
		} `json:"ocs"`
	}
	if query == nil {
		query = url.Values{}
	}
	query.Set("format", "json")
	if err := c.do(ctx, httpClient, method, ocsAPIPrefix+route, query, input, &envelope); err != nil {
		return err
	}
	if envelope.OCS.Meta.StatusCode >= 400 {
		return errorForStatus(envelope.OCS.Meta.StatusCode, envelope.OCS.Meta.Message)
	}
	if output == nil || len(envelope.OCS.Data) == 0 || string(envelope.OCS.Data) == "null" {
		return nil
	}
	if err := json.Unmarshal(envelope.OCS.Data, output); err != nil {
		return &Error{code: CodeInvalidResponse, message: "decode OCS data", cause: err}
	}
	return nil
}

// doJSON performs a regular JSON API request.
func (c *Client) doJSON(ctx context.Context, method string, route string, query url.Values, input any, output any) error {
	return c.do(ctx, c.httpClient, method, jsonAPIPrefix+route, query, input, output)
}

// do executes a single HTTP request with common headers, authentication, and error mapping.
func (c *Client) do(ctx context.Context, httpClient *http.Client, method string, route string, query url.Values, input any, output any) error {
	requestURL := *c.baseURL
	requestURL.Path = joinURLPath(c.baseURL.Path, route)
	requestURL.RawQuery = query.Encode()

	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return &Error{code: CodeInvalidArgument, message: "encode request body", cause: err}
		}
		body = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, requestURL.String(), body)
	if err != nil {
		return &Error{code: CodeInvalidArgument, message: "create request", cause: err}
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("OCS-APIRequest", "true")
	if input != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if c.auth != nil {
		if err := c.auth.Apply(request); err != nil {
			return err
		}
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return &Error{code: CodeTransport, message: "perform HTTP request", cause: err}
	}
	defer func() { _ = response.Body.Close() }()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return &Error{code: CodeInvalidResponse, message: "read response body", cause: err}
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return errorForStatus(response.StatusCode, responseMessage(responseBody))
	}
	if output == nil || len(responseBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(responseBody, output); err != nil {
		return &Error{code: CodeInvalidResponse, message: "decode JSON response", cause: err}
	}
	return nil
}

// responseMessage extracts a Nextcloud message field while retaining a useful fallback.
func responseMessage(body []byte) string {
	var value struct {
		Message string `json:"message"`
		OCS     struct {
			Meta struct {
				Message string `json:"message"`
			} `json:"meta"`
		} `json:"ocs"`
	}
	if json.Unmarshal(body, &value) == nil {
		if value.Message != "" {
			return value.Message
		}
		if value.OCS.Meta.Message != "" {
			return value.OCS.Meta.Message
		}
	}
	trimmed := strings.TrimSpace(string(body))
	if len(trimmed) > 240 {
		return trimmed[:240]
	}
	return trimmed
}

// joinURLPath preserves an installation subpath while appending an API route.
func joinURLPath(basePath string, route string) string {
	joined := path.Join("/", basePath, route)
	if strings.HasSuffix(route, "/") && !strings.HasSuffix(joined, "/") {
		joined += "/"
	}
	return joined
}

// tableID validates a positive Nextcloud identifier.
func tableID(id int64, name string) error {
	if id <= 0 {
		return NewError(CodeInvalidArgument, fmt.Sprintf("%s must be positive", name))
	}
	return nil
}
