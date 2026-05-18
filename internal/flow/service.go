package flow

import (
	"context"
	"errors"
	"fmt"
	"products/internal"
	"products/internal/flow/db"

	"github.com/jackc/pgx/v5"
)

type service struct {
	queries db.Querier
}

func (s *service) CreateFlow(ctx context.Context, req createFlowRequest, id int) (db.Flow, error) {

	flow, err := s.queries.CreateFlow(ctx, req.ToParams(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Flow{}, internal.NewNotFoundError(id, "Product")
		}
		return db.Flow{}, fmt.Errorf("failed to create flow: %w", err)
	}
	return flow, nil
}

func (s *service) GetFlowById(ctx context.Context, id int) (db.Flow, error) {
	flow, err := s.queries.GetFlow(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Flow{}, internal.NewNotFoundError(id, "Flow")
		}
		return db.Flow{}, fmt.Errorf("failed to fetch flow: %w", err)
	}
	return flow, nil
}

func (s *service) GetFlowsByProduct(ctx context.Context, id int) ([]db.Flow, error) {
	_, err := s.queries.GetProductById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, internal.NewNotFoundError(id, "Product")
		}
		return nil, fmt.Errorf("failed to fetch flows: %w", err)
	}
	flows, err := s.queries.GetFlowsByProduct(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch flows: %w", err)
	}
	return flows, nil
}
