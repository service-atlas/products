package capability

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"products/internal/capability/db"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type mockCapabilityService struct {
	createCapabilityFunc         func(ctx context.Context, req createCapabilityRequest) (db.Capability, error)
	getCapabilityFunc            func(ctx context.Context, id int) (db.Capability, error)
	getCapabilitiesByProductFunc func(ctx context.Context, id int) ([]db.GetCapabilitiesByProductRow, error)
	getCapabilitiesByFlowFunc    func(ctx context.Context, id int) ([]db.Capability, error)
}

func (m *mockCapabilityService) CreateCapability(ctx context.Context, req createCapabilityRequest) (db.Capability, error) {
	return m.createCapabilityFunc(ctx, req)
}

func (m *mockCapabilityService) GetCapability(ctx context.Context, id int) (db.Capability, error) {
	return m.getCapabilityFunc(ctx, id)
}

func (m *mockCapabilityService) GetCapabilitiesByProduct(ctx context.Context, id int) ([]db.GetCapabilitiesByProductRow, error) {
	return m.getCapabilitiesByProductFunc(ctx, id)
}

func (m *mockCapabilityService) GetCapabilitiesByFlow(ctx context.Context, id int) ([]db.Capability, error) {
	return m.getCapabilitiesByFlowFunc(ctx, id)
}

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
				FlowId: 1,
				Name:   "Test Cap",
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
				Name:   "Test Cap",
				FlowId: 1,
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

			h := &capabilityHandler{service: mockSvc}
			h.CreateCapability(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

type mockDBTX struct {
	db.DBTX
}

func (m *mockDBTX) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (m *mockDBTX) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return nil, nil
}
func (m *mockDBTX) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	return nil
}

func TestNewHandler(t *testing.T) {
	mockDB := &mockDBTX{}
	h := NewHandler(mockDB)
	if h == nil {
		t.Error("expected handler to be non-nil")
	}
}

func TestHandler_GetCapability(t *testing.T) {
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
				m.getCapabilityFunc = func(ctx context.Context, id int) (db.Capability, error) {
					return db.Capability{ID: 1, Name: "Test Cap"}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Not Found",
			id:   "999",
			mockSetup: func(m *mockCapabilityService) {
				m.getCapabilityFunc = func(ctx context.Context, id int) (db.Capability, error) {
					return db.Capability{}, NotFoundError{Msg: "Capability not found"}
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid ID",
			id:             "abc",
			mockSetup:      func(m *mockCapabilityService) {},
			expectedStatus: http.StatusBadRequest,
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
			h := &capabilityHandler{service: mockSvc}

			// We might need to mock internal.GetIntFromRequestPath or use a router that sets it.
			// For now, let's assume it works if we set it in the request context or if the router does it.
			// Many Go routers use context for path variables.

			h.GetCapability(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
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
	}{
		{
			name: "Success",
			id:   "1",
			mockSetup: func(m *mockCapabilityService) {
				m.getCapabilitiesByFlowFunc = func(ctx context.Context, id int) ([]db.Capability, error) {
					return []db.Capability{{ID: 1, Name: "Test Cap"}}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Flow Not Found",
			id:   "999",
			mockSetup: func(m *mockCapabilityService) {
				m.getCapabilitiesByFlowFunc = func(ctx context.Context, id int) ([]db.Capability, error) {
					return nil, NotFoundError{Msg: "Flow not found"}
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid ID",
			id:             "abc",
			mockSetup:      func(m *mockCapabilityService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Service Error",
			id:   "1",
			mockSetup: func(m *mockCapabilityService) {
				m.getCapabilitiesByFlowFunc = func(ctx context.Context, id int) ([]db.Capability, error) {
					return nil, errors.New("service error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
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
			h := &capabilityHandler{service: mockSvc}

			h.GetCapabilitiesByFlow(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
