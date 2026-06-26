package capability

import (
	"context"
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
	newCap, err := s.queries.CreateCapability(ctx, req.ToParams())
	if err != nil {
		return db.Capability{}, err
	}
	return newCap, nil
}
