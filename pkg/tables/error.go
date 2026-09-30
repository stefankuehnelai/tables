package tables

import "fmt"

// TablesError is an error with a stable process-compatible numeric code.
type TablesError interface {
	error
	Code() int
}

const (
	// CodeOK indicates success.
	CodeOK int = iota
	// CodeUnknown indicates an unclassified error.
	CodeUnknown
)

const (
	// CodeInvalidArgument indicates an invalid argument.
	CodeInvalidArgument int = iota + 10
	// CodeInvalidConfiguration indicates invalid client or CLI configuration.
	CodeInvalidConfiguration
)

const (
	// CodeAuthenticationFailed indicates invalid credentials.
	CodeAuthenticationFailed int = iota + 20
	// CodeUnauthorized indicates a request that requires authentication.
	CodeUnauthorized
	// CodeForbidden indicates insufficient permissions.
	CodeForbidden
)

const (
	// CodeNotFound indicates a missing resource.
	CodeNotFound int = iota + 30
	// CodeConflict indicates a conflicting resource state.
	CodeConflict
)

const (
	// CodeTransport indicates an HTTP transport failure.
	CodeTransport int = iota + 40
	// CodeInvalidResponse indicates an unexpected server response.
	CodeInvalidResponse
	// CodeServer indicates a server-side failure.
	CodeServer
)

const (
	// CodeCredentialStore indicates secure credential-store failure.
	CodeCredentialStore int = iota + 50
)

// Error is the default TablesError implementation.
type Error struct {
	ErrorCode int
	Message   string
	Status    int
	Cause     error
}

// Error returns the human-readable error message.
func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Cause != nil && e.Message == "" {
		return e.Cause.Error()
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error, if any.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Code returns the stable numeric error code.
func (e *Error) Code() int {
	if e == nil {
		return CodeOK
	}
	return e.ErrorCode
}
