package flow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"products/internal/flow/db"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockFlowQuerier struct {
	createFlowFunc        func(ctx context.Context, arg db.CreateFlowParams) (db.Flow, error)
	createFlowStepFunc    func(ctx context.Context, arg db.CreateFlowStepParams) (db.FlowStep, error)
	deleteFlowFunc        func(ctx context.Context, id int) (int64, error)
	deleteFlowStepFunc    func(ctx context.Context, id int) (int64, error)
	getFlowFunc           func(ctx context.Context, id int) (db.Flow, error)
	getFlowStepsFunc      func(ctx context.Context, flowID int) ([]db.FlowStep, error)
	getFlowsByProductFunc func(ctx context.Context, productID int) ([]db.Flow, error)
	getProductByIdFunc    func(ctx context.Context, id int) (int, error)
	getFlowStepFunc       func(ctx context.Context, id int) (db.FlowStep, error)
	updateFlowFunc        func(ctx context.Context, arg db.UpdateFlowParams) (int64, error)
	updateFlowStepFunc    func(ctx context.Context, arg db.UpdateFlowStepParams) (int64, error)
}

func (m *mockFlowQuerier) CreateFlow(ctx context.Context, arg db.CreateFlowParams) (db.Flow, error) {
	return m.createFlowFunc(ctx, arg)
}

func (m *mockFlowQuerier) CreateFlowStep(ctx context.Context, arg db.CreateFlowStepParams) (db.FlowStep, error) {
	return m.createFlowStepFunc(ctx, arg)
}

func (m *mockFlowQuerier) DeleteFlow(ctx context.Context, id int) (int64, error) {
	return m.deleteFlowFunc(ctx, id)
}

func (m *mockFlowQuerier) DeleteFlowStep(ctx context.Context, id int) (int64, error) {
	return m.deleteFlowStepFunc(ctx, id)
}

func (m *mockFlowQuerier) GetFlow(ctx context.Context, id int) (db.Flow, error) {
	return m.getFlowFunc(ctx, id)
}

func (m *mockFlowQuerier) GetFlowStep(ctx context.Context, id int) (db.FlowStep, error) {
	return m.getFlowStepFunc(ctx, id)
}

func (m *mockFlowQuerier) GetFlowSteps(ctx context.Context, flowID int) ([]db.FlowStep, error) {
	return m.getFlowStepsFunc(ctx, flowID)
}

func (m *mockFlowQuerier) GetFlowsByProduct(ctx context.Context, productID int) ([]db.Flow, error) {
	return m.getFlowsByProductFunc(ctx, productID)
}

func (m *mockFlowQuerier) GetProductById(ctx context.Context, id int) (int, error) {
	return m.getProductByIdFunc(ctx, id)
}

func (m *mockFlowQuerier) UpdateFlow(ctx context.Context, arg db.UpdateFlowParams) (int64, error) {
	return m.updateFlowFunc(ctx, arg)
}

func (m *mockFlowQuerier) UpdateFlowStep(ctx context.Context, arg db.UpdateFlowStepParams) (int64, error) {
	return m.updateFlowStepFunc(ctx, arg)
}

func TestUpdateFlow(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    any
		pathID         string
		mockSetup      func(m *mockFlowQuerier)
		expectedStatus int
	}{
		{
			name: "Success",
			requestBody: updateFlowRequest{
				Name: "Updated Flow",
			},
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1, Name: "Old Flow"}, nil
				}
				m.updateFlowFunc = func(ctx context.Context, arg db.UpdateFlowParams) (int64, error) {
					return 1, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No Fields Provided",
			requestBody:    updateFlowRequest{},
			pathID:         "1",
			mockSetup:      func(m *mockFlowQuerier) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Flow Not Found",
			requestBody: updateFlowRequest{
				Name: "Updated Flow",
			},
			pathID: "999",
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{}, pgx.ErrNoRows
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid Path ID",
			requestBody:    updateFlowRequest{Name: "Updated Flow"},
			pathID:         "abc",
			mockSetup:      func(m *mockFlowQuerier) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Malformed JSON",
			requestBody:    []byte(`{"name": "missing quote}`),
			pathID:         "1",
			mockSetup:      func(m *mockFlowQuerier) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Failure",
			requestBody: updateFlowRequest{
				Name: "Updated Flow",
			},
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1, Name: "Old Flow"}, nil
				}
				m.updateFlowFunc = func(ctx context.Context, arg db.UpdateFlowParams) (int64, error) {
					return 0, errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			h := &flowHandler{
				flowService: &postgresService{
					queries: mock,
				},
			}

			var body []byte
			switch b := tt.requestBody.(type) {
			case []byte:
				body = b
			case nil:
				body = nil
			default:
				body, _ = json.Marshal(tt.requestBody)
			}
			req := httptest.NewRequestWithContext(t.Context(), "PUT", "/api/flows/"+tt.pathID, bytes.NewBuffer(body))

			req.SetPathValue("id", tt.pathID)

			rr := httptest.NewRecorder()
			h.UpdateFlow(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestCreateFlow(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    any
		pathID         string
		mockSetup      func(m *mockFlowQuerier)
		expectedStatus int
	}{
		{
			name: "Success",
			requestBody: createFlowRequest{
				Name: "Test Flow",
			},
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.createFlowFunc = func(ctx context.Context, arg db.CreateFlowParams) (db.Flow, error) {
					if arg.Name != "Test Flow" {
						return db.Flow{}, errors.New("unexpected name")
					}
					if arg.ProductID != 1 {
						return db.Flow{}, errors.New("unexpected product id")
					}
					return db.Flow{ID: 1, Name: arg.Name, ProductID: arg.ProductID}, nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid JSON",
			requestBody:    "invalid json",
			pathID:         "1",
			mockSetup:      func(m *mockFlowQuerier) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing Name",
			requestBody:    createFlowRequest{},
			pathID:         "1",
			mockSetup:      func(m *mockFlowQuerier) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Database Failure",
			requestBody: createFlowRequest{
				Name: "Fail Flow",
			},
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.createFlowFunc = func(ctx context.Context, arg db.CreateFlowParams) (db.Flow, error) {
					return db.Flow{}, errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Context Timeout Verification",
			requestBody: createFlowRequest{
				Name: "Timeout Flow",
			},
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.createFlowFunc = func(ctx context.Context, arg db.CreateFlowParams) (db.Flow, error) {
					deadline, ok := ctx.Deadline()
					if !ok {
						return db.Flow{}, errors.New("deadline not set")
					}
					diff := time.Until(deadline)
					if diff < 4900*time.Millisecond || diff > 5100*time.Millisecond {
						return db.Flow{}, errors.New("deadline not approximately 5s")
					}
					return db.Flow{ID: 1, Name: arg.Name, ProductID: arg.ProductID}, nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Invalid Product ID",
			requestBody: createFlowRequest{
				Name: "Invalid ID Flow",
			},
			pathID:         "invalid",
			mockSetup:      func(m *mockFlowQuerier) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Product Not Found",
			requestBody: createFlowRequest{
				Name: "Missing Product Flow",
			},
			pathID: "999",
			mockSetup: func(m *mockFlowQuerier) {
				m.createFlowFunc = func(ctx context.Context, arg db.CreateFlowParams) (db.Flow, error) {
					return db.Flow{}, pgx.ErrNoRows
				}
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			h := &flowHandler{flowService: &postgresService{queries: mock}}

			var body []byte
			if s, ok := tt.requestBody.(string); ok {
				body = []byte(s)
			} else {
				var err error
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("json.Marshal requestBody failed: %v", err)
				}
			}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/flows", bytes.NewBuffer(body))
			req.SetPathValue("id", tt.pathID)
			rr := httptest.NewRecorder()

			h.CreateFlow(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusCreated {
				var flow db.Flow
				if err := json.NewDecoder(rr.Body).Decode(&flow); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if flow.ID == 0 {
					t.Error("expected flow ID to be set")
				}
				if rr.Header().Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type application/json, got %v", rr.Header().Get("Content-Type"))
				}
			}
		})
	}
}

func TestGetFlowById(t *testing.T) {
	tests := []struct {
		name           string
		pathID         string
		mockSetup      func(m *mockFlowQuerier)
		expectedStatus int
	}{
		{
			name:   "Success",
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					if id != 1 {
						return db.Flow{}, errors.New("unexpected id")
					}
					return db.Flow{ID: 1, Name: "Test Flow", ProductID: 10}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid ID",
			pathID:         "abc",
			mockSetup:      func(m *mockFlowQuerier) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Flow Not Found",
			pathID: "99",
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{}, pgx.ErrNoRows
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "Database Error",
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{}, errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			h := &flowHandler{flowService: &postgresService{queries: mock}}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/flows/1", nil)
			req.SetPathValue("id", tt.pathID)
			rr := httptest.NewRecorder()

			h.GetFlowById(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var flow db.Flow
				if err := json.NewDecoder(rr.Body).Decode(&flow); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if flow.ID != 1 {
					t.Errorf("expected flow ID 1, got %v", flow.ID)
				}
				if rr.Header().Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type application/json, got %v", rr.Header().Get("Content-Type"))
				}
			}
		})
	}
}

func TestGetFlowsByProduct(t *testing.T) {
	tests := []struct {
		name           string
		pathID         string
		mockSetup      func(m *mockFlowQuerier)
		expectedStatus int
		expectedCount  int
	}{
		{
			name:   "Success",
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.getProductByIdFunc = func(ctx context.Context, id int) (int, error) {
					return 1, nil
				}
				m.getFlowsByProductFunc = func(ctx context.Context, productID int) ([]db.Flow, error) {
					if productID != 1 {
						return nil, errors.New("unexpected product id")
					}
					return []db.Flow{
						{ID: 1, Name: "Flow 1"},
						{ID: 2, Name: "Flow 2"},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "Invalid Product ID",
			pathID:         "invalid",
			mockSetup:      func(m *mockFlowQuerier) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "No Flows Found",
			pathID: "2",
			mockSetup: func(m *mockFlowQuerier) {
				m.getProductByIdFunc = func(ctx context.Context, id int) (int, error) {
					return 2, nil
				}
				m.getFlowsByProductFunc = func(ctx context.Context, productID int) ([]db.Flow, error) {
					return nil, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:   "Product Not Found",
			pathID: "3",
			mockSetup: func(m *mockFlowQuerier) {
				m.getProductByIdFunc = func(ctx context.Context, id int) (int, error) {
					return 0, pgx.ErrNoRows
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "Database Error",
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.getProductByIdFunc = func(ctx context.Context, id int) (int, error) {
					return 1, nil
				}
				m.getFlowsByProductFunc = func(ctx context.Context, productID int) ([]db.Flow, error) {
					return nil, errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			h := &flowHandler{flowService: &postgresService{queries: mock}}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/products/1/flows", nil)
			req.SetPathValue("id", tt.pathID)
			rr := httptest.NewRecorder()

			h.GetFlowsByProduct(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var flows []db.Flow
				if err := json.NewDecoder(rr.Body).Decode(&flows); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if len(flows) != tt.expectedCount {
					t.Errorf("expected %d flows, got %v", tt.expectedCount, len(flows))
				}
				if rr.Header().Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type application/json, got %v", rr.Header().Get("Content-Type"))
				}
			}
		})
	}
}

func TestDeleteFlow(t *testing.T) {
	tests := []struct {
		name           string
		pathID         string
		mockSetup      func(m *mockFlowQuerier)
		expectedStatus int
	}{
		{
			name:   "Success",
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.deleteFlowFunc = func(ctx context.Context, id int) (int64, error) {
					if id != 1 {
						return 0, errors.New("unexpected flow id")
					}
					return 1, nil
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Invalid Flow ID",
			pathID:         "invalid",
			mockSetup:      func(m *mockFlowQuerier) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Flow Not Found",
			pathID: "999",
			mockSetup: func(m *mockFlowQuerier) {
				m.deleteFlowFunc = func(ctx context.Context, id int) (int64, error) {
					return 0, nil
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "Database Error",
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.deleteFlowFunc = func(ctx context.Context, id int) (int64, error) {
					return 0, errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			h := &flowHandler{flowService: &postgresService{queries: mock}}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/flows/1", nil)
			req.SetPathValue("id", tt.pathID)
			rr := httptest.NewRecorder()

			h.DeleteFlow(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestDeleteFlowStep(t *testing.T) {
	tests := []struct {
		name           string
		pathID         string
		mockSetup      func(m *mockFlowQuerier)
		expectedStatus int
	}{
		{
			name:   "Success",
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.deleteFlowStepFunc = func(ctx context.Context, id int) (int64, error) {
					if id != 1 {
						return 0, errors.New("unexpected flow step id")
					}
					return 1, nil
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Invalid Flow Step ID",
			pathID:         "invalid",
			mockSetup:      func(m *mockFlowQuerier) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Flow Step Not Found",
			pathID: "999",
			mockSetup: func(m *mockFlowQuerier) {
				m.deleteFlowStepFunc = func(ctx context.Context, id int) (int64, error) {
					return 0, nil
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "Database Error",
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.deleteFlowStepFunc = func(ctx context.Context, id int) (int64, error) {
					return 0, errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			h := &flowHandler{flowService: &postgresService{queries: mock}}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/flows/steps/1", nil)
			req.SetPathValue("id", tt.pathID)
			rr := httptest.NewRecorder()

			h.DeleteFlowStep(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestCreateFlowStep(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name           string
		requestBody    any
		pathID         string
		mockSetup      func(m *mockFlowQuerier)
		expectedStatus int
	}{
		{
			name: "Success",
			requestBody: createFlowStepRequest{
				Protocol: "http",
				Target:   "target",
				Current:  validUUID,
				Next:     validUUID,
			},
			pathID: "1",
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
					return db.FlowStep{ID: 10, FlowID: 1}, nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid Path ID",
			requestBody:    createFlowStepRequest{Current: validUUID, Next: validUUID},
			pathID:         "abc",
			mockSetup:      func(m *mockFlowQuerier) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Malformed JSON",
			requestBody:    []byte(`{"current": "invalid json`),
			pathID:         "1",
			mockSetup:      func(m *mockFlowQuerier) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Flow Not Found",
			requestBody: createFlowStepRequest{
				Current: validUUID,
				Next:    validUUID,
			},
			pathID: "999",
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{}, pgx.ErrNoRows
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Service Failure",
			requestBody: createFlowStepRequest{
				Current: validUUID,
				Next:    validUUID,
			},
			pathID: "1",
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
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Invalid UUID in Request Body",
			requestBody: createFlowStepRequest{
				Current: "invalid-uuid",
				Next:    validUUID,
			},
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing Current Field",
			requestBody: createFlowStepRequest{
				Protocol: "http",
				Target:   "target",
				Next:     validUUID,
			},
			pathID:         "1",
			mockSetup:      func(m *mockFlowQuerier) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Duplicate Constraint Error",
			requestBody: createFlowStepRequest{
				Protocol: "http",
				Target:   "target",
				Current:  validUUID,
				Next:     validUUID,
			},
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
				// Simulate a conflict error returned by the service layer
				m.createFlowStepFunc = func(ctx context.Context, arg db.CreateFlowStepParams) (db.FlowStep, error) {
					return db.FlowStep{}, ConflictError{Message: "Flow step already exists"}
				}
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "Dependency Validation Failure",
			requestBody: createFlowStepRequest{
				Protocol: "http",
				Target:   "target",
				Current:  validUUID,
				Next:     validUUID,
			},
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Return empty dependencies so validation fails
					json.NewEncoder(w).Encode([]serviceDependency{})
				}))
				t.Cleanup(ts.Close)
				os.Setenv("SERVICE_URL", ts.URL)

				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
	}
	client := &http.Client{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			h := &flowHandler{
				flowService: &postgresService{
					queries: mock,
					client:  client,
				},
			}

			var body []byte
			switch b := tt.requestBody.(type) {
			case []byte:
				body = b
			default:
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/flows/"+tt.pathID+"/steps", bytes.NewBuffer(body))
			req.SetPathValue("id", tt.pathID)

			rr := httptest.NewRecorder()
			h.CreateFlowStep(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusCreated {
				var step db.FlowStep
				if err := json.NewDecoder(rr.Body).Decode(&step); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if step.ID == 0 {
					t.Error("expected flow step ID to be set")
				}
			}
		})
	}
}

func TestUpdateFlowStep(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		reqBody        string
		mockSetup      func(m *mockFlowQuerier)
		expectedStatus int
	}{
		{
			name:    "Success",
			id:      "1",
			reqBody: `{"target": "https://example.com", "protocol": "HTTPS"}`,
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowStepFunc = func(ctx context.Context, id int) (db.FlowStep, error) {
					return db.FlowStep{ID: 1}, nil
				}
				m.updateFlowStepFunc = func(ctx context.Context, arg db.UpdateFlowStepParams) (int64, error) {
					return 1, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid ID",
			id:             "abc",
			reqBody:        `{}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Flow Step Not Found",
			id:      "999",
			reqBody: `{}`,
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowStepFunc = func(ctx context.Context, id int) (db.FlowStep, error) {
					return db.FlowStep{}, pgx.ErrNoRows
				}
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(m)
			}
			h := &flowHandler{flowService: &postgresService{queries: m, client: &http.Client{}}}

			req := httptest.NewRequest(http.MethodPatch, "/api/flow-steps/"+tt.id, strings.NewReader(tt.reqBody))
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			req.SetPathValue("id", tt.id)
			rr := httptest.NewRecorder()
			h.UpdateFlowStep(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestGetFlowSteps(t *testing.T) {
	tests := []struct {
		name           string
		pathID         string
		mockSetup      func(m *mockFlowQuerier)
		expectedStatus int
		expectedCount  int
	}{
		{
			name:   "Success",
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
				m.getFlowStepsFunc = func(ctx context.Context, flowID int) ([]db.FlowStep, error) {
					return []db.FlowStep{{ID: 1, FlowID: 1}, {ID: 2, FlowID: 1}}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:   "Flow Not Found",
			pathID: "404",
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{}, pgx.ErrNoRows
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid ID",
			pathID:         "abc",
			mockSetup:      func(m *mockFlowQuerier) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Database Error",
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
				m.getFlowStepsFunc = func(ctx context.Context, flowID int) ([]db.FlowStep, error) {
					return nil, errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "Success - No Flow Steps",
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
				m.getFlowStepsFunc = func(ctx context.Context, flowID int) ([]db.FlowStep, error) {
					return nil, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
	}

	client := &http.Client{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			h := &flowHandler{
				flowService: &postgresService{
					queries: mock,
					client:  client,
				},
			}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/flows/"+tt.pathID+"/steps", nil)
			req.SetPathValue("id", tt.pathID)

			rr := httptest.NewRecorder()
			h.GetFlowSteps(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var steps []db.FlowStep
				if err := json.NewDecoder(rr.Body).Decode(&steps); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if len(steps) != tt.expectedCount {
					t.Errorf("expected %d steps, got %d", tt.expectedCount, len(steps))
				}
			}
		})
	}
}

func TestGetFlowPath(t *testing.T) {
	tests := []struct {
		name           string
		pathID         string
		mockSetup      func(m *mockFlowQuerier)
		expectedStatus int
	}{
		{
			name:   "Success",
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
				m.getFlowStepsFunc = func(ctx context.Context, flowID int) ([]db.FlowStep, error) {
					u1 := [16]byte{1}
					u2 := [16]byte{2}
					return []db.FlowStep{
						{
							ID:      1,
							FlowID:  1,
							Current: pgtype.UUID{Bytes: u1, Valid: true},
							Next:    pgtype.UUID{Bytes: u2, Valid: true},
						},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Flow Not Found",
			pathID: "404",
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{}, pgx.ErrNoRows
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid ID",
			pathID:         "abc",
			mockSetup:      func(m *mockFlowQuerier) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Database Error",
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (db.Flow, error) {
					return db.Flow{ID: 1}, nil
				}
				m.getFlowStepsFunc = func(ctx context.Context, flowID int) ([]db.FlowStep, error) {
					return nil, errors.New("db error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	client := &http.Client{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			h := &flowHandler{
				flowService: &postgresService{
					queries: mock,
					client:  client,
				},
			}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/flows/"+tt.pathID+"/path", nil)
			req.SetPathValue("id", tt.pathID)

			rr := httptest.NewRecorder()
			h.GetFlowPath(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var path FlowPath
				if err := json.NewDecoder(rr.Body).Decode(&path); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if path.FlowID != 1 {
					t.Errorf("expected flow id 1, got %d", path.FlowID)
				}
				if path.Path == nil || len(path.Path) != 1 {
					t.Fatalf("expected 1 path item, got %d", len(path.Path))
				}
				item := path.Path[0]
				u1 := [16]byte{1}
				u2 := [16]byte{2}
				expectedCurrent := uuid.UUID(u1).String()
				expectedNext := uuid.UUID(u2).String()

				if item.Current != expectedCurrent {
					t.Errorf("expected current %s, got %s", expectedCurrent, item.Current)
				}
				if len(item.Next) != 1 || item.Next[0] != expectedNext {
					t.Errorf("expected next [%s], got %v", expectedNext, item.Next)
				}
			}
		})
	}
}
