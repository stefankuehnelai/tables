package tables

import (
	"errors"
	"fmt"
	"net/http"
)

// TablesError is the stable public error contract used by the library and CLI.
//revive:disable-next-line:exported TablesError is the documented stable public API name.
type TablesError interface {
	error
	Code() int
}

const (
	// CodeOK indicates success.
	CodeOK int = iota
	// CodeUnknown indicates an uncategorized failure.
	CodeUnknown
)

const (
	// CodeInvalidArgument indicates invalid caller input.
	CodeInvalidArgument int = iota + 10
	// CodeInvalidConfiguration indicates invalid client configuration.
	CodeInvalidConfiguration
)

const (
	// CodeAuthenticationFailed indicates failed credential or share-password authentication.
	CodeAuthenticationFailed int = iota + 20
	// CodeUnauthorized indicates an HTTP 401 response.
	CodeUnauthorized
	// CodeForbidden indicates an HTTP 403 response.
	CodeForbidden
)

const (
	// CodeNotFound indicates an HTTP 404 response.
	CodeNotFound int = iota + 30
	// CodeConflict indicates an HTTP 409 response.
	CodeConflict
)

const (
	// CodeTransport indicates an HTTP transport failure.
	CodeTransport int = iota + 40
	// CodeInvalidResponse indicates a response that could not be decoded.
	CodeInvalidResponse
	// CodeServer indicates a server-side HTTP failure.
	CodeServer
)

const (
	// CodeCredentialStore indicates an OS credential-store failure.
	CodeCredentialStore int = iota + 50
)

// Error is the concrete implementation of TablesError.
type Error struct {
	code       int
	statusCode int
	message    string
	cause      error
}

// NewError creates a stable library error with the given code and message.
func NewError(code int, message string) *Error {
	return &Error{code: code, message: message}
}

// Error returns the human-readable error text.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.statusCode != 0 {
		return fmt.Sprintf("tables: %s (HTTP %d)", e.message, e.statusCode)
	}
	return "tables: " + e.message
}

// Code returns the stable numeric error code.
func (e *Error) Code() int {
	if e == nil {
		return CodeUnknown
	}
	return e.code
}

// HTTPStatus returns the associated HTTP status, or zero for non-HTTP errors.
func (e *Error) HTTPStatus() int {
	if e == nil {
		return 0
	}
	return e.statusCode
}

// Unwrap exposes an underlying transport or decoding error.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// IsCode reports whether err implements TablesError with the requested code.
func IsCode(err error, code int) bool {
	var tablesErr TablesError
	return errors.As(err, &tablesErr) && tablesErr.Code() == code
}

// errorForStatus maps HTTP response status codes onto stable library codes.
func errorForStatus(status int, message string) *Error {
	code := CodeUnknown
	switch status {
	case http.StatusBadRequest:
		code = CodeInvalidArgument
	case http.StatusUnauthorized:
		code = CodeUnauthorized
	case http.StatusForbidden:
		code = CodeForbidden
	case http.StatusNotFound:
		code = CodeNotFound
	case http.StatusConflict:
		code = CodeConflict
	default:
		if status >= http.StatusInternalServerError {
			code = CodeServer
		}
	}
	if message == "" {
		message = http.StatusText(status)
	}
	return &Error{code: code, statusCode: status, message: message}
}
