package main

import (
	"context"
	"errors"
	"golang-restful-api/auth"
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
	"golang-restful-api/helper"
	"golang-restful-api/repository"
	"golang-restful-api/service"

	_ "golang-restful-api/docs"

	appmw "golang-restful-api/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Golang RESTful API
// @version 1.0.0
// @description API documentation for the Golang Product RESTful API
// @BasePath /api/v1
// @securityDefinitions.apikey AdminAPIKey
// @in header
// @name Authorization
// @description Provide the admin API key as a Bearer token.
func main() {
	// .env is optional: missing file must not crash production.
	_ = godotenv.Load()

	if appURL := os.Getenv("APP_URL"); appURL != "" {
		u, err := url.Parse(appURL)
		if err != nil {
			log.Fatalf("invalid APP_URL: %v", err)
		}

		docs.SwaggerInfo.Host = u.Host
		docs.SwaggerInfo.Schemes = []string{u.Scheme}
	}

	cfg := config.Load()
	if err := cfg.ValidateAuth(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	db, err := config.OpenDB(cfg)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer config.CloseDB(db)

	if err := config.Migrate(db, cfg); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(appmw.NewCORSMiddleware(appmw.CORSConfig{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   cfg.CORSAllowedMethods,
		AllowedHeaders:   cfg.CORSAllowedHeaders,
		ExposedHeaders:   []string{"X-Request-Id"},
		MaxAge:           cfg.CORSMaxAge,
		AllowCredentials: len(cfg.CORSAllowedOrigins) == 1 && cfg.CORSAllowedOrigins[0] != "*",
	}))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		helper.WriteSuccess(w, http.StatusOK, "ok", map[string]string{"status": "up"})
	})

	// Swagger UI: http://localhost:6767/swagger/index.html
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	//product
	productRepo := repository.NewProductRepositoryImpl(db)
	productService := service.NewProductServiceImpl(productRepo)
	productController := controller.NewProductController(productService)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/products", func(r chi.Router) {
			r.Get("/", productController.ListProducts)
			r.Get("/{id}", productController.GetProductByID)
			r.With(auth.APIKeyAuth(cfg.AdminAPIKey)).Post("/", productController.CreateProduct)
			r.With(auth.APIKeyAuth(cfg.AdminAPIKey)).Put("/{id}", productController.UpdateProduct)
			r.With(auth.APIKeyAuth(cfg.AdminAPIKey)).Delete("/{id}", productController.DeleteProduct)
		})
	})

	server := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server listening on %s", server.Addr)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server exited")
}
