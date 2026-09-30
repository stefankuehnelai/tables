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

const defaultTimeout = 30 * time.Second

// Client is a reusable Nextcloud Tables API client.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	auth       Auth
}

type ocsEnvelope struct {
	OCS struct {
		Meta struct {
			Status     string `json:"status"`
			StatusCode int    `json:"statuscode"`
			Message    string `json:"message"`
		} `json:"meta"`
		Data json.RawMessage `json:"data"`
	} `json:"ocs"`
}

// NewClient creates a client for rawURL.
func NewClient(rawURL string, options ...Option) (*Client, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, &Error{ErrorCode: CodeInvalidArgument, Message: "invalid server URL", Cause: err}
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, &Error{ErrorCode: CodeInvalidArgument, Message: "server URL must use http or https"}
	}
	if parsed.Host == "" {
		return nil, &Error{ErrorCode: CodeInvalidArgument, Message: "server URL must include a host"}
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	parsed.Path = strings.TrimRight(parsed.Path, "/")

	client := &Client{
		baseURL: parsed,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
	for _, option := range options {
		if option == nil {
			return nil, &Error{ErrorCode: CodeInvalidArgument, Message: "client option must not be nil"}
		}
		if err := option(client); err != nil {
			return nil, err
		}
	}
	return client, nil
}

// BaseURL returns a copy of the normalized server URL.
func (c *Client) BaseURL() *url.URL {
	if c == nil || c.baseURL == nil {
		return nil
	}
	copyURL := *c.baseURL
	return &copyURL
}

// endpoint resolves an API path against the configured Nextcloud server URL.
func (c *Client) endpoint(apiPath string, query url.Values) string {
	resolved := *c.baseURL
	resolved.Path = path.Join(c.baseURL.Path, apiPath)
	resolved.RawQuery = query.Encode()
	return resolved.String()
}

// doOCS executes one OCS JSON request and decodes its data payload.
func (c *Client) doOCS(
	ctx context.Context,
	httpClient *http.Client,
	auth Auth,
	method string,
	apiPath string,
	query url.Values,
	body any,
	out any,
) error {
	if query == nil {
		query = make(url.Values)
	}
	if query.Get("format") == "" {
		query.Set("format", "json")
	}
	var requestBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return &Error{ErrorCode: CodeInvalidArgument, Message: "encode request body", Cause: err}
		}
		requestBody = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, c.endpoint(apiPath, query), requestBody)
	if err != nil {
		return &Error{ErrorCode: CodeInvalidArgument, Message: "create request", Cause: err}
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("OCS-APIRequest", "true")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if auth != nil {
		if err := auth.Apply(request); err != nil {
			return &Error{ErrorCode: CodeAuthenticationFailed, Message: "apply authentication", Cause: err}
		}
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return &Error{ErrorCode: CodeTransport, Message: "perform request", Cause: err}
	}
	defer response.Body.Close()

	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return &Error{ErrorCode: CodeInvalidResponse, Message: "read response", Cause: err}
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return responseError(response.StatusCode, payload)
	}
	if out == nil {
		return nil
	}
	var envelope ocsEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return &Error{ErrorCode: CodeInvalidResponse, Message: "decode OCS response", Cause: err}
	}
	if envelope.OCS.Meta.Status == "failure" || envelope.OCS.Meta.StatusCode >= http.StatusBadRequest {
		return &Error{
			ErrorCode: statusCodeToErrorCode(envelope.OCS.Meta.StatusCode),
			Message:   envelope.OCS.Meta.Message,
			Status:    envelope.OCS.Meta.StatusCode,
		}
	}
	if len(envelope.OCS.Data) == 0 || string(envelope.OCS.Data) == "null" {
		return &Error{ErrorCode: CodeInvalidResponse, Message: "response does not contain OCS data"}
	}
	if err := json.Unmarshal(envelope.OCS.Data, out); err != nil {
		return &Error{ErrorCode: CodeInvalidResponse, Message: "decode response data", Cause: err}
	}
	return nil
}

// doJSON executes one regular JSON request and decodes its response body.
func (c *Client) doJSON(
	ctx context.Context,
	httpClient *http.Client,
	auth Auth,
	method string,
	apiPath string,
	query url.Values,
	body any,
	out any,
) error {
	var requestBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return &Error{ErrorCode: CodeInvalidArgument, Message: "encode request body", Cause: err}
		}
		requestBody = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, c.endpoint(apiPath, query), requestBody)
	if err != nil {
		return &Error{ErrorCode: CodeInvalidArgument, Message: "create request", Cause: err}
	}
	request.Header.Set("Accept", "application/json")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if auth != nil {
		if err := auth.Apply(request); err != nil {
			return &Error{ErrorCode: CodeAuthenticationFailed, Message: "apply authentication", Cause: err}
		}
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return &Error{ErrorCode: CodeTransport, Message: "perform request", Cause: err}
	}
	defer response.Body.Close()

	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return &Error{ErrorCode: CodeInvalidResponse, Message: "read response", Cause: err}
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return responseError(response.StatusCode, payload)
	}
	if out == nil {
		return nil
	}
	if len(bytes.TrimSpace(payload)) == 0 {
		return &Error{ErrorCode: CodeInvalidResponse, Message: "response body is empty"}
	}
	if err := json.Unmarshal(payload, out); err != nil {
		return &Error{ErrorCode: CodeInvalidResponse, Message: "decode JSON response", Cause: err}
	}
	return nil
}

// responseError converts an HTTP failure response into a stable Tables error.
func responseError(status int, payload []byte) error {
	message := strings.TrimSpace(string(payload))
	var envelope ocsEnvelope
	if json.Unmarshal(payload, &envelope) == nil && envelope.OCS.Meta.Message != "" {
		message = envelope.OCS.Meta.Message
	} else {
		var regular struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(payload, &regular) == nil && regular.Message != "" {
			message = regular.Message
		}
	}
	if message == "" {
		message = http.StatusText(status)
	}
	return &Error{ErrorCode: statusCodeToErrorCode(status), Message: message, Status: status}
}

// statusCodeToErrorCode maps HTTP or OCS status values to stable error codes.
func statusCodeToErrorCode(status int) int {
	switch status {
	case http.StatusBadRequest:
		return CodeInvalidArgument
	case http.StatusUnauthorized:
		return CodeUnauthorized
	case http.StatusForbidden:
		return CodeForbidden
	case http.StatusNotFound:
		return CodeNotFound
	case http.StatusConflict:
		return CodeConflict
	default:
		if status >= http.StatusInternalServerError {
			return CodeServer
		}
		return CodeUnknown
	}
}

// validateID validates a positive numeric API identifier.
func validateID(name string, id int64) error {
	if id <= 0 {
		return &Error{ErrorCode: CodeInvalidArgument, Message: fmt.Sprintf("%s must be greater than zero", name)}
	}
	return nil
}
