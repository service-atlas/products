package capability

import (
	"context"
	"errors"
	"products/internal"
	"products/internal/capability/db"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

type mockQuerier struct {
	db.Querier
	GetCapabilityFunc        func(ctx context.Context, id int) (db.Capability, error)
	CreateCapabilityStepFunc func(ctx context.Context, arg db.CreateCapabilityStepParams) (db.CapabilityStep, error)
	GetFlowStepFunc          func(ctx context.Context, id int) (int, error)
	DeleteCapabilityStepFunc func(ctx context.Context, id int) (int64, error)
	GetCapabilityStepsFunc   func(ctx context.Context, id int) ([]db.CapabilityStep, error)
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

func (m *mockQuerier) GetFlow(ctx context.Context, id int) (string, error) {
	return "", nil
}

func (m *mockQuerier) GetProduct(ctx context.Context, id int) (string, error) {
	return "", nil
}

func (m *mockQuerier) CreateCapability(ctx context.Context, arg db.CreateCapabilityParams) (db.Capability, error) {
	return db.Capability{}, nil
}

func (m *mockQuerier) DeleteCapability(ctx context.Context, id int) (int64, error) {
	return 0, nil
}

func (m *mockQuerier) DeleteCapabilityStep(ctx context.Context, id int) (int64, error) {
	if m != nil && m.DeleteCapabilityStepFunc != nil {
		return m.DeleteCapabilityStepFunc(ctx, id)
	}
	return 0, nil
}

func (m *mockQuerier) GetCapabilitiesByFlow(ctx context.Context, flowID int) ([]db.GetCapabilitiesByFlowRow, error) {
	return nil, nil
}

func (m *mockQuerier) GetCapabilitiesByProduct(ctx context.Context, productID int) ([]db.Capability, error) {
	return nil, nil
}

func (m *mockQuerier) GetCapabilitySteps(ctx context.Context, capabilityID int) ([]db.CapabilityStep, error) {
	if m != nil && m.GetCapabilityStepsFunc != nil {
		return m.GetCapabilityStepsFunc(ctx, capabilityID)
	}
	return nil, nil
}

func (m *mockQuerier) UpdateCapability(ctx context.Context, arg db.UpdateCapabilityParams) (int64, error) {
	return 0, nil
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
				return db.Capability{}, pgx.ErrNoRows
			},
			GetFlowStepFunc: func(ctx context.Context, id int) (int, error) {
				return 1, nil
			},
		}
		service := &postgresService{queries: mock}
		req := createCapabilityStepRequest{CapabilityId: 1, FlowStepId: 2, Target: "Target", Protocol: "Protocol"}

		_, err := service.CreateCapabilityStep(ctx, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if pathErr, ok := errors.AsType[internal.NotFoundError](err); ok {
			e := pathErr.Error()
			if !strings.Contains(e, "1") {
				t.Errorf("expected ID 1, got error %v", e)
			}
		} else if customErr, ok := errors.AsType[NotFoundError](err); ok {
			// This is the local NotFoundError which might be returned by GetCapability
			// However, CreateCapabilityStep is supposed to catch it and return internal.NotFoundError
			t.Errorf("expected internal.NotFoundError, got local NotFoundError: %v", customErr)
		} else {
			t.Errorf("expected internal.NotFoundError, got %T: %v", err, err)
		}
	})

	t.Run("CapabilityGeneralError", func(t *testing.T) {
		mock := &mockQuerier{
			GetCapabilityFunc: func(ctx context.Context, id int) (db.Capability, error) {
				return db.Capability{}, errors.New("some db error")
			},
			GetFlowStepFunc: func(ctx context.Context, id int) (int, error) {
				return 1, nil
			},
		}
		service := &postgresService{queries: mock}
		req := createCapabilityStepRequest{CapabilityId: 1, FlowStepId: 2, Target: "Target", Protocol: "Protocol"}

		_, err := service.CreateCapabilityStep(ctx, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "error getting capability" {
			t.Errorf("expected 'error getting capability', got %v", err.Error())
		}
	})

	t.Run("FlowStepNotFound", func(t *testing.T) {
		mock := &mockQuerier{
			GetCapabilityFunc: func(ctx context.Context, id int) (db.Capability, error) {
				return db.Capability{ID: 1}, nil
			},
			GetFlowStepFunc: func(ctx context.Context, id int) (int, error) {
				return 0, pgx.ErrNoRows
			},
		}
		service := &postgresService{queries: mock}
		req := createCapabilityStepRequest{CapabilityId: 1, FlowStepId: 2, Target: "Target", Protocol: "Protocol"}

		_, err := service.CreateCapabilityStep(ctx, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if pathErr, ok := errors.AsType[internal.NotFoundError](err); ok {
			e := pathErr.Error()
			if !strings.Contains(e, "2") {
				t.Errorf("expected ID 2, got error %v", e)
			}
		} else {
			t.Errorf("expected NotFoundError, got %T", err)
		}
	})

	t.Run("FlowStepGeneralError", func(t *testing.T) {
		mock := &mockQuerier{
			GetCapabilityFunc: func(ctx context.Context, id int) (db.Capability, error) {
				return db.Capability{ID: 1}, nil
			},
			GetFlowStepFunc: func(ctx context.Context, id int) (int, error) {
				return 0, errors.New("some db error")
			},
		}
		service := &postgresService{queries: mock}
		req := createCapabilityStepRequest{CapabilityId: 1, FlowStepId: 2, Target: "Target", Protocol: "Protocol"}

		_, err := service.CreateCapabilityStep(ctx, req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "error getting flow step" {
			t.Errorf("expected 'error getting flow step', got %v", err.Error())
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

func TestDeleteCapabilityStep(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mock := &mockQuerier{
			DeleteCapabilityStepFunc: func(ctx context.Context, id int) (int64, error) {
				return 1, nil
			},
		}
		service := &postgresService{queries: mock}

		err := service.DeleteCapabilityStep(ctx, 1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		mock := &mockQuerier{
			DeleteCapabilityStepFunc: func(ctx context.Context, id int) (int64, error) {
				return 0, nil
			},
		}
		service := &postgresService{queries: mock}

		err := service.DeleteCapabilityStep(ctx, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if _, ok := errors.AsType[internal.NotFoundError](err); !ok {
			t.Errorf("expected internal.NotFoundError, got %T", err)
		}
	})

	t.Run("GeneralError", func(t *testing.T) {
		mock := &mockQuerier{
			DeleteCapabilityStepFunc: func(ctx context.Context, id int) (int64, error) {
				return 0, errors.New("db error")
			},
		}
		service := &postgresService{queries: mock}

		err := service.DeleteCapabilityStep(ctx, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if err.Error() != "db error" {
			t.Errorf("expected 'db error', got %v", err.Error())
		}
	})
}

func TestGetCapabilitySteps(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mock := &mockQuerier{
			GetCapabilityFunc: func(ctx context.Context, id int) (db.Capability, error) {
				return db.Capability{ID: 1}, nil
			},
			GetCapabilityStepsFunc: func(ctx context.Context, id int) ([]db.CapabilityStep, error) {
				return []db.CapabilityStep{{ID: 1, CapabilityID: 1}, {ID: 2, CapabilityID: 1}}, nil
			},
		}
		service := &postgresService{queries: mock}

		steps, err := service.GetCapabilitySteps(ctx, 1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(steps) != 2 {
			t.Errorf("expected 2 steps, got %d", len(steps))
		}
	})

	t.Run("CapabilityNotFound", func(t *testing.T) {
		mock := &mockQuerier{
			GetCapabilityFunc: func(ctx context.Context, id int) (db.Capability, error) {
				return db.Capability{}, pgx.ErrNoRows
			},
		}
		service := &postgresService{queries: mock}

		_, err := service.GetCapabilitySteps(ctx, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("GeneralError", func(t *testing.T) {
		mock := &mockQuerier{
			GetCapabilityFunc: func(ctx context.Context, id int) (db.Capability, error) {
				return db.Capability{ID: 1}, nil
			},
			GetCapabilityStepsFunc: func(ctx context.Context, id int) ([]db.CapabilityStep, error) {
				return nil, errors.New("db error")
			},
		}
		service := &postgresService{queries: mock}

		_, err := service.GetCapabilitySteps(ctx, 1)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "db error" {
			t.Errorf("expected 'db error', got %v", err.Error())
		}
	})
}
