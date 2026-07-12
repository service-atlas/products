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
	GetFlowStepFunc          func(ctx context.Context, id int) (int, error)
}

func (m *mockQuerier) GetFlowStep(ctx context.Context, id int) (int, error) {
	if m != nil && m.GetFlowStepFunc != nil {
		return m.GetFlowStepFunc(ctx, id)
	}
	return 1, nil
}

func (m *mockQuerier) GetCapability(ctx context.Context, id int) (db.Capability, error) {
	if m != nil && m.GetCapabilityFunc != nil {
		return m.GetCapabilityFunc(ctx, id)
	}
	return db.Capability{ID: id}, nil
}

func (m *mockQuerier) CreateCapabilityStep(ctx context.Context, arg db.CreateCapabilityStepParams) (db.CapabilityStep, error) {
	if m != nil && m.CreateCapabilityStepFunc != nil {
		return m.CreateCapabilityStepFunc(ctx, arg)
	}
	return db.CapabilityStep{}, nil
}

func TestCreateCapabilityStep(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mock := &mockQuerier{
			GetCapabilityFunc: func(ctx context.Context, id int) (db.Capability, error) {
				return db.Capability{ID: 1}, nil
			},
			GetFlowStepFunc: func(ctx context.Context, id int) (int, error) {
				return 1, nil
			},
			CreateCapabilityStepFunc: func(ctx context.Context, arg db.CreateCapabilityStepParams) (db.CapabilityStep, error) {
				return db.CapabilityStep{CapabilityID: 1, FlowStepID: 1}, nil
			},
		}
		service := &postgresService{queries: mock}
		req := createCapabilityStepRequest{CapabilityId: 1, FlowStepId: 1, Target: "Target", Protocol: "Protocol"}

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
			GetFlowStepFunc: func(ctx context.Context, id int) (int, error) {
				return 1, nil
			},
		}
		service := &postgresService{queries: mock}
		req := createCapabilityStepRequest{CapabilityId: 1, FlowStepId: 1, Target: "Target", Protocol: "Protocol"}

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
			GetFlowStepFunc: func(ctx context.Context, id int) (int, error) {
				return 1, nil
			},
			CreateCapabilityStepFunc: func(ctx context.Context, arg db.CreateCapabilityStepParams) (db.CapabilityStep, error) {
				return db.CapabilityStep{}, errors.New("db error")
			},
		}
		service := &postgresService{queries: mock}
		req := createCapabilityStepRequest{CapabilityId: 1, FlowStepId: 1, Target: "Target", Protocol: "Protocol"}

		_, err := service.CreateCapabilityStep(ctx, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
