package auth

import (
	"crypto/subtle"
	"golang-restful-api/helper"
	"net/http"
	"strings"
)

// APIKeyAuth rejects requests that do not carry a valid admin API key.
func APIKeyAuth(expected string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r.Header.Get("Authorization"))

			if token == "" {
				token = r.Header.Get("X-Admin-Key")
			}

			if subtle.ConstantTimeCompare([]byte(token), []byte(expected)) != 1 {
				helper.WriteError(w, http.StatusUnauthorized, "Invalid credentials")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func bearerToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}
