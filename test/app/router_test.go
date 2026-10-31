package app_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang-restful-api/app"
	"golang-restful-api/config"
	"golang-restful-api/controller"

	"github.com/stretchr/testify/assert"
)

func TestNewRouterHealthAndAdminRoutes(t *testing.T) {
	cfg := config.Config{
		AdminAPIKey:        "0123456789abcdef0123456789abcdef",
		CORSAllowedOrigins: []string{"*"},
		CORSAllowedMethods: []string{"GET", "POST"},
		CORSAllowedHeaders: []string{"Content-Type", "X-Admin-Key"},
	}
	router := app.NewRouter(cfg, controller.NewProductController(nil))

	health := httptest.NewRecorder()
	router.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	assert.Equal(t, http.StatusOK, health.Code)

	protected := httptest.NewRecorder()
	router.ServeHTTP(protected, httptest.NewRequest(http.MethodPost, "/api/v1/products/", nil))
	assert.Equal(t, http.StatusUnauthorized, protected.Code)
}
