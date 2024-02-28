package client

import (
	"fmt"
	"net/http"
)

type WSConnError struct {
	StatusCode   int
	ResponseBody string
	ConnErr      error
	retryable    bool
}

var nonRetryableStatusCodes = []int{
	http.StatusBadRequest,
	http.StatusUnauthorized,
	http.StatusForbidden,
	http.StatusNotFound,
	http.StatusMethodNotAllowed,
	http.StatusNotImplemented,
	http.StatusHTTPVersionNotSupported,
}

func (e *WSConnError) Error() string {
	if e.ConnErr != nil {
		return fmt.Sprintf("Connection Error %s. Status Code: %d. Response: %s", e.ConnErr.Error(), e.StatusCode, e.ResponseBody)
	}

	return fmt.Sprintf("Connection Error. Status Code: %d. Response: %s", e.StatusCode, e.ResponseBody)
}

func (e *WSConnError) IsRetryable() bool {
	return e.retryable
}

func NewWSConnError(err error, statusCode int, respBody string) *WSConnError {
	return &WSConnError{ConnErr: err,
		StatusCode:   statusCode,
		ResponseBody: respBody,
		retryable:    isRetryableStatusCode(statusCode),
	}
}

// isRetryableStatusCode checks if the provided HTTP status code is retryable
func isRetryableStatusCode(statusCode int) bool {
	for _, nonRetryableCode := range nonRetryableStatusCodes {
		if statusCode == nonRetryableCode {
			return false
		}
	}

	return true
}
