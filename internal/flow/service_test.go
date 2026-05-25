package flow

import (
	"context"
	"errors"
	"products/internal"
	"products/internal/flow/db"
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
		expectedFlow  db.Flow
	}{
		{
			name: "Success",
			req: createFlowRequest{
				Name: "Test Flow",
			},
			productID: 1,
			mockSetup: func(m *mockFlowQuerier) {
				m.createFlowFunc = func(ctx context.Context, arg db.CreateFlowParams) (db.Flow, error) {
					return db.Flow{ID: 1, Name: arg.Name, ProductID: arg.ProductID}, nil
				}
			},
			expectedFlow: db.Flow{ID: 1, Name: "Test Flow", ProductID: 1},
		},
		{
			name: "Product Not Found",
			req: createFlowRequest{
				Name: "Test Flow",
			},
			productID: 999,
			mockSetup: func(m *mockFlowQuerier) {
				m.createFlowFunc = func(ctx context.Context, arg db.CreateFlowParams) (db.Flow, error) {
					return db.Flow{}, pgx.ErrNoRows
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
				m.createFlowFunc = func(ctx context.Context, arg db.CreateFlowParams) (db.Flow, error) {
					return db.Flow{}, errors.New("db error")
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
			s := &postgresService{queries: mock}

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
		expectedFlow  db.Flow
	}{
		{
			name: "Success",
			id:   1,
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1, Name: "Test Flow"}, nil
				}
			},
			expectedFlow: db.Flow{ID: 1, Name: "Test Flow"},
		},
		{
			name: "Not Found",
			id:   404,
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{}, pgx.ErrNoRows
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
			s := &postgresService{queries: mock}

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
		expectedFlows []db.Flow
	}{
		{
			name:      "Success",
			productID: 1,
			mockSetup: func(m *mockFlowQuerier) {
				m.getProductByIdFunc = func(ctx context.Context, id int) (int, error) {
					return 1, nil
				}
				m.getFlowsByProductFunc = func(ctx context.Context, id int) ([]db.Flow, error) {
					return []db.Flow{{ID: 1, Name: "Flow 1"}}, nil
				}
			},
			expectedFlows: []db.Flow{{ID: 1, Name: "Flow 1"}},
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
		{
			name:      "No Flows Found",
			productID: 1,
			mockSetup: func(m *mockFlowQuerier) {
				m.getProductByIdFunc = func(ctx context.Context, id int) (int, error) {
					return 1, nil
				}
				m.getFlowsByProductFunc = func(ctx context.Context, id int) ([]db.Flow, error) {
					return nil, pgx.ErrNoRows
				}
			},
			expectedFlows: []db.Flow{},
		},
		{
			name:      "No Flows Found (Nil guard)",
			productID: 1,
			mockSetup: func(m *mockFlowQuerier) {
				m.getProductByIdFunc = func(ctx context.Context, id int) (int, error) {
					return 1, nil
				}
				m.getFlowsByProductFunc = func(ctx context.Context, id int) ([]db.Flow, error) {
					return nil, nil
				}
			},
			expectedFlows: []db.Flow{},
		},
		{
			name:      "Database Error on GetProductById",
			productID: 1,
			mockSetup: func(m *mockFlowQuerier) {
				m.getProductByIdFunc = func(ctx context.Context, id int) (int, error) {
					return 0, errors.New("db error")
				}
			},
			expectedError: "failed to fetch flows: db error",
		},
		{
			name:      "Database Error on GetFlowsByProduct",
			productID: 1,
			mockSetup: func(m *mockFlowQuerier) {
				m.getProductByIdFunc = func(ctx context.Context, id int) (int, error) {
					return 1, nil
				}
				m.getFlowsByProductFunc = func(ctx context.Context, id int) ([]db.Flow, error) {
					return nil, errors.New("db error")
				}
			},
			expectedError: "failed to fetch flows: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			s := &postgresService{queries: mock}

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

func TestService_UpdateFlow(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		req           updateFlowRequest
		mockSetup     func(m *mockFlowQuerier)
		expectedError string
		expectedFlow  db.Flow
	}{
		{
			name: "Success",
			id:   1,
			req: updateFlowRequest{
				Name: "Updated Flow",
			},
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					if id == 1 {
						// Return updated flow on second call (Success case in service calls GetFlowById twice)
						return db.Flow{ID: 1, Name: "Updated Flow"}, nil
					}
					return db.Flow{}, pgx.ErrNoRows
				}
				m.updateFlowFunc = func(ctx context.Context, arg db.UpdateFlowParams) (int64, error) {
					return 1, nil
				}
			},
			expectedFlow: db.Flow{ID: 1, Name: "Updated Flow"},
		},
		{
			name: "Flow Not Found",
			id:   999,
			req: updateFlowRequest{
				Name: "Updated Flow",
			},
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{}, pgx.ErrNoRows
				}
			},
			expectedError: internal.NewNotFoundError(999, "Flow").Error(),
		},
		{
			name: "Update Failed",
			id:   1,
			req: updateFlowRequest{
				Name: "Updated Flow",
			},
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1, Name: "Old Flow"}, nil
				}
				m.updateFlowFunc = func(ctx context.Context, arg db.UpdateFlowParams) (int64, error) {
					return 0, errors.New("db error")
				}
			},
			expectedError: "failed to update flow: db error",
		},
		{
			name: "Flow Not Found on Update (RowsAffected 0)",
			id:   1,
			req: updateFlowRequest{
				Name: "Updated Flow",
			},
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1, Name: "Old Flow"}, nil
				}
				m.updateFlowFunc = func(ctx context.Context, arg db.UpdateFlowParams) (int64, error) {
					return 0, nil
				}
			},
			expectedError: internal.NewNotFoundError(1, "Flow").Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			s := &postgresService{queries: mock}

			flow, err := s.UpdateFlow(context.Background(), tt.req, tt.id)

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

			if flow.ID != tt.expectedFlow.ID || flow.Name != tt.expectedFlow.Name {
				t.Errorf("expected flow %+v, got %+v", tt.expectedFlow, flow)
			}
		})
	}
}

func TestService_DeleteFlow(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		mockSetup     func(m *mockFlowQuerier)
		expectedError string
	}{
		{
			name: "Success",
			id:   1,
			mockSetup: func(m *mockFlowQuerier) {
				m.deleteFlowFunc = func(ctx context.Context, id int) (int64, error) {
					return 1, nil
				}
			},
		},
		{
			name: "Flow Not Found",
			id:   999,
			mockSetup: func(m *mockFlowQuerier) {
				m.deleteFlowFunc = func(ctx context.Context, id int) (int64, error) {
					return 0, nil
				}
			},
			expectedError: internal.NewNotFoundError(999, "Flow").Error(),
		},
		{
			name: "Delete Failed",
			id:   1,
			mockSetup: func(m *mockFlowQuerier) {
				m.deleteFlowFunc = func(ctx context.Context, id int) (int64, error) {
					return 0, errors.New("db error")
				}
			},
			expectedError: "failed to delete flow: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			s := &postgresService{queries: mock}

			err := s.DeleteFlow(context.Background(), tt.id)

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
		})
	}
}
