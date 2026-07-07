package capability

import (
	"context"
	"products/internal/capability/db"

	"github.com/go-chi/chi/v5"
)

type capabilityService interface {
	CreateCapability(ctx context.Context, req createCapabilityRequest) (db.Capability, error)
	GetCapability(ctx context.Context, id int) (db.Capability, error)
	GetCapabilitiesByProduct(ctx context.Context, id int) ([]db.GetCapabilitiesByProductRow, error)
	GetCapabilitiesByFlow(ctx context.Context, id int) ([]db.Capability, error)
	UpdateCapability(ctx context.Context, req updateCapabilityRequest) (db.Capability, error)
}

func RegisterRoutes(r chi.Router, dbConn db.DBTX) {
	s := newPostgresService(dbConn)
	capHandler := newHandler(s)
	registerRoutesWithHandler(r, capHandler)
}

func registerRoutesWithHandler(r chi.Router, capHandler *handler) {
	r.Get("/products/{id}/capabilities", capHandler.GetCapabilitiesByProduct)
	r.Get("/flows/{id}/capabilities", capHandler.GetCapabilitiesByFlow)
	r.Route("/capabilities", func(u chi.Router) {
		u.Post("/", capHandler.CreateCapability)
		u.Get("/{id}", capHandler.GetCapability)
		u.Put("/{id}", capHandler.UpdateCapability)
	})
}
