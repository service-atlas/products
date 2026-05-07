package product

import (
	"testing"
	"time"
)

func TestCreateProductRequest_ToParams(t *testing.T) {
	req := createProductRequest{
		Name:        "Test Product",
		PlatformID:  1,
		Description: "Test Description",
	}

	params := req.ToParams()

	if params.Name != req.Name {
		t.Errorf("expected name %q, got %q", req.Name, params.Name)
	}
	if params.PlatformID != req.PlatformID {
		t.Errorf("expected platform_id %d, got %d", req.PlatformID, params.PlatformID)
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

func TestCreateProductRequest_ToParams_EmptyDescription(t *testing.T) {
	req := createProductRequest{
		Name:        "Test Product",
		PlatformID:  1,
		Description: "",
	}

	params := req.ToParams()

	if params.Description.Valid {
		t.Error("expected description to be invalid")
	}
}

func TestUpdateProductRequest_ToParams(t *testing.T) {
	req := updateProductRequest{
		PlatformID:  2,
		Name:        "Updated Product",
		Description: "Updated Description",
	}
	id := int32(10)

	params := req.ToParams(id)

	if params.ID != id {
		t.Errorf("expected id %d, got %d", id, params.ID)
	}
	if params.PlatformID != req.PlatformID {
		t.Errorf("expected platform_id %d, got %d", req.PlatformID, params.PlatformID)
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
	if !params.UpdatedAt.Valid {
		t.Error("expected updated_at to be valid")
	}
	diff := time.Since(params.UpdatedAt.Time)
	if diff < 0 {
		diff = -diff
	}
	if diff > 2*time.Second {
		t.Errorf("updated_at too far from now: %v", diff)
	}
}

func TestUpdateProductRequest_ToParams_EmptyDescription(t *testing.T) {
	req := updateProductRequest{
		PlatformID:  2,
		Name:        "Updated Product",
		Description: "",
	}

	params := req.ToParams(10)

	if params.Description.Valid {
		t.Error("expected description to be invalid")
	}
}
