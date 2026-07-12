package capability

import (
	"context"
	"products/internal"
	"products/internal/capability/db"
	"sync"
)

func (s *postgresService) CreateCapabilityStep(ctx context.Context, req createCapabilityStepRequest) (db.CapabilityStep, error) {
	errChan := make(chan string, 2)
	wg := new(sync.WaitGroup)
	wg.Go(func() {
		_, err := s.GetCapability(ctx, req.CapabilityId)
		if err != nil {
			errChan <- "capability"
		}
	})

	wg.Go(func() {
		n, err := s.queries.GetFlowStep(ctx, req.FlowStepId)
		if err != nil || n == 0 {
			errChan <- "flow_step"
		}
	})
	wg.Wait()
	close(errChan)
	for t := range errChan {
		switch t {
		case "flow_step":
			return db.CapabilityStep{}, internal.NewNotFoundError(req.FlowStepId, t)
		default:
			return db.CapabilityStep{}, internal.NewNotFoundError(req.CapabilityId, t)
		}

	}
	capStep, err := s.queries.CreateCapabilityStep(ctx, req.ToParams())
	if err != nil {
		return db.CapabilityStep{}, err
	}
	return capStep, nil
}
