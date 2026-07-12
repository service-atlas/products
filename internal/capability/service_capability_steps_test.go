package capability

import (
	"context"
	"errors"
	"products/internal/capability/db"
	"testing"
)

type mockQuerier struct {
	db.Querier
	GetCapabilityFunc        func(ctx context.Context, id int) (db.Capability, error)
	CreateCapabilityStepFunc func(ctx context.Context, arg db.CreateCapabilityStepParams) (db.CapabilityStep, error)
}

func (m *mockQuerier) GetCapability(ctx context.Context, id int) (db.Capability, error) {
	return m.GetCapabilityFunc(ctx, id)
}

func (m *mockQuerier) CreateCapabilityStep(ctx context.Context, arg db.CreateCapabilityStepParams) (db.CapabilityStep, error) {
	return m.CreateCapabilityStepFunc(ctx, arg)
}

func TestCreateCapabilityStep(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mock := &mockQuerier{
			GetCapabilityFunc: func(ctx context.Context, id int) (db.Capability, error) {
				return db.Capability{ID: 1}, nil
			},
			CreateCapabilityStepFunc: func(ctx context.Context, arg db.CreateCapabilityStepParams) (db.CapabilityStep, error) {
				return db.CapabilityStep{CapabilityID: 1, FlowStepID: 1}, nil
			},
		}
		service := &postgresService{queries: mock}
		req := createCapabilityStepRequest{CapabilityId: 1}

		res, err := service.CreateCapabilityStep(ctx, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res.CapabilityID != 1 {
			t.Errorf("expected CapabilityID 1, got %d", res.CapabilityID)
		}
	})

	t.Run("CapabilityNotFound", func(t *testing.T) {
		mock := &mockQuerier{
			GetCapabilityFunc: func(ctx context.Context, id int) (db.Capability, error) {
				return db.Capability{}, errors.New("not found")
			},
		}
		service := &postgresService{queries: mock}
		req := createCapabilityStepRequest{CapabilityId: 1}

		_, err := service.CreateCapabilityStep(ctx, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("CreateStepError", func(t *testing.T) {
		mock := &mockQuerier{
			GetCapabilityFunc: func(ctx context.Context, id int) (db.Capability, error) {
				return db.Capability{ID: 1}, nil
			},
			CreateCapabilityStepFunc: func(ctx context.Context, arg db.CreateCapabilityStepParams) (db.CapabilityStep, error) {
				return db.CapabilityStep{}, errors.New("db error")
			},
		}
		service := &postgresService{queries: mock}
		req := createCapabilityStepRequest{CapabilityId: 1}

		_, err := service.CreateCapabilityStep(ctx, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
