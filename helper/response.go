package helper

import (
	"encoding/json"
	"golang-restful-api/payload"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

// SuccessEnvelope is the standard success body.
type SuccessEnvelope struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

type ProductListEnvelope struct {
	Message string                    `json:"message"`
	Data    []payload.ProductResponse `json:"data"`
	Meta    payload.ListProductsMeta  `json:"meta"`
}

// ErrorEnvelope is the standard error body.
type ErrorEnvelope struct {
	Error string `json:"error"`
}

// WriteJSON writes v as JSON with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteSuccess writes a 2xx envelope: {"message":..., "data":..., "meta"?}.
func WriteSuccess(w http.ResponseWriter, status int, message string, data any) {
	WriteJSON(w, status, SuccessEnvelope{Message: message, Data: data})
}

// WriteSuccessWithMeta writes a 2xx envelope including pagination meta.
func WriteSuccessWithMeta(w http.ResponseWriter, status int, message string, data any, meta any) {
	WriteJSON(w, status, SuccessEnvelope{Message: message, Data: data, Meta: meta})
}

// WriteError writes a 4xx/5xx envelope: {"error": message}.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, ErrorEnvelope{Error: message})
}

// WriteAppError maps an AppError (or generic error) to a JSON error body.
func WriteAppError(w http.ResponseWriter, err error) {
	WriteError(w, CodeOf(err), MessageOf(err))
}

// ParseIDParam parses a positive unsigned ID from a route parameter.
func ParseIDParam(r *http.Request, key string) (uint, *AppError) {
	raw := chi.URLParam(r, key)
	n, err := strconv.ParseUint(raw, 10, 32)
	if err != nil || n == 0 {
		return 0, BadRequest("invalid id")
	}
	return uint(n), nil
}

func QueryInt(r *http.Request, key string, fallback int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}

// QueryString returns a trimmed query parameter, or "" when absent.
func QueryString(r *http.Request, key string) string {
	return strings.TrimSpace(r.URL.Query().Get(key))
}

// QueryFloat returns a query parameter parsed as float64. ok is false when
// the parameter is absent or malformed.
func QueryFloat(r *http.Request, key string) (float64, bool) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return 0, false
	}
	n, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

// QueryInt64 returns a query parameter parsed as int64. ok is false when
// the parameter is absent or malformed.
func QueryInt64(r *http.Request, key string) (int64, bool) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}
