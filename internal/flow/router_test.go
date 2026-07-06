package flow

import (
	"context"
	"net/http"
	"products/internal/flow/db"
	"testing"

	"github.com/go-chi/chi/v5"
)

type mockService struct {
	called map[string]bool
}

func newMockService() *mockService {
	return &mockService{called: make(map[string]bool)}
}

func (m *mockService) CreateFlow(ctx context.Context, req createFlowRequest, id int) (db.Flow, error) {
	m.called["CreateFlow"] = true
	return db.Flow{}, nil
}
func (m *mockService) GetFlowById(ctx context.Context, id int) (db.Flow, error) {
	m.called["GetFlowById"] = true
	return db.Flow{}, nil
}
func (m *mockService) GetFlowsByProduct(ctx context.Context, id int) ([]db.Flow, error) {
	m.called["GetFlowsByProduct"] = true
	return nil, nil
}
func (m *mockService) UpdateFlow(ctx context.Context, req updateFlowRequest, id int) (db.Flow, error) {
	m.called["UpdateFlow"] = true
	return db.Flow{}, nil
}
func (m *mockService) DeleteFlow(ctx context.Context, id int) error {
	m.called["DeleteFlow"] = true
	return nil
}
func (m *mockService) CreateFlowStep(ctx context.Context, req createFlowStepRequest) (db.FlowStep, error) {
	m.called["CreateFlowStep"] = true
	return db.FlowStep{}, nil
}
func (m *mockService) DeleteFlowStep(ctx context.Context, id int) error {
	m.called["DeleteFlowStep"] = true
	return nil
}
func (m *mockService) GetFlowSteps(ctx context.Context, id int) ([]db.FlowStep, error) {
	m.called["GetFlowSteps"] = true
	return nil, nil
}
func (m *mockService) GetFlowPath(ctx context.Context, id int) (FlowPath, error) {
	m.called["GetFlowPath"] = true
	return FlowPath{}, nil
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
		"POST /flows/{id}/steps",
		"GET /flows/{id}/steps",
		"GET /flows/{id}/path",
		"GET /flows/{id}/",
		"PUT /flows/{id}/",
		"DELETE /flows/{id}/",
		"DELETE /flow-steps/{id}",
		"POST /products/{id}/flows",
		"GET /products/{id}/flows",
	}

	for _, route := range want {
		if !got[route] {
			t.Errorf("expected route %q to be registered", route)
		}
	}
}
