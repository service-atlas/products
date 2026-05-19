package platform

import (
	"context"
	"products/internal/platform/db"
)

type platformService interface {
	CreatePlatform(ctx context.Context, req createPlatformRequest) (db.Platform, error)
	GetPlatform(ctx context.Context, id int) (db.Platform, error)
	GetPlatforms(ctx context.Context) ([]db.Platform, error)
	UpdatePlatform(ctx context.Context, req updatePlatformRequest, id int) (int, error)
	DeletePlatform(ctx context.Context, id int) (int, error)
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
	//TODO implement me
	panic("implement me")
}

func (p postgresService) GetPlatforms(ctx context.Context) ([]db.Platform, error) {
	//TODO implement me
	panic("implement me")
}

func (p postgresService) UpdatePlatform(ctx context.Context, req updatePlatformRequest, id int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (p postgresService) DeletePlatform(ctx context.Context, id int) (int, error) {
	//TODO implement me
	panic("implement me")
}
