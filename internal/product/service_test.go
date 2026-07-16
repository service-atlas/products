package product

import (
	"context"
	"errors"
	"products/internal"
	"products/internal/product/db"
	"testing"

	"github.com/jackc/pgx/v5"
)

type mockQuerier struct {
	createProduct         func(ctx context.Context, arg db.CreateProductParams) (db.Product, error)
	deleteProduct         func(ctx context.Context, id int) (int, error)
	getProductById        func(ctx context.Context, id int) (db.Product, error)
	getProductsByPlatform func(ctx context.Context, platformID int) ([]db.Product, error)
	updateProduct         func(ctx context.Context, arg db.UpdateProductParams) (int, error)
}

func (m *mockQuerier) CreateProduct(ctx context.Context, arg db.CreateProductParams) (db.Product, error) {
	return m.createProduct(ctx, arg)
}
func (m *mockQuerier) DeleteProduct(ctx context.Context, id int) (int, error) {
	return m.deleteProduct(ctx, id)
}
func (m *mockQuerier) GetProductById(ctx context.Context, id int) (db.Product, error) {
	return m.getProductById(ctx, id)
}
func (m *mockQuerier) GetProductsByPlatform(ctx context.Context, platformID int) ([]db.Product, error) {
	return m.getProductsByPlatform(ctx, platformID)
}
func (m *mockQuerier) UpdateProduct(ctx context.Context, arg db.UpdateProductParams) (int, error) {
	return m.updateProduct(ctx, arg)
}

func TestPostgresService_CreateProduct(t *testing.T) {
	ctx := context.Background()
	req := createProductRequest{Name: "Test", Description: "Test Description", PlatformID: 1}

	t.Run("Success", func(t *testing.T) {
		m := &mockQuerier{
			createProduct: func(ctx context.Context, arg db.CreateProductParams) (db.Product, error) {
				return db.Product{Name: arg.Name, Description: arg.Description, PlatformID: arg.PlatformID}, nil
			},
		}
		s := postgresService{queries: m}
		product, err := s.CreateProduct(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if product.Name != "Test" {
			t.Errorf("expected name Test, got %s", product.Name)
		}
		if product.PlatformID != 1 {
			t.Errorf("expected platform ID 1, got %d", product.PlatformID)
		}
		if product.Description.String != "Test Description" {
			t.Errorf("expected description Test Description, got %s", product.Description.String)
		}
	})

	t.Run("DB Error", func(t *testing.T) {
		m := &mockQuerier{
			createProduct: func(ctx context.Context, arg db.CreateProductParams) (db.Product, error) {
				return db.Product{}, errors.New("db error")
			},
		}
		s := postgresService{queries: m}
		_, err := s.CreateProduct(ctx, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestPostgresService_GetProductsByPlatform(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		m := &mockQuerier{
			getProductsByPlatform: func(ctx context.Context, platformID int) ([]db.Product, error) {
				return []db.Product{{ID: 1, PlatformID: platformID}}, nil
			},
		}
		s := postgresService{queries: m}
		products, err := s.GetProductsByPlatform(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(products) != 1 {
			t.Errorf("expected 1 product, got %d", len(products))
		}
	})

	t.Run("DB Error", func(t *testing.T) {
		dbErr := errors.New("db error")
		m := &mockQuerier{
			getProductsByPlatform: func(ctx context.Context, platformID int) ([]db.Product, error) {
				return nil, dbErr
			},
		}
		s := postgresService{queries: m}
		_, err := s.GetProductsByPlatform(ctx, 1)
		if !errors.Is(err, dbErr) {
			t.Errorf("expected error %v, got %v", dbErr, err)
		}
	})
}

func TestPostgresService_GetProductById(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		m := &mockQuerier{
			getProductById: func(ctx context.Context, id int) (db.Product, error) {
				return db.Product{ID: id}, nil
			},
		}
		s := postgresService{queries: m}
		product, err := s.GetProductById(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if product.ID != 1 {
			t.Errorf("expected ID 1, got %d", product.ID)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		m := &mockQuerier{
			getProductById: func(ctx context.Context, id int) (db.Product, error) {
				return db.Product{}, pgx.ErrNoRows
			},
		}
		s := postgresService{queries: m}
		_, err := s.GetProductById(ctx, 1)
		if !internal.IsNotFoundError(err) {
			t.Errorf("expected NotFoundError, got %v", err)
		}
	})

	t.Run("DB Error", func(t *testing.T) {
		dbErr := errors.New("db error")
		m := &mockQuerier{
			getProductById: func(ctx context.Context, id int) (db.Product, error) {
				return db.Product{}, dbErr
			},
		}
		s := postgresService{queries: m}
		_, err := s.GetProductById(ctx, 1)
		if !errors.Is(err, dbErr) {
			t.Errorf("expected error %v, got %v", dbErr, err)
		}
	})
}

func TestPostgresService_UpdateProduct(t *testing.T) {
	ctx := context.Background()
	req := updateProductRequest{Name: "New"}

	t.Run("Success", func(t *testing.T) {
		m := &mockQuerier{
			updateProduct: func(ctx context.Context, arg db.UpdateProductParams) (int, error) {
				return 1, nil
			},
		}
		s := postgresService{queries: m}
		rows, err := s.UpdateProduct(ctx, req, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rows != 1 {
			t.Errorf("expected 1 row, got %d", rows)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		m := &mockQuerier{
			updateProduct: func(ctx context.Context, arg db.UpdateProductParams) (int, error) {
				return 0, nil
			},
		}
		s := postgresService{queries: m}
		_, err := s.UpdateProduct(ctx, req, 1)
		if !internal.IsNotFoundError(err) {
			t.Errorf("expected NotFoundError, got %v", err)
		}
	})

	t.Run("DB Error", func(t *testing.T) {
		dbErr := errors.New("db error")
		m := &mockQuerier{
			updateProduct: func(ctx context.Context, arg db.UpdateProductParams) (int, error) {
				return 0, dbErr
			},
		}
		s := postgresService{queries: m}
		_, err := s.UpdateProduct(ctx, req, 1)
		if !errors.Is(err, dbErr) {
			t.Errorf("expected error %v, got %v", dbErr, err)
		}
	})
}

func TestPostgresService_DeleteProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		m := &mockQuerier{
			deleteProduct: func(ctx context.Context, id int) (int, error) {
				return 1, nil
			},
		}
		s := postgresService{queries: m}
		rows, err := s.DeleteProduct(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rows != 1 {
			t.Errorf("expected 1 row, got %d", rows)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		m := &mockQuerier{
			deleteProduct: func(ctx context.Context, id int) (int, error) {
				return 0, nil
			},
		}
		s := postgresService{queries: m}
		_, err := s.DeleteProduct(ctx, 1)
		if !internal.IsNotFoundError(err) {
			t.Errorf("expected NotFoundError, got %v", err)
		}
	})

	t.Run("DB Error", func(t *testing.T) {
		dbErr := errors.New("db error")
		m := &mockQuerier{
			deleteProduct: func(ctx context.Context, id int) (int, error) {
				return 0, dbErr
			},
		}
		s := postgresService{queries: m}
		_, err := s.DeleteProduct(ctx, 1)
		if !errors.Is(err, dbErr) {
			t.Errorf("expected error %v, got %v", dbErr, err)
		}
	})
}
