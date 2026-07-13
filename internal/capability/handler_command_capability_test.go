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

	"github.com/go-chi/chi/v5"
)

func TestHandler_CreateCapability(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    any
		mockSetup      func(m *mockCapabilityService)
		expectedStatus int
	}{
		{
			name: "Success",
			requestBody: createCapabilityRequest{
				ProductId: 1,
				Name:      "Test Cap",
			},
			mockSetup: func(m *mockCapabilityService) {
				m.createCapabilityFunc = func(ctx context.Context, req createCapabilityRequest) (db.Capability, error) {
					return db.Capability{ID: 1, Name: req.Name}, nil
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
			requestBody: createCapabilityRequest{
				Name: "", // Name is required
			},
			mockSetup:      func(m *mockCapabilityService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			requestBody: createCapabilityRequest{
				Name:      "Test Cap",
				ProductId: 1,
			},
			mockSetup: func(m *mockCapabilityService) {
				m.createCapabilityFunc = func(ctx context.Context, req createCapabilityRequest) (db.Capability, error) {
					return db.Capability{}, errors.New("service error")
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

			req := httptest.NewRequest(http.MethodPost, "/capabilities", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			mockSvc := &mockCapabilityService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}

			h := &handler{service: mockSvc}
			h.CreateCapability(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandler_UpdateCapability(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		requestBody    any
		mockSetup      func(m *mockCapabilityService)
		expectedStatus int
	}{
		{
			name: "Success",
			id:   "1",
			requestBody: updateCapabilityRequest{
				Id:   1,
				Name: "Updated Name",
			},
			mockSetup: func(m *mockCapabilityService) {
				m.updateCapabilityFunc = func(ctx context.Context, req updateCapabilityRequest) (db.Capability, error) {
					return db.Capability{ID: 1, Name: req.Name}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Not Found",
			id:   "1",
			requestBody: updateCapabilityRequest{
				Id:   1,
				Name: "Updated Name",
			},
			mockSetup: func(m *mockCapabilityService) {
				m.updateCapabilityFunc = func(ctx context.Context, req updateCapabilityRequest) (db.Capability, error) {
					return db.Capability{}, NotFoundError{Msg: "Capability not found"}
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Validation Error",
			id:             "1",
			requestBody:    updateCapabilityRequest{Name: ""},
			mockSetup:      func(m *mockCapabilityService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			id:   "1",
			requestBody: updateCapabilityRequest{
				Id:   1,
				Name: "Updated Name",
			},
			mockSetup: func(m *mockCapabilityService) {
				m.updateCapabilityFunc = func(ctx context.Context, req updateCapabilityRequest) (db.Capability, error) {
					return db.Capability{}, errors.New("service error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Verify ID Overwrite",
			id:   "1",
			requestBody: updateCapabilityRequest{
				Id:   999,
				Name: "Updated Name",
			},
			mockSetup: func(m *mockCapabilityService) {
				m.updateCapabilityFunc = func(ctx context.Context, req updateCapabilityRequest) (db.Capability, error) {
					if req.Id != 1 {
						return db.Capability{}, errors.New("ID was not overridden")
					}
					return db.Capability{ID: 1, Name: req.Name}, nil
				}
			},
			expectedStatus: http.StatusOK,
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

			req := httptest.NewRequest(http.MethodPut, "/capabilities/"+tt.id, bytes.NewBuffer(body))
			if tt.id != "" {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", tt.id)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			}

			w := httptest.NewRecorder()
			mockSvc := &mockCapabilityService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}
			h := &handler{service: mockSvc}

			h.UpdateCapability(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandler_DeleteCapability(t *testing.T) {
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
				m.deleteCapabilityFunc = func(ctx context.Context, id int) error {
					return nil
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "Not Found",
			id:   "999",
			mockSetup: func(m *mockCapabilityService) {
				m.deleteCapabilityFunc = func(ctx context.Context, id int) error {
					return internal.NewNotFoundError(999, "Capability not found")
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Service Error",
			id:   "1",
			mockSetup: func(m *mockCapabilityService) {
				m.deleteCapabilityFunc = func(ctx context.Context, id int) error {
					return errors.New("service error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Invalid ID",
			id:             "abc",
			mockSetup:      func(m *mockCapabilityService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/capabilities/"+tt.id, nil)
			if tt.id != "" {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", tt.id)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			}

			w := httptest.NewRecorder()
			mockSvc := &mockCapabilityService{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockSvc)
			}
			h := &handler{service: mockSvc}

			h.DeleteCapability(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
