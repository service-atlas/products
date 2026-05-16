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
)

type mockFlowQuerier struct {
	createFlowFunc        func(ctx context.Context, arg CreateFlowParams) (Flow, error)
	createFlowStepFunc    func(ctx context.Context, arg CreateFlowStepParams) (int64, error)
	deleteFlowFunc        func(ctx context.Context, id int) (int64, error)
	deleteFlowStepFunc    func(ctx context.Context, id int) (int64, error)
	getFlowFunc           func(ctx context.Context, id int) (Flow, error)
	getFlowStepsFunc      func(ctx context.Context, flowID int) ([]FlowStep, error)
	getFlowsByProductFunc func(ctx context.Context, productID int) ([]GetFlowsByProductRow, error)
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

func (m *mockFlowQuerier) GetFlowsByProduct(ctx context.Context, productID int) ([]GetFlowsByProductRow, error) {
	return m.getFlowsByProductFunc(ctx, productID)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockFlowQuerier{}
			if tt.mockSetup != nil {
				tt.mockSetup(mock)
			}
			h := &flowHandler{queries: mock}

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

			req := httptest.NewRequest(http.MethodPost, "/flows", bytes.NewBuffer(body))
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
