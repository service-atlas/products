package capability

import (
	"context"
	"errors"
	"log/slog"
	"products/internal/capability/db"

	"github.com/jackc/pgx/v5"
)

type capabilityService interface {
	capabilityStepService
	CreateCapability(ctx context.Context, req createCapabilityRequest) (db.Capability, error)
	GetCapability(ctx context.Context, id int) (db.Capability, error)
	GetCapabilitiesByProduct(ctx context.Context, id int) ([]db.GetCapabilitiesByProductRow, error)
	GetCapabilitiesByFlow(ctx context.Context, id int) ([]db.Capability, error)
}

type capabilityStepService interface {
}

type postgresService struct {
	queries db.Querier
}

func (s *postgresService) CreateCapability(ctx context.Context, req createCapabilityRequest) (db.Capability, error) {
	name, err := s.queries.GetFlow(ctx, req.FlowId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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

func (s *postgresService) GetCapability(ctx context.Context, id int) (db.Capability, error) {
	capability, err := s.queries.GetCapability(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Capability{}, NotFoundError{
				"Capability not found",
			}
		}
		return db.Capability{}, err
	}
	return capability, nil
}

func (s *postgresService) GetCapabilitiesByProduct(ctx context.Context, id int) ([]db.GetCapabilitiesByProductRow, error) {
	_, err := s.queries.GetProduct(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, NotFoundError{
				"Product not found",
			}
		}
		return nil, err
	}
	capabilities, err := s.queries.GetCapabilitiesByProduct(ctx, id)
	if err != nil {
		return nil, err
	}
	return capabilities, nil
}

func (s *postgresService) GetCapabilitiesByFlow(ctx context.Context, id int) ([]db.Capability, error) {
	_, err := s.queries.GetFlow(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, NotFoundError{
				"Flow not found",
			}
		}
		return nil, err
	}
	capabilities, err := s.queries.GetCapabilitiesByFlow(ctx, id)
	if err != nil {
		return nil, err
	}
	return capabilities, nil
}
