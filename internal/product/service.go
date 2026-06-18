package product

import (
	"context"
	"errors"
	"products/internal"
	"products/internal/product/db"

	"github.com/jackc/pgx/v5"
)

type productService interface {
	CreateProduct(ctx context.Context, req createProductRequest) (db.Product, error)
	GetProductsByPlatform(ctx context.Context, platformID int) ([]db.Product, error)
	GetProductById(ctx context.Context, id int) (db.Product, error)
	UpdateProduct(ctx context.Context, req updateProductRequest, id int) (int, error)
	DeleteProduct(ctx context.Context, id int) (int, error)
}

type postgresService struct {
	queries db.Querier
}

func (s *postgresService) CreateProduct(ctx context.Context, req createProductRequest) (db.Product, error) {
	return s.queries.CreateProduct(ctx, req.ToParams())
}

func (s *postgresService) GetProductsByPlatform(ctx context.Context, platformID int) ([]db.Product, error) {
	return s.queries.GetProductsByPlatform(ctx, platformID)
}

func (s *postgresService) GetProductById(ctx context.Context, id int) (db.Product, error) {
	product, err := s.queries.GetProductById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Product{}, internal.NewNotFoundError(id, "Product")
		}
		return db.Product{}, err
	}
	return product, nil
}

func (s *postgresService) UpdateProduct(ctx context.Context, req updateProductRequest, id int) (int, error) {
	rows, err := s.queries.UpdateProduct(ctx, req.ToParams(id))
	if err != nil {
		return 0, err
	}
	if rows == 0 {
		return 0, internal.NewNotFoundError(id, "Product")
	}
	return rows, nil
}

func (s *postgresService) DeleteProduct(ctx context.Context, id int) (int, error) {
	rows, err := s.queries.DeleteProduct(ctx, id)
	if err != nil {
		return 0, err
	}
	if rows == 0 {
		return 0, internal.NewNotFoundError(id, "Product")
	}
	return rows, nil
}
