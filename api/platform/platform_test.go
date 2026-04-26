package platformHandler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"products/internal/db"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockPlatformQuerier struct {
	err            error
	createPlatform func(ctx context.Context, arg db.CreatePlatformParams) error
	getPlatforms   func(ctx context.Context) ([]db.Platform, error)
	getPlatform    func(ctx context.Context, id int32) (db.Platform, error)
	deletePlatform func(ctx context.Context, id int32) error
}

func (m *mockPlatformQuerier) CreatePlatform(ctx context.Context, arg db.CreatePlatformParams) error {
	if m.createPlatform != nil {
		return m.createPlatform(ctx, arg)
	}
	return m.err
}

func (m *mockPlatformQuerier) DeletePlatform(ctx context.Context, id int32) error {
	if m.deletePlatform != nil {
		return m.deletePlatform(ctx, id)
	}
	return m.err
}

func (m *mockPlatformQuerier) GetPlatform(ctx context.Context, id int32) (db.Platform, error) {
	if m.getPlatform != nil {
		return m.getPlatform(ctx, id)
	}
	return db.Platform{}, m.err
}

func (m *mockPlatformQuerier) GetPlatforms(ctx context.Context) ([]db.Platform, error) {
	if m.getPlatforms != nil {
		return m.getPlatforms(ctx)
	}
	return nil, m.err
}

func (m *mockPlatformQuerier) UpdatePlatform(ctx context.Context, arg db.UpdatePlatformParams) error {
	return m.err
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
			h := NewPlatformHandler(mDB)

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
			mDB := &mockPlatformQuerier{
				err: tt.dbErr,
				getPlatforms: func(ctx context.Context) ([]db.Platform, error) {
					return tt.platforms, tt.dbErr
				},
			}
			h := NewPlatformHandler(mDB)

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
			name:           "Not Found",
			id:             "999",
			dbPlatform:     db.Platform{},
			dbErr:          pgx.ErrNoRows,
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
			mDB := &mockPlatformQuerier{
				err: tt.dbErr,
				getPlatform: func(ctx context.Context, id int32) (db.Platform, error) {
					return tt.dbPlatform, tt.dbErr
				},
			}
			h := NewPlatformHandler(mDB)

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mDB := &mockPlatformQuerier{
				err: tt.dbErr,
				deletePlatform: func(ctx context.Context, id int32) error {
					return tt.dbErr
				},
			}
			h := NewPlatformHandler(mDB)

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
