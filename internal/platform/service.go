package platform

import (
	"context"
	"errors"
	"products/internal"
	"products/internal/platform/db"

	"github.com/jackc/pgx/v5"
)

func newPostgresService(dbConn db.DBTX) platformService {
	queries := db.New(dbConn)
	return &postgresService{queries}
}

type postgresService struct {
	db db.Querier
}

func (p postgresService) CreatePlatform(ctx context.Context, req createPlatformRequest) (db.Platform, error) {
	platform, err := p.db.CreatePlatform(ctx, req.ToParams())
	if err != nil {
		return db.Platform{}, err
	}
	return platform, nil
}

func (p postgresService) GetPlatform(ctx context.Context, id int) (db.Platform, error) {
	platform, err := p.db.GetPlatform(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Platform{}, internal.NewNotFoundError(id, "Platform")
		}
		return db.Platform{}, err
	}
	return platform, nil
}

func (p postgresService) GetPlatforms(ctx context.Context) ([]db.Platform, error) {
	return p.db.GetPlatforms(ctx)
}

func (p postgresService) UpdatePlatform(ctx context.Context, req updatePlatformRequest, id int) (int, error) {
	rows, err := p.db.UpdatePlatform(ctx, req.ToParams(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, internal.NewNotFoundError(id, "Platform")
		}
		return 0, err
	}
	if rows == 0 {
		return 0, internal.NewNotFoundError(id, "Platform")
	}
	return rows, nil
}

func (p postgresService) DeletePlatform(ctx context.Context, id int) (int, error) {
	rows, err := p.db.DeletePlatform(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, internal.NewNotFoundError(id, "Platform")
		}
		return 0, err
	}
	if rows == 0 {
		return 0, internal.NewNotFoundError(id, "Platform")
	}
	return rows, nil
}
