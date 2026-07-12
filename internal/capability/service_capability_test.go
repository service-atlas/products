package capability

import (
	"context"
	"errors"
	"products/internal"
	"products/internal/capability/db"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"
)

type mockCapabilityQuerier struct {
	createCapabilityFunc         func(ctx context.Context, arg db.CreateCapabilityParams) (db.Capability, error)
	deleteCapabilityFunc         func(ctx context.Context, id int) (int64, error)
	getCapabilitiesByFlowFunc    func(ctx context.Context, flowID int) ([]db.GetCapabilitiesByFlowRow, error)
	getCapabilitiesByProductFunc func(ctx context.Context, productID int) ([]db.Capability, error)
	getCapabilityFunc            func(ctx context.Context, id int) (db.Capability, error)
	getFlowFunc                  func(ctx context.Context, id int) (string, error)
	getProductFunc               func(ctx context.Context, id int) (string, error)
	getFlowStepFunc              func(ctx context.Context, id int) (int, error)
	updateCapabilityFunc         func(ctx context.Context, arg db.UpdateCapabilityParams) (int64, error)
	createCapabilityStepFunc     func(ctx context.Context, arg db.CreateCapabilityStepParams) (db.CapabilityStep, error)
	getCapabilityStepsFunc       func(ctx context.Context, id int) ([]db.CapabilityStep, error)
	deleteCapabilityStepFunc     func(ctx context.Context, id int) (int64, error)
}

func (m *mockCapabilityQuerier) CreateCapability(ctx context.Context, arg db.CreateCapabilityParams) (db.Capability, error) {
	if m != nil && m.createCapabilityFunc != nil {
		return m.createCapabilityFunc(ctx, arg)
	}
	return db.Capability{}, nil
}

func (m *mockCapabilityQuerier) DeleteCapability(ctx context.Context, id int) (int64, error) {
	if m != nil && m.deleteCapabilityFunc != nil {
		return m.deleteCapabilityFunc(ctx, id)
	}
	return 1, nil
}

func (m *mockCapabilityQuerier) GetCapabilitiesByFlow(ctx context.Context, flowID int) ([]db.GetCapabilitiesByFlowRow, error) {
	if m != nil && m.getCapabilitiesByFlowFunc != nil {
		return m.getCapabilitiesByFlowFunc(ctx, flowID)
	}
	return nil, nil
}

func (m *mockCapabilityQuerier) GetCapabilitiesByProduct(ctx context.Context, productID int) ([]db.Capability, error) {
	if m != nil && m.getCapabilitiesByProductFunc != nil {
		return m.getCapabilitiesByProductFunc(ctx, productID)
	}
	return nil, nil
}

func (m *mockCapabilityQuerier) GetCapability(ctx context.Context, id int) (db.Capability, error) {
	if m != nil && m.getCapabilityFunc != nil {
		return m.getCapabilityFunc(ctx, id)
	}
	return db.Capability{ID: id}, nil
}

func (m *mockCapabilityQuerier) GetFlow(ctx context.Context, id int) (string, error) {
	if m != nil && m.getFlowFunc != nil {
		return m.getFlowFunc(ctx, id)
	}
	return "Test Flow", nil
}

func (m *mockCapabilityQuerier) GetProduct(ctx context.Context, id int) (string, error) {
	if m != nil && m.getProductFunc != nil {
		return m.getProductFunc(ctx, id)
	}
	return "Test Product", nil
}

func (m *mockCapabilityQuerier) UpdateCapability(ctx context.Context, arg db.UpdateCapabilityParams) (int64, error) {
	if m != nil && m.updateCapabilityFunc != nil {
		return m.updateCapabilityFunc(ctx, arg)
	}
	return 1, nil
}

func (m *mockCapabilityQuerier) CreateCapabilityStep(ctx context.Context, arg db.CreateCapabilityStepParams) (db.CapabilityStep, error) {
	if m != nil && m.createCapabilityStepFunc != nil {
		return m.createCapabilityStepFunc(ctx, arg)
	}
	return db.CapabilityStep{}, nil
}

func (m *mockCapabilityQuerier) GetCapabilitySteps(ctx context.Context, id int) ([]db.CapabilityStep, error) {
	if m != nil && m.getCapabilityStepsFunc != nil {
		return m.getCapabilityStepsFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockCapabilityQuerier) DeleteCapabilityStep(ctx context.Context, id int) (int64, error) {
	if m != nil && m.deleteCapabilityStepFunc != nil {
		return m.deleteCapabilityStepFunc(ctx, id)
	}
	return 1, nil
}

func (m *mockCapabilityQuerier) GetFlowStep(ctx context.Context, id int) (int, error) {
	if m != nil && m.getFlowStepFunc != nil {
		return m.getFlowStepFunc(ctx, id)
	}
	return 1, nil
}

func TestService_CreateCapability(t *testing.T) {
	tests := []struct {
		name          string
		req           createCapabilityRequest
		mockSetup     func(m *mockCapabilityQuerier)
		checkError    func(t *testing.T, err error)
		expectedError string
		expectedCap   db.Capability
	}{
		{
			name: "Success",
			req: createCapabilityRequest{
				ProductId: 1,
				Name:      "Test Cap",
			},
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getProductFunc = func(ctx context.Context, id int) (string, error) {
					return "Test Product", nil
				}
				m.createCapabilityFunc = func(ctx context.Context, arg db.CreateCapabilityParams) (db.Capability, error) {
					return db.Capability{ID: 1, Name: arg.Name}, nil
				}
			},
			expectedCap: db.Capability{ID: 1, Name: "Test Cap"},
		},
		{
			name: "Product Not Found - sql.ErrNoRows",
			req: createCapabilityRequest{
				ProductId: 999,
				Name:      "Test Cap",
			},
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getProductFunc = func(ctx context.Context, id int) (string, error) {
					return "", pgx.ErrNoRows
				}
			},
			checkError: func(t *testing.T, err error) {
				if !errors.Is(err, NotFoundError{}) {
					t.Errorf("expected NotFoundError, got %T", err)
				}
			},
			expectedError: "Flow not found",
		},
		{
			name: "Flow Not Found - Empty Name",
			req: createCapabilityRequest{
				ProductId: 999,
				Name:      "Test Cap",
			},
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getProductFunc = func(ctx context.Context, id int) (string, error) {
					return "", nil
				}
			},
			checkError: func(t *testing.T, err error) {
				if !errors.Is(err, NotFoundError{}) {
					t.Errorf("expected NotFoundError, got %T", err)
				}
			},
			expectedError: "Flow not found",
		},
		{
			name: "Database Error on GetFlow",
			req: createCapabilityRequest{
				ProductId: 1,
				Name:      "Test Cap",
			},
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getProductFunc = func(ctx context.Context, id int) (string, error) {
					return "", errors.New("db error")
				}
			},
			expectedError: "db error",
		},
		{
			name: "Database Error on CreateCapability",
			req: createCapabilityRequest{
				ProductId: 1,
				Name:      "Test Cap",
			},
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getProductFunc = func(ctx context.Context, id int) (string, error) {
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
				} else {
					if err.Error() != tt.expectedError {
						t.Errorf("expected error %q, got %q", tt.expectedError, err.Error())
					}
					if tt.checkError != nil {
						tt.checkError(t, err)
					}
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
		checkError    func(t *testing.T, err error)
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
			checkError: func(t *testing.T, err error) {
				if !errors.Is(err, NotFoundError{}) {
					t.Errorf("expected NotFoundError, got %T", err)
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
				} else {
					if err.Error() != tt.expectedError {
						t.Errorf("expected error %q, got %q", tt.expectedError, err.Error())
					}
					if tt.checkError != nil {
						tt.checkError(t, err)
					}
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
		checkError    func(t *testing.T, err error)
		expectedError string
		expectedCaps  []db.Capability
	}{
		{
			name: "Success",
			id:   1,
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getProductFunc = func(ctx context.Context, id int) (string, error) {
					return "Test Product", nil
				}
				m.getCapabilitiesByProductFunc = func(ctx context.Context, productID int) ([]db.Capability, error) {
					return []db.Capability{{ID: 1, Name: "Test Cap"}}, nil
				}
			},
			expectedCaps: []db.Capability{{ID: 1, Name: "Test Cap"}},
		},
		{
			name: "No Capabilities Returns Empty Slice",
			id:   1,
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getProductFunc = func(ctx context.Context, id int) (string, error) {
					return "Test Product", nil
				}
				m.getCapabilitiesByProductFunc = func(ctx context.Context, productID int) ([]db.Capability, error) {
					return nil, nil
				}
			},
			expectedCaps: []db.Capability{},
		},
		{
			name: "Product Not Found",
			id:   999,
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getProductFunc = func(ctx context.Context, id int) (string, error) {
					return "", pgx.ErrNoRows
				}
			},
			checkError: func(t *testing.T, err error) {
				if !errors.Is(err, NotFoundError{}) {
					t.Errorf("expected NotFoundError, got %T", err)
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
				m.getCapabilitiesByProductFunc = func(ctx context.Context, productID int) ([]db.Capability, error) {
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
				} else {
					if err.Error() != tt.expectedError {
						t.Errorf("expected error %q, got %q", tt.expectedError, err.Error())
					}
					if tt.checkError != nil {
						tt.checkError(t, err)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(caps, tt.expectedCaps) {
				t.Errorf("expected capabilities %+v, got %+v", tt.expectedCaps, caps)
			}
		})
	}
}

func TestService_GetCapabilitiesByFlow(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		mockSetup     func(m *mockCapabilityQuerier)
		checkError    func(t *testing.T, err error)
		expectedError string
		expectedCaps  []db.GetCapabilitiesByFlowRow
	}{
		{
			name: "Success",
			id:   1,
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (string, error) {
					return "Test Flow", nil
				}
				m.getCapabilitiesByFlowFunc = func(ctx context.Context, flowID int) ([]db.GetCapabilitiesByFlowRow, error) {
					return []db.GetCapabilitiesByFlowRow{{ID: 1, Name: "Test Cap", FlowName: "Flow A"}}, nil
				}
			},
			expectedCaps: []db.GetCapabilitiesByFlowRow{{ID: 1, Name: "Test Cap", FlowName: "Flow A"}},
		},
		{
			name: "No Capabilities Returns Empty Slice",
			id:   1,
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (string, error) {
					return "Test Flow", nil
				}
				m.getCapabilitiesByFlowFunc = func(ctx context.Context, flowID int) ([]db.GetCapabilitiesByFlowRow, error) {
					return nil, nil
				}
			},
			expectedCaps: []db.GetCapabilitiesByFlowRow{},
		},
		{
			name: "Flow Not Found",
			id:   999,
			mockSetup: func(m *mockCapabilityQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (string, error) {
					return "", pgx.ErrNoRows
				}
			},
			checkError: func(t *testing.T, err error) {
				if !errors.Is(err, NotFoundError{}) {
					t.Errorf("expected NotFoundError, got %T", err)
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
				m.getCapabilitiesByFlowFunc = func(ctx context.Context, flowID int) ([]db.GetCapabilitiesByFlowRow, error) {
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
				} else {
					if err.Error() != tt.expectedError {
						t.Errorf("expected error %q, got %q", tt.expectedError, err.Error())
					}
					if tt.checkError != nil {
						tt.checkError(t, err)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(caps, tt.expectedCaps) {
				t.Errorf("expected capabilities %+v, got %+v", tt.expectedCaps, caps)
			}
		})
	}
}

func TestService_UpdateCapability(t *testing.T) {
	tests := []struct {
		name          string
		req           updateCapabilityRequest
		mockSetup     func(m *mockCapabilityQuerier)
		checkError    func(t *testing.T, err error)
		expectedError string
		expectedCap   db.Capability
	}{
		{
			name: "Success",
			req: updateCapabilityRequest{
				Id:   1,
				Name: "Updated Name",
			},
			mockSetup: func(m *mockCapabilityQuerier) {
				m.updateCapabilityFunc = func(ctx context.Context, arg db.UpdateCapabilityParams) (int64, error) {
					return 1, nil
				}
				m.getCapabilityFunc = func(ctx context.Context, id int) (db.Capability, error) {
					return db.Capability{ID: 1, Name: "Updated Name"}, nil
				}
			},
			expectedCap: db.Capability{ID: 1, Name: "Updated Name"},
		},
		{
			name: "Not Found",
			req: updateCapabilityRequest{
				Id:   999,
				Name: "Updated Name",
			},
			mockSetup: func(m *mockCapabilityQuerier) {
				m.updateCapabilityFunc = func(ctx context.Context, arg db.UpdateCapabilityParams) (int64, error) {
					return 0, nil
				}
			},
			checkError: func(t *testing.T, err error) {
				if !errors.Is(err, NotFoundError{}) {
					t.Errorf("expected NotFoundError, got %T", err)
				}
			},
			expectedError: "capability not found",
		},
		{
			name: "Database Error on Update",
			req: updateCapabilityRequest{
				Id:   1,
				Name: "Updated Name",
			},
			mockSetup: func(m *mockCapabilityQuerier) {
				m.updateCapabilityFunc = func(ctx context.Context, arg db.UpdateCapabilityParams) (int64, error) {
					return 0, errors.New("db error")
				}
			},
			expectedError: "db error",
		},
		{
			name: "Database Error on GetCapability",
			req: updateCapabilityRequest{
				Id:   1,
				Name: "Updated Name",
			},
			mockSetup: func(m *mockCapabilityQuerier) {
				m.updateCapabilityFunc = func(ctx context.Context, arg db.UpdateCapabilityParams) (int64, error) {
					return 1, nil
				}
				m.getCapabilityFunc = func(ctx context.Context, id int) (db.Capability, error) {
					return db.Capability{}, errors.New("get capability error")
				}
			},
			expectedError: "get capability error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCapabilityQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			s := &postgresService{queries: mock}

			cap, err := s.UpdateCapability(context.Background(), tt.req)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
				} else {
					if tt.checkError != nil {
						tt.checkError(t, err)
					} else if err.Error() != tt.expectedError {
						t.Errorf("expected error %q, got %q", tt.expectedError, err.Error())
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(cap, tt.expectedCap) {
				t.Errorf("expected capability %+v, got %+v", tt.expectedCap, cap)
			}
		})
	}
}

func TestService_DeleteCapability(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		mockSetup     func(m *mockCapabilityQuerier)
		checkError    func(t *testing.T, err error)
		expectedError string
	}{
		{
			name: "Success",
			id:   1,
			mockSetup: func(m *mockCapabilityQuerier) {
				m.deleteCapabilityFunc = func(ctx context.Context, id int) (int64, error) {
					return 1, nil
				}
			},
		},
		{
			name: "Not Found",
			id:   999,
			mockSetup: func(m *mockCapabilityQuerier) {
				m.deleteCapabilityFunc = func(ctx context.Context, id int) (int64, error) {
					return 0, nil
				}
			},
			checkError: func(t *testing.T, err error) {
				if !errors.Is(err, internal.NotFoundError{}) {
					t.Errorf("expected NotFoundError, got %T", err)
				}
			},
			expectedError: "capability not found",
		},
		{
			name: "Database Error",
			id:   1,
			mockSetup: func(m *mockCapabilityQuerier) {
				m.deleteCapabilityFunc = func(ctx context.Context, id int) (int64, error) {
					return 0, errors.New("db error")
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

			err := s.DeleteCapability(context.Background(), tt.id)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
				} else {
					if tt.checkError != nil {
						tt.checkError(t, err)
					} else if err.Error() != tt.expectedError {
						t.Errorf("expected error %q, got %q", tt.expectedError, err.Error())
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
