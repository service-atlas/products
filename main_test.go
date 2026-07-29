package main

import (
	"context"
	"errors"
	"testing"

	"github.com/service-atlas/secrets-provider" //nolint:depguard
)

type mockProvider struct {
	dbInfo *secretsprovider.DatabaseInfo
	err    error
}

func (m *mockProvider) GetDatabaseInfo(_ context.Context) (secretsprovider.DatabaseInfo, error) {
	if m.err != nil {
		return secretsprovider.DatabaseInfo{}, m.err
	}
	return *m.dbInfo, nil
}

func (m *mockProvider) GetSecret(_ context.Context, _ string) (string, error) {
	return "", nil
}

func TestGetConnStr(t *testing.T) {
	tests := []struct {
		name        string
		dbInfo      *secretsprovider.DatabaseInfo
		err         error
		expected    string
		expectedErr bool
	}{
		{
			name: "All variables set without scheme",
			dbInfo: &secretsprovider.DatabaseInfo{
				Username: "user",
				Password: "pass",
				URL:      "localhost:5432/dbname",
			},
			expected:    "postgres://user:pass@localhost:5432/dbname",
			expectedErr: false,
		},
		{
			name: "All variables set with scheme",
			dbInfo: &secretsprovider.DatabaseInfo{
				Username: "user",
				Password: "pass",
				URL:      "postgres://localhost:5432/dbname",
			},
			expected:    "postgres://user:pass@localhost:5432/dbname",
			expectedErr: false,
		},
		{
			name:        "Error from provider",
			err:         errors.New("provider error"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mProvider := &mockProvider{
				dbInfo: tt.dbInfo,
				err:    tt.err,
			}

			got, err := getConnStr(t.Context(), mProvider)
			if (err != nil) != tt.expectedErr {
				t.Errorf("getConnStr() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}
			if got != tt.expected {
				t.Errorf("getConnStr() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
