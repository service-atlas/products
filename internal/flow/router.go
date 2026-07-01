package flow

import (
	"context"
	"products/internal/flow/db"

	"github.com/go-chi/chi/v5"
)

type flowService interface {
	CreateFlow(ctx context.Context, req createFlowRequest, id int) (db.Flow, error)
	GetFlowById(ctx context.Context, id int) (db.Flow, error)
	GetFlowsByProduct(ctx context.Context, id int) ([]db.Flow, error)
	UpdateFlow(ctx context.Context, req updateFlowRequest, id int) (db.Flow, error)
	DeleteFlow(ctx context.Context, id int) error

	CreateFlowStep(ctx context.Context, req createFlowStepRequest) (db.FlowStep, error)
	DeleteFlowStep(ctx context.Context, id int) error
	GetFlowSteps(ctx context.Context, id int) ([]db.FlowStep, error)
	GetFlowPath(ctx context.Context, id int) (FlowPath, error)
}

func RegisterRoutes(r chi.Router, dbConn db.DBTX) {
	s := newPostgresService(dbConn)
	flowHandler := newHandler(s)
	registerRoutesWithHandler(r, flowHandler)
}

func registerRoutesWithHandler(r chi.Router, flowHandler *handler) {
	r.Route("/flows/{id}", func(u chi.Router) {
		u.Post("/steps", flowHandler.CreateFlowStep)
		u.Get("/steps", flowHandler.GetFlowSteps)
		u.Get("/path", flowHandler.GetFlowPath)
		u.Get("/", flowHandler.GetFlowById)
		u.Put("/", flowHandler.UpdateFlow)
		u.Delete("/", flowHandler.DeleteFlow)
	})

	r.Delete("/flow-steps/{id}", flowHandler.DeleteFlowStep)
	r.Post("/products/{id}/flows", flowHandler.CreateFlow)
	r.Get("/products/{id}/flows", flowHandler.GetFlowsByProduct)
}
