package capability

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"products/internal"
	"products/internal/capability/db"
	"testing"
)

func TestHandler_CreateCapabilityStep(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    any
		mockSetup      func(m *mockCapabilityService)
		expectedStatus int
	}{
		{
			name: "Success",
			requestBody: createCapabilityStepRequest{
				FlowStepId:   1,
				CapabilityId: 1,
				Target:       "Target",
				Protocol:     "Protocol",
			},
			mockSetup: func(m *mockCapabilityService) {
				m.createCapabilityStepFunc = func(ctx context.Context, req createCapabilityStepRequest) (db.CapabilityStep, error) {
					return db.CapabilityStep{ID: 1, CapabilityID: req.CapabilityId, FlowStepID: req.FlowStepId}, nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid JSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *mockCapabilityService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Validation Error",
			requestBody: createCapabilityStepRequest{
				CapabilityId: 0, // Required
			},
			mockSetup:      func(m *mockCapabilityService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Capability Not Found",
			requestBody: createCapabilityStepRequest{
				FlowStepId:   1,
				CapabilityId: 999,
				Target:       "Target",
				Protocol:     "Protocol",
			},
			mockSetup: func(m *mockCapabilityService) {
				m.createCapabilityStepFunc = func(ctx context.Context, req createCapabilityStepRequest) (db.CapabilityStep, error) {
					return db.CapabilityStep{}, internal.NewNotFoundError(999, "Capability not found")
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Service Error",
			requestBody: createCapabilityStepRequest{
				FlowStepId:   1,
				CapabilityId: 1,
				Target:       "Target",
				Protocol:     "Protocol",
			},
			mockSetup: func(m *mockCapabilityService) {
				m.createCapabilityStepFunc = func(ctx context.Context, req createCapabilityStepRequest) (db.CapabilityStep, error) {
					return db.CapabilityStep{}, errors.New("service error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if s, ok := tt.requestBody.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/capability-steps", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			mockSvc := &mockCapabilityService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}

			h := &handler{service: mockSvc}
			h.CreateCapabilityStep(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandler_DeleteCapabilityStep(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockSetup      func(m *mockCapabilityService)
		expectedStatus int
	}{
		{
			name: "Success",
			id:   "1",
			mockSetup: func(m *mockCapabilityService) {
				m.deleteCapabilityStepFunc = func(ctx context.Context, id int) error {
					return nil
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Invalid ID",
			id:             "abc",
			mockSetup:      func(m *mockCapabilityService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Not Found",
			id:   "1",
			mockSetup: func(m *mockCapabilityService) {
				m.deleteCapabilityStepFunc = func(ctx context.Context, id int) error {
					return internal.NewNotFoundError(id, "capability-step")
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Service Error",
			id:   "1",
			mockSetup: func(m *mockCapabilityService) {
				m.deleteCapabilityStepFunc = func(ctx context.Context, id int) error {
					return errors.New("service error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/capability-steps/"+tt.id, nil)
			// Mocking path variable
			req.SetPathValue("id", tt.id)
			w := httptest.NewRecorder()

			mockSvc := &mockCapabilityService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}

			h := &handler{service: mockSvc}
			h.DeleteCapabilityStep(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandler_GetCapabilitySteps(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockSetup      func(m *mockCapabilityService)
		expectedStatus int
	}{
		{
			name: "Success",
			id:   "1",
			mockSetup: func(m *mockCapabilityService) {
				m.getCapabilityStepsFunc = func(ctx context.Context, id int) ([]db.CapabilityStep, error) {
					return []db.CapabilityStep{{ID: 1, CapabilityID: 1}}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid ID",
			id:             "abc",
			mockSetup:      func(m *mockCapabilityService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Capability Not Found",
			id:   "999",
			mockSetup: func(m *mockCapabilityService) {
				m.getCapabilityStepsFunc = func(ctx context.Context, id int) ([]db.CapabilityStep, error) {
					return nil, internal.NewNotFoundError(999, "Capability not found")
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Service Error",
			id:   "1",
			mockSetup: func(m *mockCapabilityService) {
				m.getCapabilityStepsFunc = func(ctx context.Context, id int) ([]db.CapabilityStep, error) {
					return nil, errors.New("service error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/capabilities/"+tt.id+"/steps", nil)
			req.SetPathValue("id", tt.id)
			w := httptest.NewRecorder()

			mockSvc := &mockCapabilityService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}

			h := &handler{service: mockSvc}
			h.GetCapabilitySteps(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
