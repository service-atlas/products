package capability

import (
	"context"
	"errors"
	"products/internal/capability/db"
	"testing"

	"github.com/jackc/pgx/v5"
)

type mockCapabilityQuerier struct {
	createCapabilityFunc         func(ctx context.Context, arg db.CreateCapabilityParams) (db.Capability, error)
	deleteCapabilityFunc         func(ctx context.Context, id int) (int64, error)
	getCapabilitiesByFlowFunc    func(ctx context.Context, flowID int) ([]db.Capability, error)
	getCapabilitiesByProductFunc func(ctx context.Context, productID int) ([]db.GetCapabilitiesByProductRow, error)
	getCapabilityFunc            func(ctx context.Context, id int) (db.Capability, error)
	getFlowFunc                  func(ctx context.Context, id int) (string, error)
	getProductFunc               func(ctx context.Context, id int) (string, error)
	updateCapabilityFunc         func(ctx context.Context, arg db.UpdateCapabilityParams) (int64, error)
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

func (m *mockCapabilityQuerier) GetCapabilitiesByProduct(ctx context.Context, productID int) ([]db.GetCapabilitiesByProductRow, error) {
	return m.getCapabilitiesByProductFunc(ctx, productID)
}

func (m *mockCapabilityQuerier) GetCapability(ctx context.Context, id int) (db.Capability, error) {
	return m.getCapabilityFunc(ctx, id)
}

func (m *mockCapabilityQuerier) GetFlow(ctx context.Context, id int) (string, error) {
	return m.getFlowFunc(ctx, id)
}

func (m *mockCapabilityQuerier) GetProduct(ctx context.Context, id int) (string, error) {
	return m.getProductFunc(ctx, id)
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
				FlowId: 1,
				Name:   "Test Cap",
			},
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (string, error) {
					return "Test Flow", nil
				}
				m.createCapabilityFunc = func(ctx context.Context, arg db.CreateCapabilityParams) (db.Capability, error) {
					return db.Capability{ID: 1, Name: arg.Name}, nil
				}
			},
			expectedCap: db.Capability{ID: 1, Name: "Test Cap"},
		},
		{
			name: "Flow Not Found - sql.ErrNoRows",
			req: createCapabilityRequest{
				FlowId: 999,
				Name:   "Test Cap",
			},
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (string, error) {
					return "", pgx.ErrNoRows
				}
			},
			expectedError: "Flow not found",
		},
		{
			name: "Flow Not Found - Empty Name",
			req: createCapabilityRequest{
				FlowId: 999,
				Name:   "Test Cap",
			},
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (string, error) {
					return "", nil
				}
			},
			expectedError: "Flow not found",
		},
		{
			name: "Database Error on GetFlow",
			req: createCapabilityRequest{
				FlowId: 1,
				Name:   "Test Cap",
			},
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (string, error) {
					return "", errors.New("db error")
				}
			},
			expectedError: "db error",
		},
		{
			name: "Database Error on CreateCapability",
			req: createCapabilityRequest{
				FlowId: 1,
				Name:   "Test Cap",
			},
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (string, error) {
					return "Test Flow", nil
				}
				m.createCapabilityFunc = func(ctx context.Context, arg db.CreateCapabilityParams) (db.Capability, error) {
					return db.Capability{}, errors.New("db error")
				}
			},
			expectedError: "error creating capability",
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

func TestService_GetCapability(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		mockSetup     func(m *mockCapabilityQuerier)
		expectedError string
		expectedCap   db.Capability
	}{
		{
			name: "Success",
			id:   1,
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getCapabilityFunc = func(ctx context.Context, id int) (db.Capability, error) {
					return db.Capability{ID: 1, Name: "Test Cap"}, nil
				}
			},
			expectedCap: db.Capability{ID: 1, Name: "Test Cap"},
		},
		{
			name: "Not Found",
			id:   999,
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getCapabilityFunc = func(ctx context.Context, id int) (db.Capability, error) {
					return db.Capability{}, pgx.ErrNoRows
				}
			},
			expectedError: "Capability not found",
		},
		{
			name: "Database Error",
			id:   1,
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getCapabilityFunc = func(ctx context.Context, id int) (db.Capability, error) {
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

			cap, err := s.GetCapability(context.Background(), tt.id)

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

func TestService_GetCapabilitiesByProduct(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		mockSetup     func(m *mockCapabilityQuerier)
		expectedError string
		expectedCaps  []db.GetCapabilitiesByProductRow
	}{
		{
			name: "Success",
			id:   1,
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getProductFunc = func(ctx context.Context, id int) (string, error) {
					return "Test Product", nil
				}
				m.getCapabilitiesByProductFunc = func(ctx context.Context, productID int) ([]db.GetCapabilitiesByProductRow, error) {
					return []db.GetCapabilitiesByProductRow{{ID: 1, Name: "Test Cap"}}, nil
				}
			},
			expectedCaps: []db.GetCapabilitiesByProductRow{{ID: 1, Name: "Test Cap"}},
		},
		{
			name: "Product Not Found",
			id:   999,
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getProductFunc = func(ctx context.Context, id int) (string, error) {
					return "", pgx.ErrNoRows
				}
			},
			expectedError: "Product not found",
		},
		{
			name: "Database Error on GetProduct",
			id:   1,
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getProductFunc = func(ctx context.Context, id int) (string, error) {
					return "", errors.New("db error")
				}
			},
			expectedError: "db error",
		},
		{
			name: "Database Error on GetCapabilities",
			id:   1,
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getProductFunc = func(ctx context.Context, id int) (string, error) {
					return "Test Product", nil
				}
				m.getCapabilitiesByProductFunc = func(ctx context.Context, productID int) ([]db.GetCapabilitiesByProductRow, error) {
					return nil, errors.New("db error")
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

			caps, err := s.GetCapabilitiesByProduct(context.Background(), tt.id)

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

			if len(caps) != len(tt.expectedCaps) {
				t.Errorf("expected %d capabilities, got %d", len(tt.expectedCaps), len(caps))
			}
		})
	}
}

func TestService_GetCapabilitiesByFlow(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		mockSetup     func(m *mockCapabilityQuerier)
		expectedError string
		expectedCaps  []db.Capability
	}{
		{
			name: "Success",
			id:   1,
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (string, error) {
					return "Test Flow", nil
				}
				m.getCapabilitiesByFlowFunc = func(ctx context.Context, flowID int) ([]db.Capability, error) {
					return []db.Capability{{ID: 1, Name: "Test Cap"}}, nil
				}
			},
			expectedCaps: []db.Capability{{ID: 1, Name: "Test Cap"}},
		},
		{
			name: "Flow Not Found",
			id:   999,
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (string, error) {
					return "", pgx.ErrNoRows
				}
			},
			expectedError: "Flow not found",
		},
		{
			name: "Database Error on GetFlow",
			id:   1,
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (string, error) {
					return "", errors.New("db error")
				}
			},
			expectedError: "db error",
		},
		{
			name: "Database Error on GetCapabilities",
			id:   1,
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (string, error) {
					return "Test Flow", nil
				}
				m.getCapabilitiesByFlowFunc = func(ctx context.Context, flowID int) ([]db.Capability, error) {
					return nil, errors.New("db error")
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

			caps, err := s.GetCapabilitiesByFlow(context.Background(), tt.id)

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

			if len(caps) != len(tt.expectedCaps) {
				t.Errorf("expected %d capabilities, got %d", len(tt.expectedCaps), len(caps))
			}
		})
	}
}
