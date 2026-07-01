package capability

import (
	"products/internal/capability/db"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, dbConn db.DBTX) {
	capHandler := newHandler(dbConn)
	r.Get("/products/{id}/capabilities", capHandler.GetCapabilitiesByProduct)
	r.Get("/flows/{id}/capabilities", capHandler.GetCapabilitiesByFlow)
	r.Route("/capabilities", func(u chi.Router) {
		u.Post("/", capHandler.CreateCapability)
		u.Get("/{id}", capHandler.GetCapability)
	})
}
