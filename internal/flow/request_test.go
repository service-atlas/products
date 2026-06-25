package flow

import (
	"products/internal/flow/db"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestUpdateFlowRequest_ToParams(t *testing.T) {
	existing := db.Flow{
		ID:   1,
		Name: "Original Name",
		Description: pgtype.Text{
			String: "Original Description",
			Valid:  true,
		},
	}

	tests := []struct {
		name     string
		request  updateFlowRequest
		id       int
		expected db.UpdateFlowParams
	}{
		{
			name: "Update Name Only",
			request: updateFlowRequest{
				Name: "New Name",
			},
			id: 1,
			expected: db.UpdateFlowParams{
				ID:   1,
				Name: "New Name",
				Description: pgtype.Text{
					String: "",
					Valid:  false, // Omitted description becomes NULL
				},
			},
		},
		{
			name: "Update Description Only",
			request: updateFlowRequest{
				Description: "New Description",
			},
			id: 1,
			expected: db.UpdateFlowParams{
				ID:   1,
				Name: "Original Name",
				Description: pgtype.Text{
					String: "New Description",
					Valid:  true,
				},
			},
		},
		{
			name: "Update Both",
			request: updateFlowRequest{
				Name:        "New Name",
				Description: "New Description",
			},
			id: 1,
			expected: db.UpdateFlowParams{
				ID:   1,
				Name: "New Name",
				Description: pgtype.Text{
					String: "New Description",
					Valid:  true,
				},
			},
		},
		{
			name:    "Update None (keep existing)",
			request: updateFlowRequest{},
			id:      1,
			expected: db.UpdateFlowParams{
				ID:   1,
				Name: "Original Name",
				Description: pgtype.Text{
					String: "",
					Valid:  false, // Omitted description becomes NULL
				},
			},
		},
		{
			name: "Explicit Empty Description Overwrites to NULL",
			request: updateFlowRequest{
				Description: "",
			},
			id: 1,
			expected: db.UpdateFlowParams{
				ID:   1,
				Name: "Original Name",
				Description: pgtype.Text{
					String: "",
					Valid:  false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := tt.request.ToParams(tt.id, existing)

			if params.ID != tt.expected.ID {
				t.Errorf("expected ID %d, got %d", tt.expected.ID, params.ID)
			}
			if params.Name != tt.expected.Name {
				t.Errorf("expected Name %q, got %q", tt.expected.Name, params.Name)
			}
			if params.Description.String != tt.expected.Description.String {
				t.Errorf("expected Description string %q, got %q", tt.expected.Description.String, params.Description.String)
			}
			if params.Description.Valid != tt.expected.Description.Valid {
				t.Errorf("expected Description valid %v, got %v", tt.expected.Description.Valid, params.Description.Valid)
			}
			if !params.UpdatedAt.Valid {
				t.Error("expected UpdatedAt to be valid")
			}
		})
	}
}

func TestToPgUUID(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedValid bool
		expectedError bool
	}{
		{
			name:          "Valid UUID",
			input:         "550e8400-e29b-41d4-a716-446655440000",
			expectedValid: true,
			expectedError: false,
		},
		{
			name:          "Invalid UUID",
			input:         "invalid-uuid",
			expectedValid: false,
			expectedError: true,
		},
		{
			name:          "Empty UUID",
			input:         "",
			expectedValid: false,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toPgUUID(tt.input)
			if (err != nil) != tt.expectedError {
				t.Errorf("toPgUUID() error = %v, expectedError %v", err, tt.expectedError)
				return
			}
			if got.Valid != tt.expectedValid {
				t.Errorf("toPgUUID() valid = %v, expected %v", got.Valid, tt.expectedValid)
			}
		})
	}
}

func TestCreateFlowStepRequest_ToParams(t *testing.T) {
	validUUID1 := "550e8400-e29b-41d4-a716-446655440000"
	validUUID2 := "67123456-e29b-41d4-a716-446655440000"

	tests := []struct {
		name          string
		request       createFlowStepRequest
		expectedError bool
	}{
		{
			name: "Success",
			request: createFlowStepRequest{
				Current: validUUID1,
				Next:    validUUID2,
				FlowId:  1,
			},
			expectedError: false,
		},
		{
			name: "Invalid Current UUID",
			request: createFlowStepRequest{
				Current: "invalid",
				Next:    validUUID2,
				FlowId:  1,
			},
			expectedError: true,
		},
		{
			name: "Invalid Next UUID",
			request: createFlowStepRequest{
				Current: validUUID1,
				Next:    "invalid",
				FlowId:  1,
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := tt.request.ToParams()
			if (err != nil) != tt.expectedError {
				t.Errorf("ToParams() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			if !tt.expectedError {
				if params.FlowID != tt.request.FlowId {
					t.Errorf("expected FlowID %d, got %d", tt.request.FlowId, params.FlowID)
				}
				if !params.Timestamp.Valid {
					t.Error("expected Timestamp to be valid")
				}
			}
		})
	}
}
