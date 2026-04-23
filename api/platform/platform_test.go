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
)

type mockPlatformQuerier struct {
	err            error
	createPlatform func(ctx context.Context, arg db.CreatePlatformParams) error
}

func (m *mockPlatformQuerier) CreatePlatform(ctx context.Context, arg db.CreatePlatformParams) error {
	if m.createPlatform != nil {
		return m.createPlatform(ctx, arg)
	}
	return m.err
}

func (m *mockPlatformQuerier) DeletePlatform(ctx context.Context, id int32) error {
	return m.err
}

func (m *mockPlatformQuerier) GetPlatform(ctx context.Context, id int32) (db.Platform, error) {
	return db.Platform{}, m.err
}

func (m *mockPlatformQuerier) GetPlatforms(ctx context.Context) ([]db.Platform, error) {
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
