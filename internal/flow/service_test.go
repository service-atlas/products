package flow

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"products/internal"
	"products/internal/flow/db"
	"strings"
	"testing"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
	client := &http.Client{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			s := &postgresService{queries: mock, client: client}

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
	client := &http.Client{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			s := &postgresService{queries: mock, client: client}

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
	client := &http.Client{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			s := &postgresService{queries: mock, client: client}

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
	client := &http.Client{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			s := &postgresService{queries: mock, client: client}

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

func TestService_CreateFlowStep(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name          string
		req           createFlowStepRequest
		mockSetup     func(m *mockFlowQuerier)
		expectedError string
		expectedStep  db.FlowStep
	}{
		{
			name: "Success",
			req: createFlowStepRequest{
				FlowId:  1,
				Current: validUUID,
				Next:    validUUID,
			},
			mockSetup: func(m *mockFlowQuerier) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					json.NewEncoder(w).Encode([]serviceDependency{
						{Id: validUUID, InteractionType: "data"},
					})
				}))
				t.Cleanup(ts.Close)
				os.Setenv("SERVICE_URL", ts.URL)

				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
				m.createFlowStepFunc = func(ctx context.Context, arg db.CreateFlowStepParams) (db.FlowStep, error) {
					return db.FlowStep{ID: 1, FlowID: 1}, nil
				}
			},
			expectedStep: db.FlowStep{ID: 1, FlowID: 1},
		},
		{
			name: "Dependency Not Found",
			req: createFlowStepRequest{
				FlowId:  1,
				Current: validUUID,
				Next:    "550e8400-e29b-41d4-a716-446655440001",
			},
			mockSetup: func(m *mockFlowQuerier) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					json.NewEncoder(w).Encode([]serviceDependency{
						{Id: validUUID, InteractionType: "data"},
					})
				}))
				t.Cleanup(ts.Close)
				os.Setenv("SERVICE_URL", ts.URL)

				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
			},
			expectedError: "required data dependency not found",
		},
		{
			name: "Service URL Not Set",
			req: createFlowStepRequest{
				FlowId:  1,
				Current: validUUID,
				Next:    validUUID,
			},
			mockSetup: func(m *mockFlowQuerier) {
				os.Unsetenv("SERVICE_URL")
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
			},
			expectedError: "SERVICE_URL is not set",
		},
		{
			name: "Flow Not Found",
			req: createFlowStepRequest{
				FlowId: 999,
			},
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{}, pgx.ErrNoRows
				}
			},
			expectedError: internal.NewNotFoundError(999, "Flow").Error(),
		},
		{
			name: "Invalid UUID in Request",
			req: createFlowStepRequest{
				FlowId:  1,
				Current: "invalid",
			},
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
			},
			expectedError: "invalid uuid",
		},
		{
			name: "Database Error on Create",
			req: createFlowStepRequest{
				FlowId:  1,
				Current: validUUID,
				Next:    validUUID,
			},
			mockSetup: func(m *mockFlowQuerier) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					json.NewEncoder(w).Encode([]serviceDependency{
						{Id: validUUID, InteractionType: "data"},
					})
				}))
				t.Cleanup(ts.Close)
				os.Setenv("SERVICE_URL", ts.URL)

				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
				m.createFlowStepFunc = func(ctx context.Context, arg db.CreateFlowStepParams) (db.FlowStep, error) {
					return db.FlowStep{}, errors.New("db error")
				}
			},
			expectedError: "failed to create flow step: db error",
		},
	}
	client := &http.Client{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			s := &postgresService{queries: mock, client: client}

			step, err := s.CreateFlowStep(t.Context(), tt.req)

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

			if step.ID != tt.expectedStep.ID || step.FlowID != tt.expectedStep.FlowID {
				t.Errorf("expected flow step %+v, got %+v", tt.expectedStep, step)
			}
		})
	}
}

func TestService_validateDependency(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	otherUUID := "550e8400-e29b-41d4-a716-446655440001"

	tests := []struct {
		name          string
		current       string
		next          string
		serviceUrl    string
		handler       http.HandlerFunc
		expectedOk    bool
		expectedError string
	}{
		{
			name:       "Success - Dependency Found",
			current:    validUUID,
			next:       otherUUID,
			serviceUrl: "http://mock-service",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if !strings.Contains(r.URL.Path, validUUID) {
					t.Errorf("expected path to contain %s, got %s", validUUID, r.URL.Path)
				}
				json.NewEncoder(w).Encode([]serviceDependency{
					{Id: otherUUID, InteractionType: "data"},
				})
			},
			expectedOk: true,
		},
		{
			name:       "Success - Case Insensitive ID",
			current:    validUUID,
			next:       "ABC-123",
			serviceUrl: "http://mock-service",
			handler: func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode([]serviceDependency{
					{Id: "abc-123", InteractionType: "data"},
				})
			},
			expectedOk: true,
		},
		{
			name:       "Dependency Not Found - Wrong ID",
			current:    validUUID,
			next:       otherUUID,
			serviceUrl: "http://mock-service",
			handler: func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode([]serviceDependency{
					{Id: "some-other-id", InteractionType: "data"},
				})
			},
			expectedOk: false,
		},
		{
			name:       "Dependency Not Found - Wrong Interaction Type",
			current:    validUUID,
			next:       otherUUID,
			serviceUrl: "http://mock-service",
			handler: func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode([]serviceDependency{
					{Id: otherUUID, InteractionType: "other"},
				})
			},
			expectedOk: false,
		},
		{
			name:       "Empty Response",
			current:    validUUID,
			next:       otherUUID,
			serviceUrl: "http://mock-service",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("[]"))
			},
			expectedOk: false,
		},
		{
			name:          "Service URL Not Set",
			current:       validUUID,
			next:          otherUUID,
			serviceUrl:    "",
			expectedError: "SERVICE_URL is not set",
		},
		{
			name:       "Trailing Slash Handled",
			current:    validUUID,
			next:       otherUUID,
			serviceUrl: "http://mock-service/",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.String(), "//") {
					t.Errorf("URL contains double slashes: %s", r.URL.String())
				}
				json.NewEncoder(w).Encode([]serviceDependency{
					{Id: otherUUID, InteractionType: "data"},
				})
			},
			expectedOk: true,
		},
		{
			name:       "API Error Status",
			current:    validUUID,
			next:       otherUUID,
			serviceUrl: "http://mock-service",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedError: "failed to fetch dependencies: status 500",
		},
		{
			name:       "Malformed JSON",
			current:    validUUID,
			next:       otherUUID,
			serviceUrl: "http://mock-service",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("{invalid json"))
			},
			expectedError: "failed to decode dependencies:",
		},
	}
	client := &http.Client{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.serviceUrl != "" {
				ts := httptest.NewServer(tt.handler)
				defer ts.Close()
				os.Setenv("SERVICE_URL", ts.URL)
				if strings.HasSuffix(tt.serviceUrl, "/") {
					os.Setenv("SERVICE_URL", ts.URL+"/")
				}
			} else {
				os.Unsetenv("SERVICE_URL")
			}

			s := &postgresService{client: client}
			ok, err := s.validateDependency(t.Context(), tt.current, tt.next)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if ok != tt.expectedOk {
				t.Errorf("expected ok=%v, got %v", tt.expectedOk, ok)
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

	client := &http.Client{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			s := &postgresService{queries: mock, client: client}

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

func TestService_DeleteFlowStep(t *testing.T) {
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
				m.deleteFlowStepFunc = func(ctx context.Context, id int) (int64, error) {
					return 1, nil
				}
			},
		},
		{
			name: "Flow Step Not Found",
			id:   999,
			mockSetup: func(m *mockFlowQuerier) {
				m.deleteFlowStepFunc = func(ctx context.Context, id int) (int64, error) {
					return 0, nil
				}
			},
			expectedError: internal.NewNotFoundError(999, "FlowStep").Error(),
		},
		{
			name: "Delete Failed",
			id:   1,
			mockSetup: func(m *mockFlowQuerier) {
				m.deleteFlowStepFunc = func(ctx context.Context, id int) (int64, error) {
					return 0, errors.New("db error")
				}
			},
			expectedError: "failed to delete flow step: db error",
		},
	}

	client := &http.Client{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			s := &postgresService{queries: mock, client: client}

			err := s.DeleteFlowStep(context.Background(), tt.id)

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

func TestService_GetFlowSteps(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		mockSetup     func(m *mockFlowQuerier)
		expectedError string
		expectedSteps []db.FlowStep
	}{
		{
			name: "Success",
			id:   1,
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
				m.getFlowStepsFunc = func(ctx context.Context, flowID int) ([]db.FlowStep, error) {
					return []db.FlowStep{{ID: 1, FlowID: 1}, {ID: 2, FlowID: 1}}, nil
				}
			},
			expectedSteps: []db.FlowStep{{ID: 1, FlowID: 1}, {ID: 2, FlowID: 1}},
		},
		{
			name: "Flow Not Found",
			id:   404,
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{}, pgx.ErrNoRows
				}
			},
			expectedError: internal.NewNotFoundError(404, "Flow").Error(),
		},
		{
			name: "Database Error on GetFlowSteps",
			id:   1,
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
				m.getFlowStepsFunc = func(ctx context.Context, flowID int) ([]db.FlowStep, error) {
					return nil, errors.New("db error")
				}
			},
			expectedError: "db error",
		},
		{
			name: "No Flow Steps Found (Return Empty Slice)",
			id:   1,
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
				m.getFlowStepsFunc = func(ctx context.Context, flowID int) ([]db.FlowStep, error) {
					return nil, nil
				}
			},
			expectedSteps: []db.FlowStep{},
		},
	}
	client := &http.Client{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			s := &postgresService{queries: mock, client: client}

			steps, err := s.GetFlowSteps(context.Background(), tt.id)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(steps) != len(tt.expectedSteps) {
				t.Errorf("expected %d steps, got %d", len(tt.expectedSteps), len(steps))
			}
		})
	}
}

func TestService_GetFlowPath(t *testing.T) {
	uuidA := uuid.New()
	uuidB := uuid.New()
	uuidC := uuid.New()
	uuidD := uuid.New()
	uuidE := uuid.New()

	toPg := func(u uuid.UUID) pgtype.UUID {
		return pgtype.UUID{Bytes: u, Valid: true}
	}

	tests := []struct {
		name          string
		id            int
		mockSetup     func(m *mockFlowQuerier)
		expectedError string
		expectedPath  FlowPath
	}{
		{
			name: "Success",
			id:   1,
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
				m.getFlowStepsFunc = func(ctx context.Context, flowID int) ([]db.FlowStep, error) {
					return []db.FlowStep{
						{ID: 1, FlowID: 1, Current: toPg(uuidA), Next: toPg(uuidB)},
						{ID: 2, FlowID: 1, Current: toPg(uuidA), Next: toPg(uuidC)},
						{ID: 3, FlowID: 1, Current: toPg(uuidB), Next: toPg(uuidD)},
					}, nil
				}
			},
			expectedPath: FlowPath{
				FlowID: 1,
				Path: []PathItem{
					{Current: uuidA.String(), Next: []string{uuidB.String(), uuidC.String()}},
					{Current: uuidB.String(), Next: []string{uuidD.String()}},
				},
			},
		},
		{
			name: "Flow Not Found",
			id:   404,
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{}, pgx.ErrNoRows
				}
			},
			expectedError: internal.NewNotFoundError(404, "Flow").Error(),
		},
		{
			name: "Database Error on GetFlowSteps",
			id:   1,
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
				m.getFlowStepsFunc = func(ctx context.Context, flowID int) ([]db.FlowStep, error) {
					return nil, errors.New("db error")
				}
			},
			expectedError: "db error",
		},
		{
			name: "No Flow Steps Found (Return Empty Path)",
			id:   1,
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
				m.getFlowStepsFunc = func(ctx context.Context, flowID int) ([]db.FlowStep, error) {
					return nil, nil
				}
			},
			expectedPath: FlowPath{FlowID: 1, Path: nil},
		},
		{
			name: "Success - Complex Ordered Path",
			id:   1,
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
				m.getFlowStepsFunc = func(ctx context.Context, flowID int) ([]db.FlowStep, error) {
					// Graph:
					// A -> B
					// B -> C
					// C -> D
					// D -> E
					// A is the only entry point (never appears in Next)
					return []db.FlowStep{
						{ID: 1, FlowID: 1, Current: toPg(uuidA), Next: toPg(uuidB)},
						{ID: 2, FlowID: 1, Current: toPg(uuidB), Next: toPg(uuidC)},
						{ID: 3, FlowID: 1, Current: toPg(uuidC), Next: toPg(uuidD)},
						{ID: 4, FlowID: 1, Current: toPg(uuidD), Next: toPg(uuidE)},
					}, nil
				}
			},
			expectedPath: FlowPath{
				FlowID: 1,
				Path: []PathItem{
					{Current: uuidA.String(), Next: []string{uuidB.String()}},
					{Current: uuidB.String(), Next: []string{uuidC.String()}},
					{Current: uuidC.String(), Next: []string{uuidD.String()}},
					{Current: uuidD.String(), Next: []string{uuidE.String()}},
				},
			},
		},
		{
			name: "Failure - Multiple Entry Points",
			id:   1,
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
				m.getFlowStepsFunc = func(ctx context.Context, flowID int) ([]db.FlowStep, error) {
					// A -> B
					// C -> D
					// Both A and C are entry points
					return []db.FlowStep{
						{ID: 1, FlowID: 1, Current: toPg(uuidA), Next: toPg(uuidB)},
						{ID: 2, FlowID: 1, Current: toPg(uuidC), Next: toPg(uuidD)},
					}, nil
				}
			},
			expectedError: "multiple entry points found in flow",
		},
		{
			name: "Failure - No Entry Point (Cycle)",
			id:   1,
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
				m.getFlowStepsFunc = func(ctx context.Context, flowID int) ([]db.FlowStep, error) {
					// A -> B
					// B -> A
					return []db.FlowStep{
						{ID: 1, FlowID: 1, Current: toPg(uuidA), Next: toPg(uuidB)},
						{ID: 2, FlowID: 1, Current: toPg(uuidB), Next: toPg(uuidA)},
					}, nil
				}
			},
			expectedError: "no entry point found in flow",
		},
	}
	client := &http.Client{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			s := &postgresService{queries: mock, client: client}

			path, err := s.GetFlowPath(context.Background(), tt.id)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if path.FlowID != tt.expectedPath.FlowID {
				t.Errorf("expected flow id %d, got %d", tt.expectedPath.FlowID, path.FlowID)
			}

			if len(path.Path) != len(tt.expectedPath.Path) {
				t.Errorf("expected %d path items, got %d", len(tt.expectedPath.Path), len(path.Path))
			}

			// For ordering-sensitive tests, we check the exact sequence
			if tt.name == "Success - Complex Ordered Path" {
				for i, expectedItem := range tt.expectedPath.Path {
					if i >= len(path.Path) {
						break
					}
					actualItem := path.Path[i]
					if actualItem.Current != expectedItem.Current {
						t.Errorf("at index %d: expected current %s, got %s", i, expectedItem.Current, actualItem.Current)
					}
					if len(actualItem.Next) != len(expectedItem.Next) {
						t.Errorf("at index %d: expected %d next steps, got %d", i, len(expectedItem.Next), len(actualItem.Next))
					}
					for j, expectedNext := range expectedItem.Next {
						if j >= len(actualItem.Next) {
							break
						}
						if actualItem.Next[j] != expectedNext {
							t.Errorf("at index %d, next %d: expected %s, got %s", i, j, expectedNext, actualItem.Next[j])
						}
					}
				}
				return
			}

			for _, expectedItem := range tt.expectedPath.Path {
				found := false
				for _, actualItem := range path.Path {
					if actualItem.Current == expectedItem.Current {
						found = true
						if len(actualItem.Next) != len(expectedItem.Next) {
							t.Errorf("for current %s: expected %d next steps, got %d", expectedItem.Current, len(expectedItem.Next), len(actualItem.Next))
						}
						// Check if all expected next steps are present
						for _, expectedNext := range expectedItem.Next {
							nextFound := false
							for _, actualNext := range actualItem.Next {
								if actualNext == expectedNext {
									nextFound = true
									break
								}
							}
							if !nextFound {
								t.Errorf("for current %s: expected next step %s not found", expectedItem.Current, expectedNext)
							}
						}
						break
					}
				}
				if !found {
					t.Errorf("expected current %s not found in path", expectedItem.Current)
				}
			}
		})
	}
}
