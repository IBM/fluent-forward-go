package client

import (
	"fmt"
)

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP Connection Error %d: %s", e.StatusCode, e.Message)
}
func NewHTTPError(statusCode int, message string) *HTTPError {
	return &HTTPError{StatusCode: statusCode, Message: message}
}
