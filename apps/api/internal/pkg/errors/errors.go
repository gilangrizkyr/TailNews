package errors

import (
	"errors"
	"net/http"
)

var (
	ErrNotFound       = errors.New("resource not found")
	ErrUnauthorized   = errors.New("unauthorized access")
	ErrForbidden      = errors.New("forbidden: insufficient permissions")
	ErrConflict       = errors.New("resource conflict or duplicate entry")
	ErrBadRequest     = errors.New("invalid request parameters")
	ErrInternalServer = errors.New("internal server error")
	ErrArticleGone    = errors.New("article has been permanently deleted")
)

// ToHTTPStatus maps common domain errors to their standard HTTP response codes.
func ToHTTPStatus(err error) (int, string) {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND"
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized, "UNAUTHORIZED"
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden, "FORBIDDEN"
	case errors.Is(err, ErrConflict):
		return http.StatusConflict, "CONFLICT"
	case errors.Is(err, ErrBadRequest):
		return http.StatusBadRequest, "BAD_REQUEST"
	case errors.Is(err, ErrArticleGone):
		return http.StatusGone, "ARTICLE_GONE"
	default:
		return http.StatusInternalServerError, "INTERNAL_SERVER_ERROR"
	}
}
