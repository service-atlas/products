package router

import (
	"log/slog"
	"net/http"
	platformHandler "products/api/platform"
	systemHandler "products/api/system"
	"products/internal"
	"products/internal/db"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func SetupRouter(dbConn db.DBTX) http.Handler {
	slog.Debug("Setting up router")
	router := chi.NewRouter()

	queries := db.New(dbConn)

	router.Use(internal.StructuredLogger(slog.Default()))
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(5))
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	registerSystemCallHandler(router)
	registerPlatformCallHandler(queries, router)
	slog.Debug("Router setup complete")
	return router
}

func registerSystemCallHandler(r *chi.Mux) {
	h := systemHandler.SystemCallHandler{}
	r.Get("/api/time", h.GetTime)
}

func registerPlatformCallHandler(q db.Querier, r *chi.Mux) {
	handler := platformHandler.NewPlatformHandler(q)
	r.Route("/api/platforms", func(u chi.Router) {
		u.Post("/", handler.CreatePlatform)
		u.Get("/", handler.GetPlatforms)
	})
}
