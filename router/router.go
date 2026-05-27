package router

import (
	"log/slog"
	"net/http"
	"products/internal"
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
	platformHandler := platform.NewPlatformHandler(dbConn)
	productHandler := product.NewProductHandler(dbConn)
	flowHandler := flow.NewHandler(dbConn)

	router.Route("/api/platforms", func(u chi.Router) {
		u.Post("/", platformHandler.CreatePlatform)
		u.Get("/", platformHandler.GetPlatforms)
		u.Route("/{id}", func(u chi.Router) {
			u.Get("/", platformHandler.GetPlatform)
			u.Delete("/", platformHandler.DeletePlatform)
			u.Put("/", platformHandler.UpdatePlatform)
			u.Get("/products", productHandler.GetProductsByPlatform)
		})

	})
	router.Route("/api/products", func(u chi.Router) {
		u.Post("/", productHandler.CreateProduct)
		u.Route("/{id}", func(u chi.Router) {
			u.Get("/", productHandler.GetProductById)
			u.Delete("/", productHandler.DeleteProduct)
			u.Put("/", productHandler.UpdateProduct)
			u.Post("/flows", flowHandler.CreateFlow)
			u.Get("/flows", flowHandler.GetFlowsByProduct)
		})

	})
	router.Route("/api/flows", func(u chi.Router) {
		u.Route("/{id}", func(u chi.Router) {
			u.Post("/steps", flowHandler.CreateFlowStep)
			u.Get("/", flowHandler.GetFlowById)
			u.Put("/", flowHandler.UpdateFlow)
			u.Delete("/", flowHandler.DeleteFlow)
		})
	})
	slog.Debug("Router setup complete")
	return router
}

func registerSystemCallHandler(r *chi.Mux) {
	h := system.NewSystemCallHandler()
	r.Get("/api/time", h.GetTime)
	r.Get("/api/version", h.GetVersion)
}
