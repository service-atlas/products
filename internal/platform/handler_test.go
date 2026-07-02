package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"products/internal"
	"products/internal/platform/db"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

type mockPlatformService struct {
	createPlatform func(ctx context.Context, req createPlatformRequest) (db.Platform, error)
	getPlatform    func(ctx context.Context, id int) (db.Platform, error)
	getPlatforms   func(ctx context.Context) ([]db.Platform, error)
	updatePlatform func(ctx context.Context, req updatePlatformRequest, id int) (int, error)
	deletePlatform func(ctx context.Context, id int) (int, error)
}

func (m *mockPlatformService) CreatePlatform(ctx context.Context, req createPlatformRequest) (db.Platform, error) {
	if m.createPlatform != nil {
		return m.createPlatform(ctx, req)
	}
	return db.Platform{}, nil
}

func (m *mockPlatformService) GetPlatform(ctx context.Context, id int) (db.Platform, error) {
	if m.getPlatform != nil {
		return m.getPlatform(ctx, id)
	}
	return db.Platform{}, nil
}

func (m *mockPlatformService) GetPlatforms(ctx context.Context) ([]db.Platform, error) {
	if m.getPlatforms != nil {
		return m.getPlatforms(ctx)
	}
	return nil, nil
}

func (m *mockPlatformService) UpdatePlatform(ctx context.Context, req updatePlatformRequest, id int) (int, error) {
	if m.updatePlatform != nil {
		return m.updatePlatform(ctx, req, id)
	}
	return 0, nil
}

func (m *mockPlatformService) DeletePlatform(ctx context.Context, id int) (int, error) {
	if m.deletePlatform != nil {
		return m.deletePlatform(ctx, id)
	}
	return 0, nil
}

func TestCreatePlatform(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    any
		dbErr          error
		expectedStatus int
	}{
		{
			name: "Success",
			requestBody: createPlatformRequest{
				Name:        "Test Platform",
				Description: "Test Description",
			},
			dbErr:          nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Missing Name",
			requestBody: createPlatformRequest{
				Name:        "",
				Description: "Test Description",
			},
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			requestBody:    "not a json",
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "DB Error",
			requestBody: createPlatformRequest{
				Name:        "Test Platform",
				Description: "Test Description",
			},
			dbErr:          errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mSvc := &mockPlatformService{
				createPlatform: func(ctx context.Context, req createPlatformRequest) (db.Platform, error) {
					return db.Platform{}, tt.dbErr
				},
			}
			h := &handler{service: mSvc}

			var body []byte
			if s, ok := tt.requestBody.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/platforms", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			h.CreatePlatform(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestGetPlatforms(t *testing.T) {
	tests := []struct {
		name           string
		dbErr          error
		platforms      []db.Platform
		expectedStatus int
	}{
		{
			name: "Success",
			platforms: []db.Platform{
				{ID: 1, Name: "Platform 1", Description: pgtype.Text{String: "Desc 1", Valid: true}},
				{ID: 2, Name: "Platform 2", Description: pgtype.Text{String: "Desc 2", Valid: true}},
			},
			dbErr:          nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty Success",
			platforms:      []db.Platform{},
			dbErr:          nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "DB Error",
			platforms:      nil,
			dbErr:          errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mSvc := &mockPlatformService{
				getPlatforms: func(ctx context.Context) ([]db.Platform, error) {
					return tt.platforms, tt.dbErr
				},
			}
			h := &handler{service: mSvc}

			req := httptest.NewRequest(http.MethodGet, "/api/platforms", nil)
			rr := httptest.NewRecorder()

			h.GetPlatforms(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var got []db.Platform
				err := json.Unmarshal(rr.Body.Bytes(), &got)
				if err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if len(got) != len(tt.platforms) {
					t.Errorf("expected %d platforms, got %d", len(tt.platforms), len(got))
				}
			}
		})
	}
}

func TestGetPlatform(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		dbPlatform     db.Platform
		dbErr          error
		expectedStatus int
	}{
		{
			name:           "Success",
			id:             "1",
			dbPlatform:     db.Platform{ID: 1, Name: "Platform 1"},
			dbErr:          nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid ID",
			id:             "abc",
			dbPlatform:     db.Platform{},
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid ID (Zero)",
			id:             "0",
			dbPlatform:     db.Platform{},
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid ID (Negative)",
			id:             "-1",
			dbPlatform:     db.Platform{},
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Not Found",
			id:             "999",
			dbPlatform:     db.Platform{},
			dbErr:          internal.NewNotFoundError(999, "Platform"),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "DB Error",
			id:             "1",
			dbPlatform:     db.Platform{},
			dbErr:          errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mSvc := &mockPlatformService{
				getPlatform: func(ctx context.Context, id int) (db.Platform, error) {
					return tt.dbPlatform, tt.dbErr
				},
			}
			h := &handler{service: mSvc}

			req := httptest.NewRequest(http.MethodGet, "/api/platforms/"+tt.id, nil)
			req.SetPathValue("id", tt.id)
			rr := httptest.NewRecorder()

			h.GetPlatform(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var got db.Platform
				err := json.Unmarshal(rr.Body.Bytes(), &got)
				if err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if got.ID != tt.dbPlatform.ID || got.Name != tt.dbPlatform.Name {
					t.Errorf("expected platform %+v, got %+v", tt.dbPlatform, got)
				}
			}
		})
	}
}

func TestDeletePlatform(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		dbErr          error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success",
			id:             "1",
			dbErr:          nil,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Invalid ID",
			id:             "abc",
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "DB Error",
			id:             "1",
			dbErr:          errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error",
		},
		{
			name:           "DB Error (pgx.ErrNoRows)",
			id:             "1",
			dbErr:          internal.NewNotFoundError(1, "Platform"),
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Platform not found",
		},
		{
			name:           "Invalid ID (Zero)",
			id:             "0",
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid ID (Negative)",
			id:             "-1",
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Not Found",
			id:             "999",
			dbErr:          internal.NewNotFoundError(999, "Platform"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mSvc := &mockPlatformService{
				deletePlatform: func(ctx context.Context, id int) (int, error) {
					return id, tt.dbErr
				},
			}
			h := &handler{service: mSvc}

			req := httptest.NewRequest(http.MethodDelete, "/api/platforms/"+tt.id, nil)
			req.SetPathValue("id", tt.id)
			rr := httptest.NewRecorder()

			h.DeletePlatform(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedBody != "" {
				if rr.Code >= 400 {
					var envelope internal.ErrorEnvelope
					if err := json.NewDecoder(rr.Body).Decode(&envelope); err != nil {
						t.Fatalf("failed to decode error response: %v", err)
					}
					if envelope.Detail != tt.expectedBody {
						t.Errorf("expected detail %q, got %q", tt.expectedBody, envelope.Detail)
					}
				} else if rr.Body.String() != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, rr.Body.String())
				}
			}
		})
	}
}

func TestUpdatePlatform(t *testing.T) {
	tests := []struct {
		name           string
		pathID         string
		requestBody    any
		dbErr          error
		expectedStatus int
	}{
		{
			name:   "Success",
			pathID: "1",
			requestBody: db.Platform{
				ID:          1,
				Name:        "Updated Platform",
				Description: pgtype.Text{String: "Updated Description", Valid: true},
			},
			dbErr:          nil,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "Invalid Path ID",
			pathID: "abc",
			requestBody: db.Platform{
				ID:   1,
				Name: "Updated Platform",
			},
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Invalid Path ID (Zero)",
			pathID: "0",
			requestBody: db.Platform{
				ID:   0,
				Name: "Updated Platform",
			},
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Invalid Path ID (Negative)",
			pathID: "-1",
			requestBody: db.Platform{
				ID:   -1,
				Name: "Updated Platform",
			},
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Invalid Path ID (Overflow)",
			pathID: "2147483648",
			requestBody: db.Platform{
				ID:   1,
				Name: "Updated Platform",
			},
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "ID Mismatch",
			pathID: "2",
			requestBody: db.Platform{
				ID:   1,
				Name: "Updated Platform",
			},
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Missing Name",
			pathID: "1",
			requestBody: db.Platform{
				ID:   1,
				Name: "",
			},
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			pathID:         "1",
			requestBody:    "not a json",
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Platform Not Found",
			pathID: "999",
			requestBody: db.Platform{
				ID:   999,
				Name: "Non-existent",
			},
			dbErr:          internal.NewNotFoundError(999, "Platform"),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "DB Error",
			pathID: "1",
			requestBody: db.Platform{
				ID:   1,
				Name: "Test Platform",
			},
			dbErr:          errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mSvc := &mockPlatformService{
				updatePlatform: func(ctx context.Context, req updatePlatformRequest, id int) (int, error) {
					if tt.name == "Success" {
						expectedBody := tt.requestBody.(db.Platform)
						if id != expectedBody.ID {
							t.Errorf("expected ID %d, got %d", expectedBody.ID, id)
						}
						if req.Name != expectedBody.Name {
							t.Errorf("expected Name %s, got %s", expectedBody.Name, req.Name)
						}
						if req.Description != expectedBody.Description.String {
							t.Errorf("expected Description %s, got %s", expectedBody.Description.String, req.Description)
						}
					}
					return id, tt.dbErr
				},
			}
			h := &handler{service: mSvc}

			var body []byte
			if s, ok := tt.requestBody.(string); ok {
				body = []byte(s)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPut, "/api/platforms/"+tt.pathID, bytes.NewBuffer(body))
			if tt.pathID != "" {
				req.SetPathValue("id", tt.pathID)
			}
			rr := httptest.NewRecorder()

			h.UpdatePlatform(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}
