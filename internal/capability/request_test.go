package capability

import (
	"testing"
)

func TestCreateCapabilityRequest_ToParams(t *testing.T) {
	req := createCapabilityRequest{
		ProductId:   1,
		Name:        "Test Cap",
		Description: "Test Description",
	}

	params := req.ToParams()

	if params.ProductID != req.ProductId {
		t.Errorf("expected ProductId %d, got %d", req.ProductId, params.ProductID)
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
		ProductId: 1,
		Name:      "Test Cap",
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
				Name:      "Test",
				ProductId: 1,
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
			name: "Missing product id",
			req: createCapabilityRequest{
				Name: "Test",
			},
			wantErr: true,
		},
		{
			name: "Invalid name",
			req: createCapabilityRequest{
				Name: "   ",
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
		Id:          id,
	}

	params := req.ToParams()

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

	params := req.ToParams()

	if params.Description.Valid {
		t.Error("expected Description to be invalid for empty string")
	}
}

func TestUpdateCapabilityRequest_Validate(t *testing.T) {
	tests := []struct {
		Req       updateCapabilityRequest
		WantError bool
		ErrText   string
	}{
		{
			Req: updateCapabilityRequest{
				Name: "",
				Id:   1,
			},
			WantError: true,
			ErrText:   "name is required",
		},
		{
			Req: updateCapabilityRequest{
				Name:        "Test Cap",
				Description: "",
				Id:          1,
			},
			WantError: false,
		},
		{
			Req: updateCapabilityRequest{
				Name:        "Test Cap",
				Description: "Test Description",
				Id:          1,
			},
			WantError: false,
		},
		{
			Req: updateCapabilityRequest{
				Name:        "Test Cap",
				Description: "Test Description",
			},
			WantError: true,
			ErrText:   "id is required",
		},
	}
	for _, test := range tests {
		err := test.Req.Validate()
		if err != nil && !test.WantError {
			t.Errorf("expected no error, got %v", err)
		}
		if err == nil && test.WantError {
			t.Errorf("expected error, got nil")
		}
		if err != nil && test.WantError && err.Error() != test.ErrText {
			t.Errorf("expected error %s, got %v", test.ErrText, err)
		}
	}
}

func TestCreateCapabilityStepRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     createCapabilityStepRequest
		wantErr bool
	}{
		{
			name: "Valid request",
			req: createCapabilityStepRequest{
				CapabilityId: 1,
				FlowStepId:   1,
				Target:       "target",
				Protocol:     "protocol",
			},
			wantErr: false,
		},
		{
			name: "Missing capability id",
			req: createCapabilityStepRequest{
				FlowStepId: 1,
				Target:     "target",
				Protocol:   "protocol",
			},
			wantErr: true,
		},
		{
			name: "Missing flow step id",
			req: createCapabilityStepRequest{
				CapabilityId: 1,
				Target:       "target",
				Protocol:     "protocol",
			},
			wantErr: true,
		},
		{
			name: "Missing target",
			req: createCapabilityStepRequest{
				CapabilityId: 1,
				FlowStepId:   1,
				Protocol:     "protocol",
			},
			wantErr: true,
		},
		{
			name: "Missing protocol",
			req: createCapabilityStepRequest{
				CapabilityId: 1,
				FlowStepId:   1,
				Target:       "target",
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

func TestCreateCapabilityStepRequest_ToParams(t *testing.T) {
	req := createCapabilityStepRequest{
		CapabilityId: 1,
		FlowStepId:   2,
		Target:       "test-target",
		Protocol:     "test-protocol",
	}

	params := req.ToParams()

	if params.CapabilityID != req.CapabilityId {
		t.Errorf("expected CapabilityID %d, got %d", req.CapabilityId, params.CapabilityID)
	}
	if params.FlowStepID != req.FlowStepId {
		t.Errorf("expected FlowStepID %d, got %d", req.FlowStepId, params.FlowStepID)
	}
	if !params.Target.Valid || params.Target.String != req.Target {
		t.Errorf("expected Target %s, got %v", req.Target, params.Target)
	}
	if !params.Protocol.Valid || params.Protocol.String != req.Protocol {
		t.Errorf("expected Protocol %s, got %v", req.Protocol, params.Protocol)
	}
	if !params.Timestamp.Valid {
		t.Error("expected Timestamp to be valid")
	}
}
