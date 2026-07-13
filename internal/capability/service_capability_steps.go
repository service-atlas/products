package capability

import (
	"context"
	"errors"
	"log/slog"
	"products/internal"
	"products/internal/capability/db"
	"sync"

	"github.com/jackc/pgx/v5"
)

func (s *postgresService) CreateCapabilityStep(ctx context.Context, req createCapabilityStepRequest) (db.CapabilityStep, error) {
	errChan := make(chan string, 2)
	wg := new(sync.WaitGroup)
	wg.Go(func() {
		_, err := s.GetCapability(ctx, req.CapabilityId)
		if err != nil {
			if _, ok := errors.AsType[NotFoundError](err); ok || errors.Is(err, pgx.ErrNoRows) {
				errChan <- "capability_not_found"
			} else {
				internal.LoggerFromContext(ctx).Error("error getting capability in create capability", slog.String("error", err.Error()))
				errChan <- "capability_general"
			}
		}
	})

	wg.Go(func() {
		_, err := s.queries.GetFlowStep(ctx, req.FlowStepId)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				errChan <- "flow_step_not_found"
			} else {
				internal.LoggerFromContext(ctx).Error("error getting flow step in create capability", slog.String("error", err.Error()))
				errChan <- "flow_step_general"
			}
		}

	})
	wg.Wait()
	close(errChan)
	for t := range errChan {
		switch t {
		case "flow_step_not_found":
			return db.CapabilityStep{}, internal.NewNotFoundError(req.FlowStepId, t)
		case "flow_step_general":
			return db.CapabilityStep{}, errors.New("error getting flow step")
		case "capability_not_found":
			return db.CapabilityStep{}, internal.NewNotFoundError(req.CapabilityId, t)
		case "capability_general":
			return db.CapabilityStep{}, errors.New("error getting capability")
		}

	}
	capStep, err := s.queries.CreateCapabilityStep(ctx, req.ToParams())
	if err != nil {
		return db.CapabilityStep{}, err
	}
	return capStep, nil
}
