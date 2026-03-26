package apierrors

import (
	"errors"
	"net/http"
)

var (
	ErrUpstreamTimeout = errors.New("upstream timeout")
	ErrUpstreamBadData = errors.New("upstream returned invalid data")
	ErrCacheMiss       = errors.New("cache miss")
	ErrFlightNotFound  = errors.New("flight not found")
)

func StatusFor(err error) int {
	switch {
	case err == nil:
		return http.StatusOK
	case errors.Is(err, ErrFlightNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrUpstreamTimeout):
		return http.StatusGatewayTimeout
	case errors.Is(err, ErrUpstreamBadData):
		return http.StatusBadGateway
	default:
		return http.StatusServiceUnavailable
	}
}
