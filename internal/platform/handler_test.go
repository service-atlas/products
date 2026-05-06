package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockPlatformQuerier struct {
	err            error
	createPlatform func(ctx context.Context, arg CreatePlatformParams) error
	getPlatforms   func(ctx context.Context) ([]Platform, error)
	getPlatform    func(ctx context.Context, id int32) (Platform, error)
	deletePlatform func(ctx context.Context, id int32) (int32, error)
	updatePlatform func(ctx context.Context, arg UpdatePlatformParams) (int32, error)
}

func (m *mockPlatformQuerier) CreatePlatform(ctx context.Context, arg CreatePlatformParams) error {
	if m.createPlatform != nil {
		return m.createPlatform(ctx, arg)
	}
	return m.err
}

func (m *mockPlatformQuerier) DeletePlatform(ctx context.Context, id int32) (int32, error) {
	if m.deletePlatform != nil {
		return m.deletePlatform(ctx, id)
	}
	return -1, m.err
}

func (m *mockPlatformQuerier) GetPlatform(ctx context.Context, id int32) (Platform, error) {
	if m.getPlatform != nil {
		return m.getPlatform(ctx, id)
	}
	return Platform{}, m.err
}

func (m *mockPlatformQuerier) GetPlatforms(ctx context.Context) ([]Platform, error) {
	if m.getPlatforms != nil {
		return m.getPlatforms(ctx)
	}
	return nil, m.err
}

func (m *mockPlatformQuerier) UpdatePlatform(ctx context.Context, arg UpdatePlatformParams) (int32, error) {
	if m.updatePlatform != nil {
		return m.updatePlatform(ctx, arg)
	}
	return -1, m.err
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
			requestBody: CreatePlatformRequest{
				Name:        "Test Platform",
				Description: "Test Description",
			},
			dbErr:          nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Missing Name",
			requestBody: CreatePlatformRequest{
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
			requestBody: CreatePlatformRequest{
				Name:        "Test Platform",
				Description: "Test Description",
			},
			dbErr:          errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mDB := &mockPlatformQuerier{err: tt.dbErr}
			h := &platformHandler{queries: mDB}

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
		platforms      []Platform
		expectedStatus int
	}{
		{
			name: "Success",
			platforms: []Platform{
				{ID: 1, Name: "Platform 1", Description: pgtype.Text{String: "Desc 1", Valid: true}},
				{ID: 2, Name: "Platform 2", Description: pgtype.Text{String: "Desc 2", Valid: true}},
			},
			dbErr:          nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty Success",
			platforms:      []Platform{},
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
			mDB := &mockPlatformQuerier{
				err: tt.dbErr,
				getPlatforms: func(ctx context.Context) ([]Platform, error) {
					return tt.platforms, tt.dbErr
				},
			}
			h := &platformHandler{queries: mDB}

			req := httptest.NewRequest(http.MethodGet, "/api/platforms", nil)
			rr := httptest.NewRecorder()

			h.GetPlatforms(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var got []Platform
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
		dbPlatform     Platform
		dbErr          error
		expectedStatus int
	}{
		{
			name:           "Success",
			id:             "1",
			dbPlatform:     Platform{ID: 1, Name: "Platform 1"},
			dbErr:          nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid ID",
			id:             "abc",
			dbPlatform:     Platform{},
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid ID (Zero)",
			id:             "0",
			dbPlatform:     Platform{},
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid ID (Negative)",
			id:             "-1",
			dbPlatform:     Platform{},
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid ID (Overflow)",
			id:             "2147483648",
			dbPlatform:     Platform{},
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Not Found",
			id:             "999",
			dbPlatform:     Platform{},
			dbErr:          pgx.ErrNoRows,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "DB Error",
			id:             "1",
			dbPlatform:     Platform{},
			dbErr:          errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mDB := &mockPlatformQuerier{
				err: tt.dbErr,
				getPlatform: func(ctx context.Context, id int32) (Platform, error) {
					return tt.dbPlatform, tt.dbErr
				},
			}
			h := &platformHandler{queries: mDB}

			req := httptest.NewRequest(http.MethodGet, "/api/platforms/"+tt.id, nil)
			req.SetPathValue("id", tt.id)
			rr := httptest.NewRecorder()

			h.GetPlatform(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var got Platform
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
			name:           "Invalid ID (Overflow)",
			id:             "2147483648",
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Not Found",
			id:             "999",
			dbErr:          pgx.ErrNoRows,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mDB := &mockPlatformQuerier{
				err: tt.dbErr,
				deletePlatform: func(ctx context.Context, id int32) (int32, error) {
					if i, e := strconv.Atoi(tt.id); e == nil {
						return int32(i), tt.dbErr
					}
					return -1, tt.dbErr
				},
			}
			h := &platformHandler{queries: mDB}

			req := httptest.NewRequest(http.MethodDelete, "/api/platforms/"+tt.id, nil)
			req.SetPathValue("id", tt.id)
			rr := httptest.NewRecorder()

			h.DeletePlatform(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
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
			requestBody: Platform{
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
			requestBody: Platform{
				ID:   1,
				Name: "Updated Platform",
			},
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Invalid Path ID (Zero)",
			pathID: "0",
			requestBody: Platform{
				ID:   0,
				Name: "Updated Platform",
			},
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Invalid Path ID (Negative)",
			pathID: "-1",
			requestBody: Platform{
				ID:   -1,
				Name: "Updated Platform",
			},
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Invalid Path ID (Overflow)",
			pathID: "2147483648",
			requestBody: Platform{
				ID:   1,
				Name: "Updated Platform",
			},
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "ID Mismatch",
			pathID: "2",
			requestBody: Platform{
				ID:   1,
				Name: "Updated Platform",
			},
			dbErr:          nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Missing Name",
			pathID: "1",
			requestBody: Platform{
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
			requestBody: Platform{
				ID:   999,
				Name: "Non-existent",
			},
			dbErr:          pgx.ErrNoRows,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "DB Error",
			pathID: "1",
			requestBody: Platform{
				ID:   1,
				Name: "Test Platform",
			},
			dbErr:          errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mDB := &mockPlatformQuerier{
				err: tt.dbErr,
				updatePlatform: func(ctx context.Context, arg UpdatePlatformParams) (int32, error) {
					if tt.name == "Success" {
						expectedBody := tt.requestBody.(Platform)
						if arg.ID != expectedBody.ID {
							t.Errorf("expected ID %d, got %d", expectedBody.ID, arg.ID)
						}
						if arg.Name != expectedBody.Name {
							t.Errorf("expected Name %s, got %s", expectedBody.Name, arg.Name)
						}
						if arg.Description.String != expectedBody.Description.String {
							t.Errorf("expected Description %s, got %s", expectedBody.Description.String, arg.Description.String)
						}
						if !arg.Updatedat.Valid {
							t.Error("expected UpdatedAt to be valid")
						}
						if arg.Updatedat.Time.IsZero() {
							t.Error("expected UpdatedAt to be non-zero")
						}
					}
					return arg.ID, tt.dbErr
				},
			}
			h := &platformHandler{queries: mDB}

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
