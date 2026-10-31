package config_test

import (
	"testing"

	"golang-restful-api/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	for _, key := range []string{
		"APP_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_HOST",
		"DB_PORT",
		"DB_NAME",
		"AUTO_MIGRATE",
		"CORS_ALLOWED_ORIGINS",
		"CORS_ALLOWED_METHODS",
		"CORS_ALLOWED_HEADERS",
		"CORS_MAX_AGE",
		"ADMIN_API_KEY",
	} {
		t.Setenv(key, "")
	}

	cfg := config.Load()

	assert.Equal(t, "6767", cfg.AppPort)
	assert.Equal(t, "root", cfg.DBUser)
	assert.Empty(t, cfg.DBPassword)
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, "3306", cfg.DBPort)
	assert.Equal(t, "products", cfg.DBName)
	assert.False(t, cfg.AutoMigrate)
	assert.Equal(t, []string{"*"}, cfg.CORSAllowedOrigins)
	assert.Equal(t, []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, cfg.CORSAllowedMethods)
	assert.Equal(t, []string{"Accept", "Content-Type", "Authorization", "X-Request-Id", "X-Admin-Key"}, cfg.CORSAllowedHeaders)
	assert.Equal(t, 300, cfg.CORSMaxAge)
	assert.Empty(t, cfg.AdminAPIKey)
}

func TestLoadEnvironmentOverrides(t *testing.T) {
	t.Setenv("APP_PORT", "8080")
	t.Setenv("DB_USER", "app")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_HOST", "db.internal")
	t.Setenv("DB_PORT", "3307")
	t.Setenv("DB_NAME", "catalog")
	t.Setenv("AUTO_MIGRATE", "true")
	t.Setenv("CORS_ALLOWED_ORIGINS", " https://shop.example, https://admin.example ")
	t.Setenv("CORS_ALLOWED_METHODS", "GET, PATCH")
	t.Setenv("CORS_ALLOWED_HEADERS", "Authorization, X-Trace-Id")
	t.Setenv("CORS_MAX_AGE", "600")
	t.Setenv("ADMIN_API_KEY", "0123456789abcdef0123456789abcdef")

	cfg := config.Load()

	assert.Equal(t, "8080", cfg.AppPort)
	assert.Equal(t, "app", cfg.DBUser)
	assert.Equal(t, "secret", cfg.DBPassword)
	assert.Equal(t, "db.internal", cfg.DBHost)
	assert.Equal(t, "3307", cfg.DBPort)
	assert.Equal(t, "catalog", cfg.DBName)
	assert.True(t, cfg.AutoMigrate)
	assert.Equal(t, []string{"https://shop.example", "https://admin.example"}, cfg.CORSAllowedOrigins)
	assert.Equal(t, []string{"GET", "PATCH"}, cfg.CORSAllowedMethods)
	assert.Equal(t, []string{"Authorization", "X-Trace-Id"}, cfg.CORSAllowedHeaders)
	assert.Equal(t, 600, cfg.CORSMaxAge)
	assert.Equal(t, "0123456789abcdef0123456789abcdef", cfg.AdminAPIKey)
}

func TestLoadInvalidValuesUseDefaults(t *testing.T) {
	t.Setenv("AUTO_MIGRATE", "sometimes")
	t.Setenv("CORS_MAX_AGE", "0")
	t.Setenv("CORS_ALLOWED_ORIGINS", " , ")

	cfg := config.Load()

	assert.False(t, cfg.AutoMigrate)
	assert.Equal(t, 300, cfg.CORSMaxAge)
	assert.Equal(t, []string{"*"}, cfg.CORSAllowedOrigins)
}

func TestConfigDSN(t *testing.T) {
	cfg := config.Config{
		DBUser:     "app",
		DBPassword: "secret",
		DBHost:     "db.internal",
		DBPort:     "3307",
		DBName:     "catalog",
	}

	assert.Equal(t, "app:secret@tcp(db.internal:3307)/catalog?charset=utf8mb4&parseTime=True&loc=Local", cfg.DSN())
}

func TestValidateAuth(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		wantErr string
	}{
		{name: "missing", wantErr: "ADMIN_API_KEY is required"},
		{name: "too short", apiKey: "0123456789abcdef0123456789abcde", wantErr: "ADMIN_API_KEY must be at least 32 characters"},
		{name: "valid", apiKey: "0123456789abcdef0123456789abcdef"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := (config.Config{AdminAPIKey: tt.apiKey}).ValidateAuth()
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.EqualError(t, err, tt.wantErr)
		})
	}
}
