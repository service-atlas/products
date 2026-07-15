package capability

import (
	"context"
	"errors"
	"log/slog"
	"products/internal"
	"products/internal/capability/db"

	"github.com/jackc/pgx/v5"
)

func (s *postgresService) CreateCapability(ctx context.Context, req createCapabilityRequest) (db.Capability, error) {
	name, err := s.queries.GetProduct(ctx, req.ProductId)
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
			return db.Capability{}, internal.NewNotFoundError(id, "capability")
		}
		return db.Capability{}, err
	}
	return capability, nil
}

func (s *postgresService) GetCapabilitiesByProduct(ctx context.Context, id int) ([]db.Capability, error) {
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
	if capabilities == nil {
		capabilities = []db.Capability{}
	}
	return capabilities, nil
}

func (s *postgresService) GetCapabilitiesByFlow(ctx context.Context, id int) ([]db.GetCapabilitiesByFlowRow, error) {
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
	if capabilities == nil {
		capabilities = []db.GetCapabilitiesByFlowRow{}
	}
	return capabilities, nil
}

func (s *postgresService) UpdateCapability(ctx context.Context, req updateCapabilityRequest) (db.Capability, error) {
	rowsAffected, err := s.queries.UpdateCapability(ctx, req.ToParams())
	if err != nil {
		return db.Capability{}, err
	}
	if rowsAffected == 0 {
		return db.Capability{}, NotFoundError{Msg: "capability not found"}
	}
	capability, err := s.queries.GetCapability(ctx, req.Id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Capability{}, NotFoundError{Msg: "capability not found"}
		}
		return db.Capability{}, err
	}
	return capability, nil
}

func (s *postgresService) DeleteCapability(ctx context.Context, id int) error {
	rowsAffected, err := s.queries.DeleteCapability(ctx, id)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return internal.NewNotFoundError(id, "capability")
	}
	return nil
}
