package capability

import (
	"context"
	"products/internal"
	"products/internal/capability/db"
)

func (s *postgresService) CreateCapabilityStep(ctx context.Context, step createCapabilityStepRequest) (db.CapabilityStep, error) {
	_, err := s.GetCapability(ctx, step.CapabilityId)
	if err != nil {
		return db.CapabilityStep{}, internal.NewNotFoundError(step.CapabilityId, "capability_step")
	}
	capStep, err := s.queries.CreateCapabilityStep(ctx, step.ToParams())
	if err != nil {
		return db.CapabilityStep{}, err
	}
	return capStep, nil
}
