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

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type mockCapabilityService struct {
	createCapabilityFunc func(ctx context.Context, req createCapabilityRequest) (db.Capability, error)
}

func (m *mockCapabilityService) CreateCapability(ctx context.Context, req createCapabilityRequest) (db.Capability, error) {
	return m.createCapabilityFunc(ctx, req)
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
				Name: "Test Cap",
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
