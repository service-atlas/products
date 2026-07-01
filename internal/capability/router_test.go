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

func (m *mockService) CreateCapability(ctx context.Context, req createCapabilityRequest) (db.Capability, error) {
	m.called["CreateCapability"] = true
	return db.Capability{}, nil
}
func (m *mockService) GetCapability(ctx context.Context, id int) (db.Capability, error) {
	m.called["GetCapability"] = true
	return db.Capability{}, nil
}
func (m *mockService) GetCapabilitiesByProduct(ctx context.Context, id int) ([]db.GetCapabilitiesByProductRow, error) {
	m.called["GetCapabilitiesByProduct"] = true
	return nil, nil
}
func (m *mockService) GetCapabilitiesByFlow(ctx context.Context, id int) ([]db.Capability, error) {
	m.called["GetCapabilitiesByFlow"] = true
	return nil, nil
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
	}

	for _, route := range want {
		if !got[route] {
			t.Errorf("expected route %q to be registered", route)
		}
	}
}
