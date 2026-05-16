package flow

import (
	"testing"
)

func TestCreateFlowRequest_ToParams(t *testing.T) {
	req := createFlowRequest{
		Name:        "Test Flow",
		Description: "Test Description",
	}
	productID := 123
	params := req.ToParams(productID)

	if params.Name != req.Name {
		t.Errorf("expected Name %v, got %v", req.Name, params.Name)
	}
	if params.ProductID != productID {
		t.Errorf("expected ProductID %v, got %v", productID, params.ProductID)
	}
	if !params.Description.Valid || params.Description.String != req.Description {
		t.Errorf("expected Description %v, got %v", req.Description, params.Description.String)
	}
	if !params.TimeStamp.Valid {
		t.Error("expected TimeStamp to be valid")
	}
}
