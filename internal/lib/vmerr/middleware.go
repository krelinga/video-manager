package vmerr

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/krelinga/video-manager-api/go/vmapi"
)

func Middleware(w http.ResponseWriter, r *http.Request, err error) {
	var httpErr *httpError
	if !errors.As(err, &httpErr) {
		httpErr = &httpError{
			statusCode: 500,
			wrapped:    fmt.Errorf("unhandled internal server error: %w", err),
		}
	}

	// Much of this was copied & pasted from http.Error().
	h := w.Header()
	h.Del("Content-Length")
	h.Set("Content-Type", "text/json; charset=utf-8")
	h.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(httpErr.statusCode)

	errJson := vmapi.ErrorResponse{
		Message: httpErr.Error(),
	}
	if err := json.NewEncoder(w).Encode(errJson); err != nil {
		http.Error(w, "Failed to encode error response", http.StatusInternalServerError)
		return
	}
}
