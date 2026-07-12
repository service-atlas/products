package capability

import (
	"context"
	"products/internal"
	"products/internal/capability/db"
)

func (s *postgresService) CreateCapabilityStep(ctx context.Context, req createCapabilityStepRequest) (db.CapabilityStep, error) {
	_, err := s.GetCapability(ctx, req.CapabilityId)
	if err != nil {
		return db.CapabilityStep{}, internal.NewNotFoundError(req.CapabilityId, "capability_step")
	}
	capStep, err := s.queries.CreateCapabilityStep(ctx, req.ToParams())
	if err != nil {
		return db.CapabilityStep{}, err
	}
	return capStep, nil
}
