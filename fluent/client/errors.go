package client

import (
	"fmt"
	"net/http"
)

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP Error %d: %s", e.StatusCode, e.Message)
}

var (
	ErrUnauthorized = &HTTPError{StatusCode: http.StatusUnauthorized, Message: "Unauthorized"}
	ErrForbidden    = &HTTPError{StatusCode: http.StatusForbidden, Message: "Forbidden"}
	// Add more custom errors as needed
)
