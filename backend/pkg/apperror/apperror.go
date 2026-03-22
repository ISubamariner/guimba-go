package apperror

import (
	"fmt"
	"net/http"
)

// Code represents an application error code.
type Code string

const (
	CodeNotFound       Code = "NOT_FOUND"
	CodeValidation     Code = "VALIDATION_ERROR"
	CodeUnauthorized   Code = "UNAUTHORIZED"
	CodeForbidden      Code = "FORBIDDEN"
	CodeConflict       Code = "CONFLICT"
	CodeInternal       Code = "INTERNAL_ERROR"
	CodeBadRequest     Code = "BAD_REQUEST"
)

// AppError is the standard application error type.
type AppError struct {
	Code    Code     `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
	HTTPStatus int   `json:"-"`
	Err     error    `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewNotFound(resource string, id any) *AppError {
	return &AppError{
		Code:       CodeNotFound,
		Message:    fmt.Sprintf("%s with ID %v not found", resource, id),
		HTTPStatus: http.StatusNotFound,
	}
}

func NewValidation(message string, details ...string) *AppError {
	return &AppError{
		Code:       CodeValidation,
		Message:    message,
		Details:    details,
		HTTPStatus: http.StatusUnprocessableEntity,
	}
}

func NewUnauthorized(message string) *AppError {
	return &AppError{
		Code:       CodeUnauthorized,
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
	}
}

func NewForbidden(message string) *AppError {
	return &AppError{
		Code:       CodeForbidden,
		Message:    message,
		HTTPStatus: http.StatusForbidden,
	}
}

func NewConflict(message string) *AppError {
	return &AppError{
		Code:       CodeConflict,
		Message:    message,
		HTTPStatus: http.StatusConflict,
	}
}

func NewInternal(err error) *AppError {
	return &AppError{
		Code:       CodeInternal,
		Message:    "An internal error occurred",
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
	}
}

func NewBadRequest(message string) *AppError {
	return &AppError{
		Code:       CodeBadRequest,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}
