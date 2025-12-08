package domain

import (
	"errors"
	"fmt"
)

// Common errors for the domain layer.
var (
	// ErrNotFound is returned when an entity is not found in the repository.
	ErrNotFound = errors.New("not found")

	// ErrUnauthorized is returned when the user is not authenticated.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when the user doesn't have permission.
	ErrForbidden = errors.New("forbidden")

	// ErrBadRequest is returned for invalid input data.
	ErrBadRequest = errors.New("bad request")

	// ErrConflict is returned when there's a resource conflict (e.g., duplicate).
	ErrConflict = errors.New("conflict")

	// ErrTooManyRequests is returned when rate limit is exceeded.
	ErrTooManyRequests = errors.New("too many requests")

	// ErrInternal is returned for unexpected server errors.
	ErrInternal = errors.New("internal server error")

	// ErrServiceUnavailable is returned when the service is temporarily unavailable.
	ErrServiceUnavailable = errors.New("service unavailable")
)

// AppError wraps an error with additional context for HTTP handling.
type AppError struct {
	Err     error  // The underlying error
	Message string // User-facing message
	Code    int    // HTTP status code
	Details string // Additional details for logging
}

func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Error constructors for common HTTP errors

func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Err:     ErrNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Code:    404,
	}
}

func NewUnauthorizedError(message string) *AppError {
	if message == "" {
		message = "You must be logged in to access this resource"
	}
	return &AppError{
		Err:     ErrUnauthorized,
		Message: message,
		Code:    401,
	}
}

func NewForbiddenError(message string) *AppError {
	if message == "" {
		message = "You don't have permission to access this resource"
	}
	return &AppError{
		Err:     ErrForbidden,
		Message: message,
		Code:    403,
	}
}

func NewBadRequestError(message string) *AppError {
	if message == "" {
		message = "Invalid request"
	}
	return &AppError{
		Err:     ErrBadRequest,
		Message: message,
		Code:    400,
	}
}

func NewConflictError(message string) *AppError {
	if message == "" {
		message = "Resource conflict"
	}
	return &AppError{
		Err:     ErrConflict,
		Message: message,
		Code:    409,
	}
}

func NewTooManyRequestsError(message string) *AppError {
	if message == "" {
		message = "Too many requests. Please try again later."
	}
	return &AppError{
		Err:     ErrTooManyRequests,
		Message: message,
		Code:    429,
	}
}

func NewInternalError(details string) *AppError {
	return &AppError{
		Err:     ErrInternal,
		Message: "Something went wrong. Please try again later.",
		Code:    500,
		Details: details,
	}
}

func NewServiceUnavailableError(message string) *AppError {
	if message == "" {
		message = "Service temporarily unavailable. Please try again later."
	}
	return &AppError{
		Err:     ErrServiceUnavailable,
		Message: message,
		Code:    503,
	}
}

// WrapError wraps a generic error with AppError context.
func WrapError(err error, message string, code int) *AppError {
	return &AppError{
		Err:     err,
		Message: message,
		Code:    code,
		Details: err.Error(),
	}
}

// IsAppError checks if an error is an AppError.
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError extracts AppError from an error if present.
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}
