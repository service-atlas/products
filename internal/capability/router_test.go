package capability

import (
	"context"
	"net/http"
	"products/internal/capability/db"
	"testing"

	"github.com/go-chi/chi/v5"
)

type mockService struct {
	called map[string]bool
}

func newMockService() *mockService {
	return &mockService{called: make(map[string]bool)}
}

func (m *mockService) CreateCapability(_ context.Context, _ createCapabilityRequest) (db.Capability, error) {
	m.called["CreateCapability"] = true
	return db.Capability{}, nil
}
func (m *mockService) GetCapability(_ context.Context, _ int) (db.Capability, error) {
	m.called["GetCapability"] = true
	return db.Capability{}, nil
}
func (m *mockService) GetCapabilitiesByProduct(_ context.Context, _ int) ([]db.Capability, error) {
	m.called["GetCapabilitiesByProduct"] = true
	return nil, nil
}
func (m *mockService) GetCapabilitiesByFlow(_ context.Context, _ int) ([]db.GetCapabilitiesByFlowRow, error) {
	m.called["GetCapabilitiesByFlow"] = true
	return nil, nil
}
func (m *mockService) UpdateCapability(_ context.Context, _ updateCapabilityRequest) (db.Capability, error) {
	m.called["UpdateCapability"] = true
	return db.Capability{}, nil
}

func (m *mockService) DeleteCapability(_ context.Context, _ int) error {
	m.called["DeleteCapability"] = true
	return nil
}

func (m *mockService) CreateCapabilityStep(_ context.Context, _ createCapabilityStepRequest) (db.CapabilityStep, error) {
	m.called["CreateCapabilityStep"] = true
	return db.CapabilityStep{}, nil
}

func (m *mockService) DeleteCapabilityStep(_ context.Context, _ int) error {
	m.called["DeleteCapabilityStep"] = true
	return nil
}

func (m *mockService) GetCapabilitySteps(_ context.Context, _ int) ([]db.CapabilityStep, error) {
	m.called["GetCapabilitySteps"] = true
	return []db.CapabilityStep{}, nil
}

func TestRegisterRoutesWithHandler(t *testing.T) {
	h := newHandler(newMockService())

	r := chi.NewRouter()
	registerRoutesWithHandler(r, h)

	got := map[string]bool{}

	err := chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		got[method+" "+route] = true

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"GET /products/{id}/capabilities",
		"GET /flows/{id}/capabilities",
		"POST /capabilities/",
		"GET /capabilities/{id}",
		"PUT /capabilities/{id}",
		"DELETE /capabilities/{id}",
		"POST /capability-steps/",
		"DELETE /capability-steps/{id}",
		"GET /capabilities/{id}/steps",
	}

	for _, route := range want {
		if !got[route] {
			t.Errorf("expected route %q to be registered", route)
		}
	}
}
