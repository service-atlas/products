package flow

import (
	"context"
	"errors"
	"products/internal"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestService_CreateFlow(t *testing.T) {
	tests := []struct {
		name          string
		req           createFlowRequest
		productID     int
		mockSetup     func(m *mockFlowQuerier)
		expectedError string
		expectedFlow  Flow
	}{
		{
			name: "Success",
			req: createFlowRequest{
				Name: "Test Flow",
			},
			productID: 1,
			mockSetup: func(m *mockFlowQuerier) {
				m.createFlowFunc = func(ctx context.Context, arg CreateFlowParams) (Flow, error) {
					return Flow{ID: 1, Name: arg.Name, ProductID: arg.ProductID}, nil
				}
			},
			expectedFlow: Flow{ID: 1, Name: "Test Flow", ProductID: 1},
		},
		{
			name: "Product Not Found",
			req: createFlowRequest{
				Name: "Test Flow",
			},
			productID: 999,
			mockSetup: func(m *mockFlowQuerier) {
				m.createFlowFunc = func(ctx context.Context, arg CreateFlowParams) (Flow, error) {
					return Flow{}, pgx.ErrNoRows
				}
			},
			expectedError: internal.NewNotFoundError(999, "Product").Error(),
		},
		{
			name: "Database Error",
			req: createFlowRequest{
				Name: "Test Flow",
			},
			productID: 1,
			mockSetup: func(m *mockFlowQuerier) {
				m.createFlowFunc = func(ctx context.Context, arg CreateFlowParams) (Flow, error) {
					return Flow{}, errors.New("db error")
				}
			},
			expectedError: "failed to create flow: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			s := &service{queries: mock}

			flow, err := s.CreateFlow(context.Background(), tt.req, tt.productID)

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

			if flow != tt.expectedFlow {
				t.Errorf("expected flow %+v, got %+v", tt.expectedFlow, flow)
			}
		})
	}
}

func TestService_GetFlowById(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		mockSetup     func(m *mockFlowQuerier)
		expectedError string
		expectedFlow  Flow
	}{
		{
			name: "Success",
			id:   1,
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (Flow, error) {
					return Flow{ID: 1, Name: "Test Flow"}, nil
				}
			},
			expectedFlow: Flow{ID: 1, Name: "Test Flow"},
		},
		{
			name: "Not Found",
			id:   404,
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (Flow, error) {
					return Flow{}, pgx.ErrNoRows
				}
			},
			expectedError: internal.NewNotFoundError(404, "Flow").Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			s := &service{queries: mock}

			flow, err := s.GetFlowById(context.Background(), tt.id)

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

			if flow != tt.expectedFlow {
				t.Errorf("expected flow %+v, got %+v", tt.expectedFlow, flow)
			}
		})
	}
}

func TestService_GetFlowsByProduct(t *testing.T) {
	tests := []struct {
		name          string
		productID     int
		mockSetup     func(m *mockFlowQuerier)
		expectedError string
		expectedFlows []Flow
	}{
		{
			name:      "Success",
			productID: 1,
			mockSetup: func(m *mockFlowQuerier) {
				m.getProductByIdFunc = func(ctx context.Context, id int) (int, error) {
					return 1, nil
				}
				m.getFlowsByProductFunc = func(ctx context.Context, id int) ([]Flow, error) {
					return []Flow{{ID: 1, Name: "Flow 1"}}, nil
				}
			},
			expectedFlows: []Flow{{ID: 1, Name: "Flow 1"}},
		},
		{
			name:      "Product Not Found",
			productID: 999,
			mockSetup: func(m *mockFlowQuerier) {
				m.getProductByIdFunc = func(ctx context.Context, id int) (int, error) {
					return 0, pgx.ErrNoRows
				}
			},
			expectedError: internal.NewNotFoundError(999, "Product").Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			s := &service{queries: mock}

			flows, err := s.GetFlowsByProduct(context.Background(), tt.productID)

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

			if len(flows) != len(tt.expectedFlows) {
				t.Errorf("expected %d flows, got %d", len(tt.expectedFlows), len(flows))
			}
		})
	}
}
