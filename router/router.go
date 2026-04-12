package router

import (
	sampleHandler "chi-boilerplate/api/sample"
	systemHandler "chi-boilerplate/api/system"
	"chi-boilerplate/internal"
	sampleRepo "chi-boilerplate/repo/sample"
	"chi-boilerplate/service/sample"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"log/slog"
	"net/http"
)

func SetupRouter() http.Handler {
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
	db := getDb()
	registerSampleCallHandler(db, router)
	slog.Debug("Router setup complete")
	return router
}

func getDb() string {
	return "sample-db"
}

func registerSystemCallHandler(r *chi.Mux) {
	h := systemHandler.SystemCallHandler{}
	r.Get("/api/time", h.GetTime)
}

func registerSampleCallHandler(s string, r *chi.Mux) {
	repo := sampleRepo.NewSampleRepo(s)
	service := sampleService.NewSampleService(repo)
	handler := sampleHandler.NewSampleCallHandler(service)
	r.Route("/api/sample", func(u chi.Router) {
		u.Get("/", handler.GetSample)
		u.Get("/error", handler.GetError)

	})
}
