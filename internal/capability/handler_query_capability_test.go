package capability

import (
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

func TestHandler_GetCapability(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockSetup      func(m *mockCapabilityService)
		expectedStatus int
		assertBody     func(t *testing.T, body []byte)
	}{
		{
			name: "Success",
			id:   "1",
			mockSetup: func(m *mockCapabilityService) {
				m.getCapabilityFunc = func(ctx context.Context, id int) (db.Capability, error) {
					return db.Capability{ID: 1, Name: "Test Cap"}, nil
				}
			},
			expectedStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var got db.Capability
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if got.ID != 1 || got.Name != "Test Cap" {
					t.Errorf("unexpected capability: %+v", got)
				}
			},
		},
		{
			name: "Not Found",
			id:   "999",
			mockSetup: func(m *mockCapabilityService) {
				m.getCapabilityFunc = func(ctx context.Context, id int) (db.Capability, error) {
					return db.Capability{}, internal.NewNotFoundError(999, "capability")
				}
			},
			expectedStatus: http.StatusNotFound,
			assertBody: func(t *testing.T, body []byte) {
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if got["detail"] != "Capability not found" {
					t.Errorf("unexpected detail: %v", got["detail"])
				}
				if got["status"] != float64(http.StatusNotFound) {
					t.Errorf("unexpected status: %v", got["status"])
				}
			},
		},
		{
			name:           "Invalid ID",
			id:             "abc",
			mockSetup:      func(m *mockCapabilityService) {},
			expectedStatus: http.StatusBadRequest,
			assertBody: func(t *testing.T, body []byte) {
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if got["detail"] != "Invalid capability ID" {
					t.Errorf("unexpected detail: %v", got["detail"])
				}
				if got["status"] != float64(http.StatusBadRequest) {
					t.Errorf("unexpected status: %v", got["status"])
				}
			},
		},
		{
			name: "Service Error",
			id:   "1",
			mockSetup: func(m *mockCapabilityService) {
				m.getCapabilityFunc = func(ctx context.Context, id int) (db.Capability, error) {
					return db.Capability{}, errors.New("service error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			assertBody: func(t *testing.T, body []byte) {
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if got["detail"] != "Failed to fetch capability" {
					t.Errorf("unexpected detail: %v", got["detail"])
				}
				if got["status"] != float64(http.StatusInternalServerError) {
					t.Errorf("unexpected status: %v", got["status"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/capabilities/"+tt.id, nil)
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

			h.GetCapability(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if tt.assertBody != nil {
				tt.assertBody(t, w.Body.Bytes())
			}
		})
	}
}

func TestHandler_GetCapabilitiesByFlow(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockSetup      func(m *mockCapabilityService)
		expectedStatus int
		assertBody     func(t *testing.T, body []byte)
	}{
		{
			name: "Success",
			id:   "1",
			mockSetup: func(m *mockCapabilityService) {
				m.getCapabilitiesByFlowFunc = func(ctx context.Context, id int) ([]db.GetCapabilitiesByFlowRow, error) {
					return []db.GetCapabilitiesByFlowRow{{ID: 1, Name: "Test Cap"}}, nil
				}
			},
			expectedStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var got []db.Capability
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if len(got) != 1 || got[0].ID != 1 || got[0].Name != "Test Cap" {
					t.Errorf("unexpected capabilities: %+v", got)
				}
			},
		},
		{
			name: "Flow Not Found",
			id:   "999",
			mockSetup: func(m *mockCapabilityService) {
				m.getCapabilitiesByFlowFunc = func(ctx context.Context, id int) ([]db.GetCapabilitiesByFlowRow, error) {
					return nil, internal.NewNotFoundError(999, "flow")
				}
			},
			expectedStatus: http.StatusNotFound,
			assertBody: func(t *testing.T, body []byte) {
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if got["detail"] != "Flow not found" {
					t.Errorf("unexpected detail: %v", got["detail"])
				}
				if got["status"] != float64(http.StatusNotFound) {
					t.Errorf("unexpected status: %v", got["status"])
				}
			},
		},
		{
			name:           "Invalid ID",
			id:             "abc",
			mockSetup:      func(m *mockCapabilityService) {},
			expectedStatus: http.StatusBadRequest,
			assertBody: func(t *testing.T, body []byte) {
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if got["detail"] != "Invalid capability ID" {
					t.Errorf("unexpected detail: %v", got["detail"])
				}
				if got["status"] != float64(http.StatusBadRequest) {
					t.Errorf("unexpected status: %v", got["status"])
				}
			},
		},
		{
			name: "Service Error",
			id:   "1",
			mockSetup: func(m *mockCapabilityService) {
				m.getCapabilitiesByFlowFunc = func(ctx context.Context, id int) ([]db.GetCapabilitiesByFlowRow, error) {
					return nil, errors.New("service error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			assertBody: func(t *testing.T, body []byte) {
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if got["detail"] != "Failed to fetch capabilities" {
					t.Errorf("unexpected detail: %v", got["detail"])
				}
				if got["status"] != float64(http.StatusInternalServerError) {
					t.Errorf("unexpected status: %v", got["status"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/flows/"+tt.id+"/capabilities", nil)
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

			h.GetCapabilitiesByFlow(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if tt.assertBody != nil {
				tt.assertBody(t, w.Body.Bytes())
			}
		})
	}
}

func TestHandler_GetCapabilitiesByProduct(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockSetup      func(m *mockCapabilityService)
		expectedStatus int
		assertBody     func(t *testing.T, body []byte)
	}{
		{
			name: "Success",
			id:   "1",
			mockSetup: func(m *mockCapabilityService) {
				m.getCapabilitiesByProductFunc = func(ctx context.Context, id int) ([]db.Capability, error) {
					return []db.Capability{{ID: 1, Name: "Test Cap"}}, nil
				}
			},
			expectedStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var got []db.Capability
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if len(got) != 1 || got[0].ID != 1 || got[0].Name != "Test Cap" {
					t.Errorf("unexpected capabilities: %+v", got)
				}
			},
		},
		{
			name: "Product Not Found",
			id:   "999",
			mockSetup: func(m *mockCapabilityService) {
				m.getCapabilitiesByProductFunc = func(ctx context.Context, id int) ([]db.Capability, error) {
					return nil, internal.NewNotFoundError(999, "capability")
				}
			},
			expectedStatus: http.StatusNotFound,
			assertBody: func(t *testing.T, body []byte) {
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if got["detail"] != "Product not found" {
					t.Errorf("unexpected detail: %v", got["detail"])
				}
				if got["status"] != float64(http.StatusNotFound) {
					t.Errorf("unexpected status: %v", got["status"])
				}
			},
		},
		{
			name:           "Invalid ID",
			id:             "abc",
			mockSetup:      func(m *mockCapabilityService) {},
			expectedStatus: http.StatusBadRequest,
			assertBody: func(t *testing.T, body []byte) {
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if got["detail"] != "Invalid capability ID" {
					t.Errorf("unexpected detail: %v", got["detail"])
				}
				if got["status"] != float64(http.StatusBadRequest) {
					t.Errorf("unexpected status: %v", got["status"])
				}
			},
		},
		{
			name: "Service Error",
			id:   "1",
			mockSetup: func(m *mockCapabilityService) {
				m.getCapabilitiesByProductFunc = func(ctx context.Context, id int) ([]db.Capability, error) {
					return nil, errors.New("service error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			assertBody: func(t *testing.T, body []byte) {
				var got map[string]any
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if got["detail"] != "Failed to fetch capabilities" {
					t.Errorf("unexpected detail: %v", got["detail"])
				}
				if got["status"] != float64(http.StatusInternalServerError) {
					t.Errorf("unexpected status: %v", got["status"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/products/"+tt.id+"/capabilities", nil)
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

			h.GetCapabilitiesByProduct(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if tt.assertBody != nil {
				tt.assertBody(t, w.Body.Bytes())
			}
		})
	}
}
