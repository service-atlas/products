package platform

import (
	"context"
	"products/internal/platform/db"

	"github.com/go-chi/chi/v5"
)

type platformService interface {
	CreatePlatform(ctx context.Context, req createPlatformRequest) (db.Platform, error)
	GetPlatform(ctx context.Context, id int) (db.Platform, error)
	GetPlatforms(ctx context.Context) ([]db.Platform, error)
	UpdatePlatform(ctx context.Context, req updatePlatformRequest, id int) (int, error)
	DeletePlatform(ctx context.Context, id int) (int, error)
}

func RegisterRoutes(r chi.Router, dbConn db.DBTX) {
	s := newPostgresService(dbConn)
	h := newHandler(s)
	registerRoutesWithHandler(r, h)
}

func registerRoutesWithHandler(r chi.Router, platformHandler *handler) {
	r.Route("/platforms", func(u chi.Router) {
		u.Post("/", platformHandler.CreatePlatform)
		u.Get("/", platformHandler.GetPlatforms)
		u.Route("/{id}", func(u chi.Router) {
			u.Get("/", platformHandler.GetPlatform)
			u.Put("/", platformHandler.UpdatePlatform)
			u.Delete("/", platformHandler.DeletePlatform)
		})
	})
}
