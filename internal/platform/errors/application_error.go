package errors

import "net/http"

type ApplicationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Status  int    `json:"-"`
}

func (e *ApplicationError) Error() string { return e.Code + ": " + e.Message }

func New(code string, status int, message string) *ApplicationError {
	return &ApplicationError{Code: code, Status: status, Message: message}
}

var InternalServerError = New("common:internal-server-error", http.StatusInternalServerError, "internal server error")
