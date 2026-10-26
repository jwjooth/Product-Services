package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"golang-restful-api/entities"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config holds all application configuration loaded from environment.
type Config struct {
	AppPort            string
	DBUser             string
	DBPassword         string
	DBHost             string
	DBPort             string
	DBName             string
	CORSAllowedOrigins []string
	CORSAllowedMethods []string
	CORSAllowedHeaders []string
	CORSMaxAge         int
}

// Load reads configuration from environment variables with sane defaults.
// Missing .env file is NOT fatal; env vars / platform defaults are used.
func Load() Config {
	return Config{
		AppPort:            getEnv("APP_PORT", "6767"),
		DBUser:             getEnv("DB_USER", "root"),
		DBPassword:         getEnv("DB_PASSWORD", ""),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "3306"),
		DBName:             getEnv("DB_NAME", "products"),
		CORSAllowedOrigins: getEnvList("CORS_ALLOWED_ORIGINS", []string{"*"}),
		CORSMaxAge:         getEnvInt("CORS_MAX_AGE", 300),
		CORSAllowedHeaders: getEnvList("CORS_ALLOWED_HEADERS", []string{
			"Accept",
			"Content-Type",
			"Authorization",
			"X-Request-Id",
		}),
		CORSAllowedMethods: getEnvList("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
	}
}

// DSN builds a MySQL DSN from the config.
func (c Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

// OpenDB opens a GORM MySQL connection with production-ready pool settings.
func OpenDB(cfg Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	maxIdle := getEnvInt("DB_MAX_IDLE_CONNS", 10)
	maxOpen := getEnvInt("DB_MAX_OPEN_CONNS", 100)

	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

// Migrate runs AutoMigrate for all entities. Safe for dev; use versioned
// migrations (golang-migrate) for production schema evolution.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&entities.ProductEntity{})
}

// CloseDB closes the underlying sql.DB connection.
func CloseDB(db *gorm.DB) {
	if db == nil {
		return
	}
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}

func getEnvList(key string, fallback []string) []string {
	if v := os.Getenv(key); v != "" {
		var result []string
		for item := range strings.SplitSeq(v, ",") {
			if cleaned := strings.TrimSpace(item); cleaned != "" {
				result = append(result, cleaned)
			}
		}
		if len(result) == 0 {
			return fallback
		}
		return result
	}
	return fallback
}
