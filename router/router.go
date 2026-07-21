package router

import (
	"log/slog"
	"net/http"
	"products/internal/capability"
	"products/internal/db"
	"products/internal/flow"
	"products/internal/platform"
	"products/internal/product"
	"products/internal/system"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/service-atlas/go-common/corsconfig"
	"github.com/service-atlas/go-common/httphelpers"
	"github.com/service-atlas/go-common/httplog"
)

func InitializeRouter(dbConn db.DBTX) http.Handler {
	slog.Debug("Setting up router")
	router := chi.NewRouter()
	// Set the path value look function to chi.UrlParam in case http ever fails
	httphelpers.SetPathValueLookup(chi.URLParam)

	router.Use(httplog.WebRequestLogger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(5))
	corsCfg := corsconfig.GetCORSConfig()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   corsCfg.AllowedOrigins,
		AllowedMethods:   corsCfg.AllowedMethods,
		AllowedHeaders:   corsCfg.AllowedHeaders,
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
