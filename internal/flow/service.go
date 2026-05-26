package flow

import (
	"context"
	"errors"
	"fmt"
	"products/internal"
	"products/internal/flow/db"

	"github.com/jackc/pgx/v5"
)

type flowService interface {
	CreateFlow(ctx context.Context, req createFlowRequest, id int) (db.Flow, error)
	GetFlowById(ctx context.Context, id int) (db.Flow, error)
	GetFlowsByProduct(ctx context.Context, id int) ([]db.Flow, error)
	UpdateFlow(ctx context.Context, req updateFlowRequest, id int) (db.Flow, error)
	DeleteFlow(ctx context.Context, id int) error
	CreateFlowStep(ctx context.Context, req createFlowStepRequest) (db.FlowStep, error)
}

type postgresService struct {
	queries db.Querier
}

func (s *postgresService) CreateFlowStep(ctx context.Context, req createFlowStepRequest) (db.FlowStep, error) {
	_, err := s.GetFlowById(ctx, req.FlowId)
	if err != nil {
		return db.FlowStep{}, err
	}
	params, err := req.ToParams()
	if err != nil {
		return db.FlowStep{}, err
	}
	flowStep, err := s.queries.CreateFlowStep(ctx, params)
	if err != nil {
		return db.FlowStep{}, fmt.Errorf("failed to create flow step: %w", err)
	}
	return flowStep, nil
}

func (s *postgresService) CreateFlow(ctx context.Context, req createFlowRequest, id int) (db.Flow, error) {

	flow, err := s.queries.CreateFlow(ctx, req.ToParams(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Flow{}, internal.NewNotFoundError(id, "Product")
		}
		return db.Flow{}, fmt.Errorf("failed to create flow: %w", err)
	}
	return flow, nil
}

func (s *postgresService) GetFlowById(ctx context.Context, id int) (db.Flow, error) {
	flow, err := s.queries.GetFlow(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Flow{}, internal.NewNotFoundError(id, "Flow")
		}
		return db.Flow{}, fmt.Errorf("failed to fetch flow: %w", err)
	}
	return flow, nil
}

func (s *postgresService) GetFlowsByProduct(ctx context.Context, id int) ([]db.Flow, error) {
	_, err := s.queries.GetProductById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, internal.NewNotFoundError(id, "Product")
		}
		return nil, fmt.Errorf("failed to fetch flows: %w", err)
	}
	flows, err := s.queries.GetFlowsByProduct(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []db.Flow{}, nil
		}
		return nil, fmt.Errorf("failed to fetch flows: %w", err)
	}
	if flows == nil {
		flows = []db.Flow{}
	}
	return flows, nil
}

func (s *postgresService) UpdateFlow(ctx context.Context, req updateFlowRequest, id int) (db.Flow, error) {
	existing, err := s.GetFlowById(ctx, id)
	if err != nil {
		return db.Flow{}, err
	}

	rowsAffected, err := s.queries.UpdateFlow(ctx, req.ToParams(id, existing))
	if err != nil {
		return db.Flow{}, fmt.Errorf("failed to update flow: %w", err)
	}
	if rowsAffected == 0 {
		return db.Flow{}, internal.NewNotFoundError(id, "Flow")
	}

	return s.GetFlowById(ctx, id)
}
func (s *postgresService) DeleteFlow(ctx context.Context, id int) error {
	rowsAffected, err := s.queries.DeleteFlow(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete flow: %w", err)
	}
	if rowsAffected == 0 {
		return internal.NewNotFoundError(id, "Flow")
	}
	return nil
}
