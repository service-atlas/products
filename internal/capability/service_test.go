package capability

import (
	"context"
	"errors"
	"products/internal/capability/db"
	"testing"
)

type mockCapabilityQuerier struct {
	createCapabilityFunc      func(ctx context.Context, arg db.CreateCapabilityParams) (db.Capability, error)
	deleteCapabilityFunc      func(ctx context.Context, id int) (int64, error)
	getCapabilitiesByFlowFunc func(ctx context.Context, flowID int) ([]db.Capability, error)
	getCapabilityFunc         func(ctx context.Context, id int) (db.Capability, error)
	updateCapabilityFunc      func(ctx context.Context, arg db.UpdateCapabilityParams) (int64, error)
}

func (m *mockCapabilityQuerier) CreateCapability(ctx context.Context, arg db.CreateCapabilityParams) (db.Capability, error) {
	return m.createCapabilityFunc(ctx, arg)
}

func (m *mockCapabilityQuerier) DeleteCapability(ctx context.Context, id int) (int64, error) {
	return m.deleteCapabilityFunc(ctx, id)
}

func (m *mockCapabilityQuerier) GetCapabilitiesByFlow(ctx context.Context, flowID int) ([]db.Capability, error) {
	return m.getCapabilitiesByFlowFunc(ctx, flowID)
}

func (m *mockCapabilityQuerier) GetCapability(ctx context.Context, id int) (db.Capability, error) {
	return m.getCapabilityFunc(ctx, id)
}

func (m *mockCapabilityQuerier) UpdateCapability(ctx context.Context, arg db.UpdateCapabilityParams) (int64, error) {
	return m.updateCapabilityFunc(ctx, arg)
}

func TestService_CreateCapability(t *testing.T) {
	tests := []struct {
		name          string
		req           createCapabilityRequest
		mockSetup     func(m *mockCapabilityQuerier)
		expectedError string
		expectedCap   db.Capability
	}{
		{
			name: "Success",
			req: createCapabilityRequest{
				Name: "Test Cap",
			},
			mockSetup: func(m *mockCapabilityQuerier) {
				m.createCapabilityFunc = func(ctx context.Context, arg db.CreateCapabilityParams) (db.Capability, error) {
					return db.Capability{ID: 1, Name: arg.Name}, nil
				}
			},
			expectedCap: db.Capability{ID: 1, Name: "Test Cap"},
		},
		{
			name: "Database Error",
			req: createCapabilityRequest{
				Name: "Test Cap",
			},
			mockSetup: func(m *mockCapabilityQuerier) {
				m.createCapabilityFunc = func(ctx context.Context, arg db.CreateCapabilityParams) (db.Capability, error) {
					return db.Capability{}, errors.New("db error")
				}
			},
			expectedError: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCapabilityQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			s := &postgresService{queries: mock}

			cap, err := s.CreateCapability(context.Background(), tt.req)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
				} else if err.Error() != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cap.ID != tt.expectedCap.ID || cap.Name != tt.expectedCap.Name {
				t.Errorf("expected capability %+v, got %+v", tt.expectedCap, cap)
			}
		})
	}
}
