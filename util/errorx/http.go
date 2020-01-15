package errorx

import (
	"net/http"
)

// Unauthorized generates a 401 error.
func BadRequest(format string, args ...interface{}) Error {
	return NewWithSkip(3, nil, http.StatusBadRequest, format, args...)
}

// Unauthorized generates a 401 error.
func Unauthorized(format string, args ...interface{}) Error {
	return NewWithSkip(3, nil, http.StatusUnauthorized, format, args...)
}

// Forbidden generates a 403 error.
func Forbidden(format string, args ...interface{}) Error {
	return NewWithSkip(3, nil, http.StatusForbidden, format, args...)
}

// NotFound generates a 404 error.
func NotFound(format string, args ...interface{}) Error {
	return NewWithSkip(3, nil, http.StatusNotFound, format, args...)
}

// MethodNotAllowed generates a 405 error.
func MethodNotAllowed(format string, args ...interface{}) Error {
	return NewWithSkip(3, nil, http.StatusMethodNotAllowed, format, args...)
}

// Timeout generates a 408 error.
func Timeout(format string, args ...interface{}) Error {
	return NewWithSkip(3, nil, http.StatusRequestTimeout, format, args...)
}

// Conflict generates a 409 error.
func Conflict(format string, args ...interface{}) Error {
	return NewWithSkip(3, nil, http.StatusConflict, format, args...)
}

// InternalServerError generates a 500 error.
func InternalServerError(format string, args ...interface{}) Error {
	return NewWithSkip(3, nil, http.StatusInternalServerError, format, args...)
}
