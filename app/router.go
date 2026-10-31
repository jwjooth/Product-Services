package app

import (
	"net/http"
	"time"

	"golang-restful-api/auth"
	"golang-restful-api/config"
	"golang-restful-api/controller"
	"golang-restful-api/helper"
	appmw "golang-restful-api/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

// NewRouter assembles the HTTP middleware and routes for the application.
func NewRouter(cfg config.Config, products controller.ProductController) http.Handler {
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
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/products", func(r chi.Router) {
			r.Get("/", products.ListProducts)
			r.Get("/{id}", products.GetProductByID)
			r.With(auth.APIKeyAuth(cfg.AdminAPIKey)).Post("/", products.CreateProduct)
			r.With(auth.APIKeyAuth(cfg.AdminAPIKey)).Put("/{id}", products.UpdateProduct)
			r.With(auth.APIKeyAuth(cfg.AdminAPIKey)).Delete("/{id}", products.DeleteProduct)
		})
	})

	return r
}
