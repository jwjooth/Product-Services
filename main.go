package main

import (
	"context"
	"errors"
	"fmt"
	"golang-restful-api/app"
	"golang-restful-api/docs"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang-restful-api/config"
	"golang-restful-api/controller"
	"golang-restful-api/repository"
	"golang-restful-api/service"

	_ "golang-restful-api/docs"

	"github.com/joho/godotenv"
)

// @title Golang RESTful API
// @version 1.2.0
// @description API documentation for the Golang Product RESTful API
// @BasePath /api/v1
// @securityDefinitions.apikey AdminAPIKey
// @in header
// @name Authorization
// @description Provide the admin API key as a Bearer token.
func main() {
	if err := run(); err != nil {
		log.Printf("application stopped: %v", err)
		os.Exit(1)
	}
}

func run() error {
	// .env is optional: missing file must not crash production.
	_ = godotenv.Load()

	if appURL := os.Getenv("APP_URL"); appURL != "" {
		u, err := url.Parse(appURL)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
			return fmt.Errorf("invalid APP_URL %q", appURL)
		}

		docs.SwaggerInfo.Host = u.Host
		docs.SwaggerInfo.Schemes = []string{u.Scheme}
	}

	cfg := config.Load()
	if err := cfg.ValidateAuth(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	db, err := config.OpenDB(cfg)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer config.CloseDB(db)

	if err := config.Migrate(db, cfg); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	productRepo := repository.NewProductRepositoryImpl(db)
	productService := service.NewProductServiceImpl(productRepo)
	productController := controller.NewProductController(productService)
	handler := app.NewRouter(cfg, productController)

	server := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("server listening on %s", server.Addr)
		serverErr <- server.ListenAndServe()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("listen: %w", err)
	case <-ctx.Done():
	}
	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}
	if err := <-serverErr; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server stopped unexpectedly: %w", err)
	}
	log.Println("server exited")
	return nil
}
