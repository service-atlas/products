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

func TestUpdateFlowStepRequest_ToParams(t *testing.T) {

	tests := []struct {
		name     string
		req      updateFlowStepRequest
		existing db.FlowStep
		expected db.UpdateFlowStepParams
	}{
		{
			name: "Update All Fields",
			req: updateFlowStepRequest{
				Target:   new("https://new.com"),
				Protocol: new("HTTPS"),
			},
			existing: db.FlowStep{ID: 1, Target: pgtype.Text{String: "old", Valid: true}, Protocol: pgtype.Text{String: "HTTP", Valid: true}},
			expected: db.UpdateFlowStepParams{
				ID:       1,
				Target:   pgtype.Text{String: "https://new.com", Valid: true},
				Protocol: pgtype.Text{String: "HTTPS", Valid: true},
			},
		},
		{
			name: "Update Only Target",
			req: updateFlowStepRequest{
				Target: new("https://new.com"),
			},
			existing: db.FlowStep{ID: 1, Target: pgtype.Text{String: "old", Valid: true}, Protocol: pgtype.Text{String: "HTTP", Valid: true}},
			expected: db.UpdateFlowStepParams{
				ID:       1,
				Target:   pgtype.Text{String: "https://new.com", Valid: true},
				Protocol: pgtype.Text{String: "HTTP", Valid: true},
			},
		},
		{
			name: "Update Only Protocol",
			req: updateFlowStepRequest{
				Protocol: new("HTTPS"),
			},
			existing: db.FlowStep{ID: 1, Target: pgtype.Text{String: "old", Valid: true}, Protocol: pgtype.Text{String: "HTTP", Valid: true}},
			expected: db.UpdateFlowStepParams{
				ID:       1,
				Target:   pgtype.Text{String: "old", Valid: true},
				Protocol: pgtype.Text{String: "HTTPS", Valid: true},
			},
		},
		{
			name: "Explicitly Empty Protocol",
			req: updateFlowStepRequest{
				Protocol: new(""),
			},
			existing: db.FlowStep{ID: 1, Target: pgtype.Text{String: "old", Valid: true}, Protocol: pgtype.Text{String: "HTTP", Valid: true}},
			expected: db.UpdateFlowStepParams{
				ID:       1,
				Target:   pgtype.Text{String: "old", Valid: true},
				Protocol: pgtype.Text{String: "", Valid: true},
			},
		},
		{
			name:     "Omitted Fields (keep existing)",
			req:      updateFlowStepRequest{},
			existing: db.FlowStep{ID: 1, Target: pgtype.Text{String: "old", Valid: true}, Protocol: pgtype.Text{String: "HTTP", Valid: true}},
			expected: db.UpdateFlowStepParams{
				ID:       1,
				Target:   pgtype.Text{String: "old", Valid: true},
				Protocol: pgtype.Text{String: "HTTP", Valid: true},
			},
		},
		{
			name:     "No Fields Provided",
			req:      updateFlowStepRequest{Target: nil, Protocol: nil},
			existing: db.FlowStep{ID: 1, Target: pgtype.Text{String: "old", Valid: true}, Protocol: pgtype.Text{String: "HTTP", Valid: true}},
			expected: db.UpdateFlowStepParams{
				ID:       1,
				Target:   pgtype.Text{String: "old", Valid: true},
				Protocol: pgtype.Text{String: "HTTP", Valid: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := tt.req.ToParams(tt.existing.ID, tt.existing)
			if params.ID != tt.expected.ID {
				t.Errorf("expected ID %d, got %d", tt.expected.ID, params.ID)
			}
			if params.Target.String != tt.expected.Target.String {
				t.Errorf("expected target string %q, got %q", tt.expected.Target.String, params.Target.String)
			}
			if params.Target.Valid != tt.expected.Target.Valid {
				t.Errorf("expected target valid %v, got %v", tt.expected.Target.Valid, params.Target.Valid)
			}
			if params.Protocol.String != tt.expected.Protocol.String {
				t.Errorf("expected protocol string %q, got %q", tt.expected.Protocol.String, params.Protocol.String)
			}
			if params.Protocol.Valid != tt.expected.Protocol.Valid {
				t.Errorf("expected protocol valid %v, got %v", tt.expected.Protocol.Valid, params.Protocol.Valid)
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
				Protocol: "http",
				Target:   "target",
				Current:  validUUID1,
				Next:     validUUID2,
				FlowId:   1,
			},
			expectedError: false,
		},
		{
			name: "Invalid Current UUID",
			request: createFlowStepRequest{
				Protocol: "http",
				Target:   "target",
				Current:  "invalid",
				Next:     validUUID2,
				FlowId:   1,
			},
			expectedError: true,
		},
		{
			name: "Invalid Next UUID",
			request: createFlowStepRequest{
				Protocol: "http",
				Target:   "target",
				Current:  validUUID1,
				Next:     "invalid",
				FlowId:   1,
			},
			expectedError: true,
		},
		{
			name: "Empty Protocol and Target",
			request: createFlowStepRequest{
				Protocol: "",
				Target:   "",
				Current:  validUUID1,
				Next:     validUUID2,
				FlowId:   1,
			},
			expectedError: false,
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
				if params.Protocol.String != tt.request.Protocol {
					t.Errorf("expected Protocol %q, got %q", tt.request.Protocol, params.Protocol.String)
				}
				if params.Protocol.Valid != (tt.request.Protocol != "") {
					t.Errorf("expected Protocol valid %v, got %v", tt.request.Protocol != "", params.Protocol.Valid)
				}
				if params.Target.String != tt.request.Target {
					t.Errorf("expected Target %q, got %q", tt.request.Target, params.Target.String)
				}
				if params.Target.Valid != (tt.request.Target != "") {
					t.Errorf("expected Target valid %v, got %v", tt.request.Target != "", params.Target.Valid)
				}
				if !params.Timestamp.Valid {
					t.Error("expected Timestamp to be valid")
				}
			}
		})
	}
}
