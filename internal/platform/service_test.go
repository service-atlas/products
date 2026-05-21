package platform

import (
	"context"
	"errors"
	"products/internal"
	"products/internal/platform/db"
	"testing"

	"github.com/jackc/pgx/v5"
)

type mockQuerier struct {
	createPlatform func(ctx context.Context, arg db.CreatePlatformParams) (db.Platform, error)
	deletePlatform func(ctx context.Context, id int) (int, error)
	getPlatform    func(ctx context.Context, id int) (db.Platform, error)
	getPlatforms   func(ctx context.Context) ([]db.Platform, error)
	updatePlatform func(ctx context.Context, arg db.UpdatePlatformParams) (int, error)
}

func (m *mockQuerier) CreatePlatform(ctx context.Context, arg db.CreatePlatformParams) (db.Platform, error) {
	return m.createPlatform(ctx, arg)
}
func (m *mockQuerier) DeletePlatform(ctx context.Context, id int) (int, error) {
	return m.deletePlatform(ctx, id)
}
func (m *mockQuerier) GetPlatform(ctx context.Context, id int) (db.Platform, error) {
	return m.getPlatform(ctx, id)
}
func (m *mockQuerier) GetPlatforms(ctx context.Context) ([]db.Platform, error) {
	return m.getPlatforms(ctx)
}
func (m *mockQuerier) UpdatePlatform(ctx context.Context, arg db.UpdatePlatformParams) (int, error) {
	return m.updatePlatform(ctx, arg)
}

func TestPostgresService_CreatePlatform(t *testing.T) {
	ctx := context.Background()
	req := createPlatformRequest{Name: "Test", Description: "Desc"}

	t.Run("Success", func(t *testing.T) {
		m := &mockQuerier{
			createPlatform: func(ctx context.Context, arg db.CreatePlatformParams) (db.Platform, error) {
				return db.Platform{ID: 1, Name: arg.Name}, nil
			},
		}
		s := postgresService{db: m}
		platform, err := s.CreatePlatform(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if platform.Name != "Test" {
			t.Errorf("expected name Test, got %s", platform.Name)
		}
	})

	t.Run("DB Error", func(t *testing.T) {
		m := &mockQuerier{
			createPlatform: func(ctx context.Context, arg db.CreatePlatformParams) (db.Platform, error) {
				return db.Platform{}, errors.New("db error")
			},
		}
		s := postgresService{db: m}
		_, err := s.CreatePlatform(ctx, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestPostgresService_GetPlatform(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		m := &mockQuerier{
			getPlatform: func(ctx context.Context, id int) (db.Platform, error) {
				return db.Platform{ID: id, Name: "Test"}, nil
			},
		}
		s := postgresService{db: m}
		platform, err := s.GetPlatform(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if platform.ID != 1 {
			t.Errorf("expected ID 1, got %d", platform.ID)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		m := &mockQuerier{
			getPlatform: func(ctx context.Context, id int) (db.Platform, error) {
				return db.Platform{}, pgx.ErrNoRows
			},
		}
		s := postgresService{db: m}
		_, err := s.GetPlatform(ctx, 1)
		if !errors.Is(err, internal.NotFoundError{}) {
			t.Errorf("expected NotFoundError, got %v", err)
		}
	})

	t.Run("DB Error", func(t *testing.T) {
		m := &mockQuerier{
			getPlatform: func(ctx context.Context, id int) (db.Platform, error) {
				return db.Platform{}, errors.New("db error")
			},
		}
		s := postgresService{db: m}
		_, err := s.GetPlatform(ctx, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestPostgresService_GetPlatforms(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		m := &mockQuerier{
			getPlatforms: func(ctx context.Context) ([]db.Platform, error) {
				return []db.Platform{{ID: 1}, {ID: 2}}, nil
			},
		}
		s := postgresService{db: m}
		platforms, err := s.GetPlatforms(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(platforms) != 2 {
			t.Errorf("expected 2 platforms, got %d", len(platforms))
		}
	})
}

func TestPostgresService_UpdatePlatform(t *testing.T) {
	ctx := context.Background()
	req := updatePlatformRequest{Name: "New Name"}

	t.Run("Success", func(t *testing.T) {
		m := &mockQuerier{
			updatePlatform: func(ctx context.Context, arg db.UpdatePlatformParams) (int, error) {
				return 1, nil
			},
		}
		s := postgresService{db: m}
		rows, err := s.UpdatePlatform(ctx, req, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rows != 1 {
			t.Errorf("expected 1 row, got %d", rows)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		m := &mockQuerier{
			updatePlatform: func(ctx context.Context, arg db.UpdatePlatformParams) (int, error) {
				return 0, nil
			},
		}
		s := postgresService{db: m}
		_, err := s.UpdatePlatform(ctx, req, 1)
		if !errors.Is(err, internal.NotFoundError{}) {
			t.Errorf("expected NotFoundError, got %v", err)
		}
	})
}

func TestPostgresService_DeletePlatform(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		m := &mockQuerier{
			deletePlatform: func(ctx context.Context, id int) (int, error) {
				return 1, nil
			},
		}
		s := postgresService{db: m}
		rows, err := s.DeletePlatform(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rows != 1 {
			t.Errorf("expected 1 row, got %d", rows)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		m := &mockQuerier{
			deletePlatform: func(ctx context.Context, id int) (int, error) {
				return 0, nil
			},
		}
		s := postgresService{db: m}
		_, err := s.DeletePlatform(ctx, 1)
		if !errors.Is(err, internal.NotFoundError{}) {
			t.Errorf("expected NotFoundError, got %v", err)
		}
	})
}
