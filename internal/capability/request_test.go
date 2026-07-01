package capability

import (
	"testing"
)

func TestCreateCapabilityRequest_ToParams(t *testing.T) {
	req := createCapabilityRequest{
		FlowId:      1,
		Name:        "Test Cap",
		Description: "Test Description",
	}

	params := req.ToParams()

	if params.FlowID != req.FlowId {
		t.Errorf("expected FlowID %d, got %d", req.FlowId, params.FlowID)
	}
	if params.Name != req.Name {
		t.Errorf("expected Name %s, got %s", req.Name, params.Name)
	}
	if !params.Description.Valid || params.Description.String != req.Description {
		t.Errorf("expected Description %s, got %v", req.Description, params.Description)
	}
	if !params.Timestamp.Valid {
		t.Error("expected Timestamp to be valid")
	}
}

func TestCreateCapabilityRequest_ToParams_EmptyDescription(t *testing.T) {
	req := createCapabilityRequest{
		FlowId: 1,
		Name:   "Test Cap",
	}

	params := req.ToParams()

	if params.Description.Valid {
		t.Error("expected Description to be invalid for empty string")
	}
}

func TestCreateCapabilityRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     createCapabilityRequest
		wantErr bool
	}{
		{
			name: "Valid request",
			req: createCapabilityRequest{
				Name:   "Test",
				FlowId: 1,
			},
			wantErr: false,
		},
		{
			name: "Missing name",
			req: createCapabilityRequest{
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "Missing flow id",
			req: createCapabilityRequest{
				Name: "Test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateCapabilityRequest_ToParams(t *testing.T) {
	id := 1
	req := updateCapabilityRequest{
		Name:        "Test Cap",
		Description: "Test Description",
	}

	params := req.ToParams(id)

	if params.ID != id {
		t.Errorf("expected ID %d, got %d", id, params.ID)
	}
	if params.Name != req.Name {
		t.Errorf("expected Name %s, got %s", req.Name, params.Name)
	}
	if !params.Description.Valid || params.Description.String != req.Description {
		t.Errorf("expected Description %s, got %v", req.Description, params.Description)
	}
	if !params.UpdatedAt.Valid {
		t.Error("expected UpdatedAt to be valid")
	}
}

func TestUpdateCapabilityRequest_ToParams_EmptyDescription(t *testing.T) {
	req := updateCapabilityRequest{
		Name: "Test Cap",
	}

	params := req.ToParams(1)

	if params.Description.Valid {
		t.Error("expected Description to be invalid for empty string")
	}
}
