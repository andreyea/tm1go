package tm1

import (
	"fmt"
	"net/http"
)

// TM1Error represents a TM1-specific error
type TM1Error struct {
	Code       string        `json:"code"`
	Message    string        `json:"message"`
	StatusCode int           `json:"status_code"`
	Details    []ErrorDetail `json:"details,omitempty"`
}

// ErrorDetail represents additional error details
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *TM1Error) Error() string {
	if e.Details != nil && len(e.Details) > 0 {
		return fmt.Sprintf("TM1 Error [%d]: %s - %s (Code: %s)",
			e.StatusCode, e.Message, e.Details[0].Message, e.Code)
	}
	return fmt.Sprintf("TM1 Error [%d]: %s (Code: %s)", e.StatusCode, e.Message, e.Code)
}

// TM1RestException represents a REST API exception
type TM1RestException struct {
	*TM1Error
	URL    string
	Method string
}

// Error implements the error interface for TM1RestException
func (e *TM1RestException) Error() string {
	return fmt.Sprintf("TM1 REST Exception [%s %s]: %s", e.Method, e.URL, e.TM1Error.Error())
}

// TM1TimeoutException represents a timeout exception
type TM1TimeoutException struct {
	*TM1Error
	Timeout float64
}

// Error implements the error interface for TM1TimeoutException
func (e *TM1TimeoutException) Error() string {
	return fmt.Sprintf("TM1 Timeout Exception (%.2fs): %s", e.Timeout, e.TM1Error.Error())
}

// TM1VersionDeprecationException represents a version deprecation exception
type TM1VersionDeprecationException struct {
	*TM1Error
	Version string
}

// Error implements the error interface for TM1VersionDeprecationException
func (e *TM1VersionDeprecationException) Error() string {
	return fmt.Sprintf("TM1 Version Deprecation (v%s): %s", e.Version, e.TM1Error.Error())
}

// Common error constructors

// NewTM1Error creates a new TM1Error
func NewTM1Error(code, message string, statusCode int) *TM1Error {
	return &TM1Error{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// NewTM1RestException creates a new TM1RestException
func NewTM1RestException(method, url, code, message string, statusCode int) *TM1RestException {
	return &TM1RestException{
		TM1Error: NewTM1Error(code, message, statusCode),
		URL:      url,
		Method:   method,
	}
}

// NewTM1TimeoutException creates a new TM1TimeoutException
func NewTM1TimeoutException(timeout float64, message string) *TM1TimeoutException {
	return &TM1TimeoutException{
		TM1Error: NewTM1Error("TIMEOUT", message, http.StatusRequestTimeout),
		Timeout:  timeout,
	}
}

// NewTM1VersionDeprecationException creates a new TM1VersionDeprecationException
func NewTM1VersionDeprecationException(version, message string) *TM1VersionDeprecationException {
	return &TM1VersionDeprecationException{
		TM1Error: NewTM1Error("VERSION_DEPRECATED", message, http.StatusBadRequest),
		Version:  version,
	}
}
