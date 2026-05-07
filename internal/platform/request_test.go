package platform

import (
	"testing"
	"time"
)

func TestCreatePlatformRequest_ToParams(t *testing.T) {
	req := createPlatformRequest{
		Name:        "Test Platform",
		Description: "Test Description",
	}

	params := req.ToParams()

	if params.Name != req.Name {
		t.Errorf("expected name %q, got %q", req.Name, params.Name)
	}
	if !params.Description.Valid {
		t.Error("expected description to be valid")
	}
	if params.Description.String != req.Description {
		t.Errorf("expected description %q, got %q", req.Description, params.Description.String)
	}
	if !params.Timestamp.Valid {
		t.Error("expected timestamp to be valid")
	}
	diff := time.Since(params.Timestamp.Time)
	if diff < 0 {
		diff = -diff
	}
	if diff > 2*time.Second {
		t.Errorf("timestamp too far from now: %v", diff)
	}
}

func TestCreatePlatformRequest_ToParams_EmptyDescription(t *testing.T) {
	req := createPlatformRequest{
		Name:        "Test Platform",
		Description: "",
	}

	params := req.ToParams()

	if params.Description.Valid {
		t.Error("expected description to be invalid")
	}
}

func TestUpdatePlatformRequest_ToParams(t *testing.T) {
	req := updatePlatformRequest{
		ID:          1,
		Name:        "Updated Platform",
		Description: "Updated Description",
	}
	id := int32(1)

	params := req.ToParams(id)

	if params.ID != id {
		t.Errorf("expected id %d, got %d", id, params.ID)
	}
	if params.Name != req.Name {
		t.Errorf("expected name %q, got %q", req.Name, params.Name)
	}
	if !params.Description.Valid {
		t.Error("expected description to be valid")
	}
	if params.Description.String != req.Description {
		t.Errorf("expected description %q, got %q", req.Description, params.Description.String)
	}
	if !params.Updatedat.Valid {
		t.Error("expected updatedat to be valid")
	}
	diff := time.Since(params.Updatedat.Time)
	if diff < 0 {
		diff = -diff
	}
	if diff > 2*time.Second {
		t.Errorf("updatedat too far from now: %v", diff)
	}
}

func TestUpdatePlatformRequest_ToParams_EmptyDescription(t *testing.T) {
	req := updatePlatformRequest{
		ID:          1,
		Name:        "Updated Platform",
		Description: "",
	}

	params := req.ToParams(1)

	if params.Description.Valid {
		t.Error("expected description to be invalid")
	}
}

func TestUpdatePlatformRequest_ToParams_IDPrecedence(t *testing.T) {
	req := updatePlatformRequest{
		ID:          100, // ID from body
		Name:        "Precedence Test",
		Description: "Testing ID precedence",
	}
	pathID := int32(200) // ID from path/argument

	params := req.ToParams(pathID)

	// Verify that the method argument ID is preferred over the struct field ID
	if params.ID != pathID {
		t.Errorf("expected id %d (path ID), got %d", pathID, params.ID)
	}

	// Verify that Name and Description are preserved
	if params.Name != req.Name {
		t.Errorf("expected name %q, got %q", req.Name, params.Name)
	}
	if params.Description.String != req.Description {
		t.Errorf("expected description %q, got %q", req.Description, params.Description.String)
	}
}
