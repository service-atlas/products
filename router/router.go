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

type productRoutes struct {
	productHandler  product.Handler
	platformHandler platform.Handler
	flowHandler     flow.Handler
}

func (h *productRoutes) setupRoutes(router *chi.Mux) {
	router.Route("/platforms", func(u chi.Router) {
		u.Post("/", h.platformHandler.CreatePlatform)
		u.Get("/", h.platformHandler.GetPlatforms)
		u.Route("/{id}", func(u chi.Router) {
			u.Get("/", h.platformHandler.GetPlatform)
			u.Delete("/", h.platformHandler.DeletePlatform)
			u.Put("/", h.platformHandler.UpdatePlatform)
			u.Get("/products", h.productHandler.GetProductsByPlatform)
		})

	})
	router.Route("/products", func(u chi.Router) {
		u.Post("/", h.productHandler.CreateProduct)
		u.Route("/{id}", func(u chi.Router) {
			u.Get("/", h.productHandler.GetProductById)
			u.Delete("/", h.productHandler.DeleteProduct)
			u.Put("/", h.productHandler.UpdateProduct)
			u.Post("/flows", h.flowHandler.CreateFlow)
			u.Get("/flows", h.flowHandler.GetFlowsByProduct)
		})

	})
	router.Route("/flows", func(u chi.Router) {
		u.Route("/{id}", func(u chi.Router) {
			u.Post("/steps", h.flowHandler.CreateFlowStep)
			u.Get("/steps", h.flowHandler.GetFlowSteps)
			u.Get("/path", h.flowHandler.GetFlowPath)
			u.Get("/", h.flowHandler.GetFlowById)
			u.Put("/", h.flowHandler.UpdateFlow)
			u.Delete("/", h.flowHandler.DeleteFlow)
		})
	})

	router.Route("/flow-steps/{id}", func(u chi.Router) {
		u.Delete("/", h.flowHandler.DeleteFlowStep)
		u.Patch("/", h.flowHandler.UpdateFlowStep)
	})
}

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
	prodRouter := &productRoutes{
		productHandler:  product.NewProductHandler(dbConn),
		flowHandler:     flow.NewHandler(dbConn),
		platformHandler: platform.NewPlatformHandler(dbConn),
	}
	prodRouter.setupRoutes(router)

	slog.Debug("Router setup complete")
	return router
}

func setupRoutes(router *chi.Mux) {

}

func registerSystemCallHandler(r *chi.Mux) {
	h := system.NewSystemCallHandler()
	r.Get("/api/time", h.GetTime)
	r.Get("/api/version", h.GetVersion)
}
