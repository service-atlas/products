package capability

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"products/internal/capability/db"
)

type capabilityService interface {
	capabilityStepService
	CreateCapability(ctx context.Context, req createCapabilityRequest) (db.Capability, error)
}

type capabilityStepService interface {
}

type postgresService struct {
	queries db.Querier
}

func (s *postgresService) CreateCapability(ctx context.Context, req createCapabilityRequest) (db.Capability, error) {
	name, err := s.queries.GetFlow(ctx, req.FlowId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.Capability{}, NotFoundError{
				"Flow not found",
			}
		}
		return db.Capability{}, err
	}
	if name == "" {
		return db.Capability{}, NotFoundError{
			"Flow not found",
		}
	}
	newCap, err := s.queries.CreateCapability(ctx, req.ToParams())
	if err != nil {
		slog.Error("error creating capability: ", slog.String("error", err.Error()))
		return db.Capability{}, errors.New("error creating capability")
	}
	return newCap, nil
}
