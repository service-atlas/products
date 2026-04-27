package platformHandler

import (
	"context"
	"products/internal/db"
)

type PlatformQuerier interface {
	CreatePlatform(ctx context.Context, arg db.CreatePlatformParams) error
	DeletePlatform(ctx context.Context, id int32) (int32, error)
	GetPlatform(ctx context.Context, id int32) (db.Platform, error)
	GetPlatforms(ctx context.Context) ([]db.Platform, error)
	UpdatePlatform(ctx context.Context, arg db.UpdatePlatformParams) error
}
