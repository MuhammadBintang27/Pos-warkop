package types

import (
	"fmt"
	"net/http"
)

type HTTPError struct {
	Status  int
	Message string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Status, e.Message)
}

func NewHTTPError(status int, message string) *HTTPError {
	return &HTTPError{Status: status, Message: message}
}

func BadRequest(message string) *HTTPError { return NewHTTPError(http.StatusBadRequest, message) }

func NotFound(what string) *HTTPError {
	return NewHTTPError(http.StatusNotFound, what+" tidak ditemukan")
}

func Forbidden(message string) *HTTPError { return NewHTTPError(http.StatusForbidden, message) }
