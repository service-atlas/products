package router

import (
	"log/slog"
	"net/http"
	"products/internal"
	"products/internal/capability"
	"products/internal/db"
	"products/internal/flow"
	"products/internal/platform"
	"products/internal/product"
	"products/internal/system"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func SetupRouter(dbConn db.DBTX) http.Handler {
	slog.Debug("Setting up router")
	router := chi.NewRouter()

	router.Use(internal.RequestIDLogger)
	router.Use(internal.WebRequestLogger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(5))
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	registerSystemCallHandler(router)
	platform.RegisterRoutes(router, dbConn)
	product.RegisterRoutes(router, dbConn)
	flow.RegisterRoutes(router, dbConn)
	capability.RegisterRoutes(router, dbConn)

	slog.Debug("Router setup complete")
	return router
}

func registerSystemCallHandler(r *chi.Mux) {
	h := system.NewSystemCallHandler()
	r.Get("/api/time", h.GetTime)
	r.Get("/api/version", h.GetVersion)
}
