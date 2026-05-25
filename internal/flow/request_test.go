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
					String: "Original Description",
					Valid:  true,
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
					String: "Original Description",
					Valid:  true,
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
