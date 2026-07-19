package capability

import (
	"context"
	"errors"
	"log/slog"
	"products/internal"
	"products/internal/capability/db"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/service-atlas/go-common/httplog"
)

func (s *postgresService) CreateCapabilityStep(ctx context.Context, req createCapabilityStepRequest) (db.CapabilityStep, error) {
	errChan := make(chan error, 2)
	wg := new(sync.WaitGroup)
	wg.Go(func() {
		_, err := s.GetCapability(ctx, req.CapabilityId)
		if err != nil {
			if _, ok := errors.AsType[internal.NotFoundError](err); ok || errors.Is(err, pgx.ErrNoRows) {
				errChan <- internal.NewNotFoundError(req.CapabilityId, "capability")
				return
			}
			httplog.LoggerFromContext(ctx).Error("error getting capability in create capability", slog.String("error", err.Error()))
			errChan <- err
		}

	})

	wg.Go(func() {
		logger := httplog.LoggerFromContext(ctx)
		flowStep, err := s.queries.GetFlowStep(ctx, req.FlowStepId)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				errChan <- internal.NewNotFoundError(req.FlowStepId, "flow_step")
				return
			}
			logger.Error("error getting flow step in create capability", slog.String("error", err.Error()))
			errChan <- err
			return
		}

		flowIds, err := s.queries.GetFlowsFromSteps(ctx, req.CapabilityId)
		if err != nil {
			errChan <- err
			return
		}
		if len(flowIds) > 1 {
			logger.Error("capability steps from database contain multiple flows", slog.Int("capability_id", req.CapabilityId))
			errChan <- errors.New("multiple flow ids found in existing steps")
			return
		}
		if len(flowIds) == 1 && flowIds[0] != flowStep.FlowID {
			errChan <- internal.NewValidationErr("flow step doesn't belong to already bound flow")
		}

	})
	wg.Wait()
	close(errChan)
	for err := range errChan {
		return db.CapabilityStep{}, err
	}
	capStep, err := s.queries.CreateCapabilityStep(ctx, req.ToParams())
	if err != nil {
		return db.CapabilityStep{}, err
	}
	return capStep, nil
}

func (s *postgresService) DeleteCapabilityStep(ctx context.Context, id int) error {
	rows, err := s.queries.DeleteCapabilityStep(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		return internal.NewNotFoundError(id, "capability-step")
	}
	return nil
}

func (s *postgresService) GetCapabilitySteps(ctx context.Context, id int) ([]db.CapabilityStep, error) {
	_, err := s.GetCapability(ctx, id)
	if err != nil {
		return nil, err
	}
	steps, err := s.queries.GetCapabilitySteps(ctx, id)
	if err != nil {
		return nil, err
	}
	if steps == nil {
		steps = []db.CapabilityStep{}
	}
	return steps, nil
}
