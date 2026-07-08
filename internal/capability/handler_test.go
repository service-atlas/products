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
	updateCapabilityFunc         func(ctx context.Context, req updateCapabilityRequest) (db.Capability, error)
	deleteCapabilityFunc         func(ctx context.Context, id int) error
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

func (m *mockCapabilityService) UpdateCapability(ctx context.Context, req updateCapabilityRequest) (db.Capability, error) {
	return m.updateCapabilityFunc(ctx, req)
}

func (m *mockCapabilityService) DeleteCapability(ctx context.Context, id int) error {
	return m.deleteCapabilityFunc(ctx, id)
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

			h := &handler{service: mockSvc}
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
	mockService := &mockCapabilityService{}
	h := newHandler(mockService)
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
					return db.Capability{}, NotFoundError{Msg: "Capability not found"}
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
				m.getCapabilitiesByFlowFunc = func(ctx context.Context, id int) ([]db.Capability, error) {
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
			name: "Flow Not Found",
			id:   "999",
			mockSetup: func(m *mockCapabilityService) {
				m.getCapabilitiesByFlowFunc = func(ctx context.Context, id int) ([]db.Capability, error) {
					return nil, NotFoundError{Msg: "Flow not found"}
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
				m.getCapabilitiesByFlowFunc = func(ctx context.Context, id int) ([]db.Capability, error) {
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
				m.getCapabilitiesByProductFunc = func(ctx context.Context, id int) ([]db.GetCapabilitiesByProductRow, error) {
					return []db.GetCapabilitiesByProductRow{{ID: 1, Name: "Test Cap", FlowName: "Flow A"}}, nil
				}
			},
			expectedStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var got []db.GetCapabilitiesByProductRow
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if len(got) != 1 || got[0].ID != 1 || got[0].Name != "Test Cap" || got[0].FlowName != "Flow A" {
					t.Errorf("unexpected capabilities: %+v", got)
				}
			},
		},
		{
			name: "Product Not Found",
			id:   "999",
			mockSetup: func(m *mockCapabilityService) {
				m.getCapabilitiesByProductFunc = func(ctx context.Context, id int) ([]db.GetCapabilitiesByProductRow, error) {
					return nil, NotFoundError{Msg: "Product not found"}
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
				m.getCapabilitiesByProductFunc = func(ctx context.Context, id int) ([]db.GetCapabilitiesByProductRow, error) {
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
			expectedStatus: http.StatusOK,
		},
		{
			name: "Not Found",
			id:   "999",
			mockSetup: func(m *mockCapabilityService) {
				m.deleteCapabilityFunc = func(ctx context.Context, id int) error {
					return NotFoundError{Msg: "Capability not found"}
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
