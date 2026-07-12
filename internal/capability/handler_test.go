package capability

import (
	"context"
	"products/internal/capability/db"
	"testing"
)

type mockCapabilityService struct {
	createCapabilityFunc         func(ctx context.Context, req createCapabilityRequest) (db.Capability, error)
	getCapabilityFunc            func(ctx context.Context, id int) (db.Capability, error)
	getCapabilitiesByProductFunc func(ctx context.Context, id int) ([]db.Capability, error)
	getCapabilitiesByFlowFunc    func(ctx context.Context, id int) ([]db.GetCapabilitiesByFlowRow, error)
	updateCapabilityFunc         func(ctx context.Context, req updateCapabilityRequest) (db.Capability, error)
	deleteCapabilityFunc         func(ctx context.Context, id int) error
	createCapabilityStepFunc     func(ctx context.Context, req createCapabilityStepRequest) (db.CapabilityStep, error)
}

func (m *mockCapabilityService) CreateCapability(ctx context.Context, req createCapabilityRequest) (db.Capability, error) {
	return m.createCapabilityFunc(ctx, req)
}

func (m *mockCapabilityService) GetCapability(ctx context.Context, id int) (db.Capability, error) {
	return m.getCapabilityFunc(ctx, id)
}

func (m *mockCapabilityService) GetCapabilitiesByProduct(ctx context.Context, id int) ([]db.Capability, error) {
	return m.getCapabilitiesByProductFunc(ctx, id)
}

func (m *mockCapabilityService) GetCapabilitiesByFlow(ctx context.Context, id int) ([]db.GetCapabilitiesByFlowRow, error) {
	return m.getCapabilitiesByFlowFunc(ctx, id)
}

func (m *mockCapabilityService) UpdateCapability(ctx context.Context, req updateCapabilityRequest) (db.Capability, error) {
	return m.updateCapabilityFunc(ctx, req)
}

func (m *mockCapabilityService) DeleteCapability(ctx context.Context, id int) error {
	return m.deleteCapabilityFunc(ctx, id)
}

func (m *mockCapabilityService) CreateCapabilityStep(ctx context.Context, req createCapabilityStepRequest) (db.CapabilityStep, error) {
	return m.createCapabilityStepFunc(ctx, req)
}

func TestNewHandler(t *testing.T) {
	mockService := &mockCapabilityService{}
	h := newHandler(mockService)
	if h == nil {
		t.Error("expected handler to be non-nil")
	}
}
