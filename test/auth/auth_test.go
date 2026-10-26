package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang-restful-api/auth"
	"golang-restful-api/config"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

const testKey = "0123456789abcdef0123456789abcdef" // 32 chars

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestAPIKeyAuth(t *testing.T) {
	mw := auth.APIKeyAuth(testKey)
	ok := func(r *http.Request) *httptest.ResponseRecorder {
		rr := httptest.NewRecorder()
		mw(okHandler()).ServeHTTP(rr, r)
		return rr
	}

	cases := []struct {
		name   string
		header func(*http.Request)
		want   int
	}{
		{"valid bearer", func(r *http.Request) { r.Header.Set("Authorization", "Bearer "+testKey) }, http.StatusOK},
		{"lowercase bearer", func(r *http.Request) { r.Header.Set("Authorization", "bearer "+testKey) }, http.StatusOK},
		{"valid x-admin-key", func(r *http.Request) { r.Header.Set("X-Admin-Key", testKey) }, http.StatusOK},
		{"missing credentials", func(r *http.Request) {}, http.StatusUnauthorized},
		{"wrong bearer", func(r *http.Request) { r.Header.Set("Authorization", "Bearer wrong") }, http.StatusUnauthorized},
		{"wrong x-admin-key", func(r *http.Request) { r.Header.Set("X-Admin-Key", "wrong") }, http.StatusUnauthorized},
		{"non-bearer scheme", func(r *http.Request) { r.Header.Set("Authorization", "Basic "+testKey) }, http.StatusUnauthorized},
		{"bearer takes precedence", func(r *http.Request) {
			r.Header.Set("Authorization", "Bearer wrong")
			r.Header.Set("X-Admin-Key", testKey)
		}, http.StatusUnauthorized},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/products", nil)
			tc.header(r)
			rr := ok(r)
			assert.Equal(t, tc.want, rr.Code)
			if tc.want == http.StatusUnauthorized {
				assert.JSONEq(t, `{"error":"Invalid credentials"}`, rr.Body.String())
			}
		})
	}
}

func TestValidateAuth(t *testing.T) {
	assert.Error(t, config.Config{}.ValidateAuth())
	assert.Error(t, config.Config{AdminAPIKey: "short-key-31-chars-1234567890"}.ValidateAuth()) // 31 chars
	assert.NoError(t, config.Config{AdminAPIKey: testKey}.ValidateAuth())
}

func TestAPIKeyAuth_wrappedRoute(t *testing.T) {
	r := chi.NewRouter()
	r.With(auth.APIKeyAuth(testKey)).Post("/products", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	// No key -> 401.
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/products", nil))
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.JSONEq(t, `{"error":"Invalid credentials"}`, rr.Body.String())

	// Valid key -> handler runs.
	req := httptest.NewRequest(http.MethodPost, "/products", nil)
	req.Header.Set("X-Admin-Key", testKey)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)
}
