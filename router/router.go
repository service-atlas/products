package router

import (
	"log/slog"
	"net/http"
	platformHandler "products/api/platform"
	productHandler "products/api/product"
	systemHandler "products/api/system"
	"products/internal"
	"products/internal/db"
	platformDb "products/internal/db/platform"
	productDb "products/internal/db/product"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func SetupRouter(dbConn db.DBTX) http.Handler {
	slog.Debug("Setting up router")
	router := chi.NewRouter()

	store := db.New(dbConn)

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
	registerPlatformCallHandler(store.Platform, router)
	registerProductCallHandler(store.Product, router)
	slog.Debug("Router setup complete")
	return router
}

func registerSystemCallHandler(r *chi.Mux) {
	h := systemHandler.SystemCallHandler{}
	r.Get("/api/time", h.GetTime)
}

func registerPlatformCallHandler(q platformDb.Querier, r *chi.Mux) {
	handler := platformHandler.NewPlatformHandler(q)
	r.Route("/api/platforms", func(u chi.Router) {
		u.Post("/", handler.CreatePlatform)
		u.Get("/", handler.GetPlatforms)
		u.Get("/{id}", handler.GetPlatform)
		u.Delete("/{id}", handler.DeletePlatform)
		u.Put("/{id}", handler.UpdatePlatform)
	})
}

func registerProductCallHandler(q productDb.Querier, r *chi.Mux) {
	handler := productHandler.NewProductHandler(q)
	r.Route("/api/products", func(u chi.Router) {
		u.Post("/", handler.CreateProduct)
	})
}
