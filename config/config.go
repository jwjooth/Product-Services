package config

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"golang-restful-api/entities"
	"golang-restful-api/migrations"

	"github.com/golang-migrate/migrate/v4"
	sqlMigrate "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
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
	AutoMigrate        bool
	CORSAllowedOrigins []string
	CORSAllowedMethods []string
	CORSAllowedHeaders []string
	CORSMaxAge         int
	AdminAPIKey        string
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
		AutoMigrate:        getEnvBool("AUTO_MIGRATE", false),
		CORSAllowedOrigins: getEnvList("CORS_ALLOWED_ORIGINS", []string{"*"}),
		CORSMaxAge:         getEnvInt("CORS_MAX_AGE", 300),
		CORSAllowedHeaders: getEnvList("CORS_ALLOWED_HEADERS", []string{
			"Accept",
			"Content-Type",
			"Authorization",
			"X-Request-Id",
			"X-Admin-Key",
		}),
		CORSAllowedMethods: getEnvList("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		AdminAPIKey:        getEnv("ADMIN_API_KEY", ""),
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

// Migrate applies schema changes using GORM AutoMigrate or versioned SQL migrations.
// Versioned migrations only run when RUN_MIGRATIONS=true (default: false).
func Migrate(db *gorm.DB, cfg Config) error {
	if cfg.AutoMigrate {
		log.Println("AUTO_MIGRATE is enabled; using GORM AutoMigrate (intended for development only)")
		if err := migratePriceToCentsBeforeAutoMigrate(db); err != nil {
			return fmt.Errorf("price-to-cents migration before automigrate: %w", err)
		}
		return db.AutoMigrate(&entities.ProductEntity{})
	}

	if !getEnvBool("RUN_MIGRATIONS", false) {
		log.Println("RUN_MIGRATIONS is not enabled; skipping versioned migrations")
		return nil
	}

	// Open a dedicated connection for migrations so the shared pool is not
	// affected when golang-migrate closes the database driver.
	migrationDB, err := sql.Open("mysql", cfg.DSN()+"?multiStatements=true")
	if err != nil {
		return fmt.Errorf("open migration connection: %w", err)
	}
	defer migrationDB.Close()

	driver, err := sqlMigrate.WithInstance(migrationDB, &sqlMigrate.Config{})
	if err != nil {
		return err
	}

	sourceDriver, err := iofs.New(migrations.Files, ".")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "mysql", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	version, dirty, verErr := m.Version()
	if verErr != nil {
		log.Printf("Migrations applied but could not read schema version: %v", verErr)
	} else {
		log.Printf("Database schema version: %d (dirty: %v)", version, dirty)
	}
	return nil
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

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if value, err := strconv.ParseBool(v); err == nil {
			return value
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

func (c Config) ValidateAuth() error {
	if c.AdminAPIKey == "" {
		return errors.New("ADMIN_API_KEY is required")
	}
	if len(c.AdminAPIKey) < 32 {
		return errors.New("ADMIN_API_KEY must be at least 32 characters")
	}
	return nil
}

// migratePriceToCentsBeforeAutoMigrate converts existing floating-point/decimal
// prices to integer cents before GORM alters the column to BIGINT, so values
// like 19.99 survive as 1999 instead of being truncated to 19. It is idempotent:
// if the column is already BIGINT nothing is changed; if the table/column does
// not exist yet AutoMigrate will create it.
func migratePriceToCentsBeforeAutoMigrate(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	var dataType string
	err = sqlDB.QueryRow(`
		SELECT DATA_TYPE FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'product' AND COLUMN_NAME = 'price'
	`).Scan(&dataType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if dataType != "bigint" {
		_, err = sqlDB.Exec(`UPDATE product SET price = ROUND(price * 100)`)
		return err
	}
	return nil
}
