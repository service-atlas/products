package flow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

type mockFlowQuerier struct {
	createFlowFunc        func(ctx context.Context, arg CreateFlowParams) (Flow, error)
	createFlowStepFunc    func(ctx context.Context, arg CreateFlowStepParams) (int64, error)
	deleteFlowFunc        func(ctx context.Context, id int) (int64, error)
	deleteFlowStepFunc    func(ctx context.Context, id int) (int64, error)
	getFlowFunc           func(ctx context.Context, id int) (Flow, error)
	getFlowStepsFunc      func(ctx context.Context, flowID int) ([]FlowStep, error)
	getFlowsByProductFunc func(ctx context.Context, productID int) ([]Flow, error)
	getProductByIdFunc    func(ctx context.Context, id int) (int, error)
	updateFlowFunc        func(ctx context.Context, arg UpdateFlowParams) (int64, error)
}

func (m *mockFlowQuerier) CreateFlow(ctx context.Context, arg CreateFlowParams) (Flow, error) {
	return m.createFlowFunc(ctx, arg)
}

func (m *mockFlowQuerier) CreateFlowStep(ctx context.Context, arg CreateFlowStepParams) (int64, error) {
	return m.createFlowStepFunc(ctx, arg)
}

func (m *mockFlowQuerier) DeleteFlow(ctx context.Context, id int) (int64, error) {
	return m.deleteFlowFunc(ctx, id)
}

func (m *mockFlowQuerier) DeleteFlowStep(ctx context.Context, id int) (int64, error) {
	return m.deleteFlowStepFunc(ctx, id)
}

func (m *mockFlowQuerier) GetFlow(ctx context.Context, id int) (Flow, error) {
	return m.getFlowFunc(ctx, id)
}

func (m *mockFlowQuerier) GetFlowSteps(ctx context.Context, flowID int) ([]FlowStep, error) {
	return m.getFlowStepsFunc(ctx, flowID)
}

func (m *mockFlowQuerier) GetFlowsByProduct(ctx context.Context, productID int) ([]Flow, error) {
	return m.getFlowsByProductFunc(ctx, productID)
}

func (m *mockFlowQuerier) GetProductById(ctx context.Context, id int) (int, error) {
	return m.getProductByIdFunc(ctx, id)
}

func (m *mockFlowQuerier) UpdateFlow(ctx context.Context, arg UpdateFlowParams) (int64, error) {
	return m.updateFlowFunc(ctx, arg)
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
				m.createFlowFunc = func(ctx context.Context, arg CreateFlowParams) (Flow, error) {
					if arg.Name != "Test Flow" {
						return Flow{}, errors.New("unexpected name")
					}
					if arg.ProductID != 1 {
						return Flow{}, errors.New("unexpected product id")
					}
					return Flow{ID: 1, Name: arg.Name, ProductID: arg.ProductID}, nil
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
				m.createFlowFunc = func(ctx context.Context, arg CreateFlowParams) (Flow, error) {
					return Flow{}, errors.New("db error")
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
				m.createFlowFunc = func(ctx context.Context, arg CreateFlowParams) (Flow, error) {
					deadline, ok := ctx.Deadline()
					if !ok {
						return Flow{}, errors.New("deadline not set")
					}
					diff := time.Until(deadline)
					if diff < 4900*time.Millisecond || diff > 5100*time.Millisecond {
						return Flow{}, errors.New("deadline not approximately 5s")
					}
					return Flow{ID: 1, Name: arg.Name, ProductID: arg.ProductID}, nil
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
				m.createFlowFunc = func(ctx context.Context, arg CreateFlowParams) (Flow, error) {
					return Flow{}, pgx.ErrNoRows
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
			h := &flowHandler{flowService: &service{queries: mock}}

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
				var flow Flow
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
				m.getFlowFunc = func(ctx context.Context, id int) (Flow, error) {
					if id != 1 {
						return Flow{}, errors.New("unexpected id")
					}
					return Flow{ID: 1, Name: "Test Flow", ProductID: 10}, nil
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
				m.getFlowFunc = func(ctx context.Context, id int) (Flow, error) {
					return Flow{}, pgx.ErrNoRows
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "Database Error",
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.getFlowFunc = func(ctx context.Context, id int) (Flow, error) {
					return Flow{}, errors.New("db error")
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
			h := &flowHandler{flowService: &service{queries: mock}}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/flows/1", nil)
			req.SetPathValue("id", tt.pathID)
			rr := httptest.NewRecorder()

			h.GetFlowById(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var flow Flow
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
	}{
		{
			name:   "Success",
			pathID: "1",
			mockSetup: func(m *mockFlowQuerier) {
				m.getProductByIdFunc = func(ctx context.Context, id int) (int, error) {
					return 1, nil
				}
				m.getFlowsByProductFunc = func(ctx context.Context, productID int) ([]Flow, error) {
					if productID != 1 {
						return nil, errors.New("unexpected product id")
					}
					return []Flow{
						{ID: 1, Name: "Flow 1"},
						{ID: 2, Name: "Flow 2"},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
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
				m.getFlowsByProductFunc = func(ctx context.Context, productID int) ([]Flow, error) {
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
			h := &flowHandler{flowService: &service{queries: mock}}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/products/1/flows", nil)
			req.SetPathValue("id", tt.pathID)
			rr := httptest.NewRecorder()

			h.GetFlowsByProduct(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var flows []Flow
				if err := json.NewDecoder(rr.Body).Decode(&flows); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if len(flows) != 2 {
					t.Errorf("expected 2 flows, got %v", len(flows))
				}
				if rr.Header().Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type application/json, got %v", rr.Header().Get("Content-Type"))
				}
			}
		})
	}
}
