package main

import (
	"testing"
)

func TestGetConnStr(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		expected    string
		expectedErr bool
	}{
		{
			name: "All variables set without scheme",
			env: map[string]string{
				"DB_USERNAME": "user",
				"DB_PASSWORD": "pass",
				"DB_URL":      "localhost:5432/dbname",
			},
			expected:    "postgres://user:pass@localhost:5432/dbname",
			expectedErr: false,
		},
		{
			name: "All variables set with scheme",
			env: map[string]string{
				"DB_USERNAME": "user",
				"DB_PASSWORD": "pass",
				"DB_URL":      "postgres://localhost:5432/dbname",
			},
			expected:    "postgres://user:pass@localhost:5432/dbname",
			expectedErr: false,
		},
		{
			name: "Missing DB_USERNAME",
			env: map[string]string{
				"DB_PASSWORD": "pass",
				"DB_URL":      "localhost:5432/dbname",
			},
			expectedErr: true,
		},
		{
			name: "Missing DB_PASSWORD",
			env: map[string]string{
				"DB_USERNAME": "user",
				"DB_URL":      "localhost:5432/dbname",
			},
			expectedErr: true,
		},
		{
			name: "Missing DB_URL",
			env: map[string]string{
				"DB_USERNAME": "user",
				"DB_PASSWORD": "pass",
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env first to ensure isolation
			t.Setenv("DB_USERNAME", "")
			t.Setenv("DB_PASSWORD", "")
			t.Setenv("DB_URL", "")

			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			got, err := getConnStr()
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
