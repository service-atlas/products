package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// Mock Platform Handler
type mockPlatformHandler struct {
	called map[string]bool
}

func newMockPlatformHandler() *mockPlatformHandler {
	return &mockPlatformHandler{called: make(map[string]bool)}
}

func (m *mockPlatformHandler) CreatePlatform(w http.ResponseWriter, r *http.Request) {
	m.called["CreatePlatform"] = true
}
func (m *mockPlatformHandler) GetPlatforms(w http.ResponseWriter, r *http.Request) {
	m.called["GetPlatforms"] = true
}
func (m *mockPlatformHandler) GetPlatform(w http.ResponseWriter, r *http.Request) {
	m.called["GetPlatform"] = true
}
func (m *mockPlatformHandler) UpdatePlatform(w http.ResponseWriter, r *http.Request) {
	m.called["UpdatePlatform"] = true
}
func (m *mockPlatformHandler) DeletePlatform(w http.ResponseWriter, r *http.Request) {
	m.called["DeletePlatform"] = true
}

// Mock Product Handler
type mockProductHandler struct {
	called map[string]bool
}

func newMockProductHandler() *mockProductHandler {
	return &mockProductHandler{called: make(map[string]bool)}
}

func (m *mockProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	m.called["CreateProduct"] = true
}
func (m *mockProductHandler) GetProductsByPlatform(w http.ResponseWriter, r *http.Request) {
	m.called["GetProductsByPlatform"] = true
}
func (m *mockProductHandler) GetProductById(w http.ResponseWriter, r *http.Request) {
	m.called["GetProductById"] = true
}
func (m *mockProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	m.called["UpdateProduct"] = true
}
func (m *mockProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	m.called["DeleteProduct"] = true
}

// Mock Flow Handler
type mockFlowHandler struct {
	called map[string]bool
}

func newMockFlowHandler() *mockFlowHandler {
	return &mockFlowHandler{called: make(map[string]bool)}
}

func (m *mockFlowHandler) CreateFlow(w http.ResponseWriter, r *http.Request) {
	m.called["CreateFlow"] = true
}
func (m *mockFlowHandler) GetFlowById(w http.ResponseWriter, r *http.Request) {
	m.called["GetFlowById"] = true
}
func (m *mockFlowHandler) GetFlowsByProduct(w http.ResponseWriter, r *http.Request) {
	m.called["GetFlowsByProduct"] = true
}
func (m *mockFlowHandler) UpdateFlow(w http.ResponseWriter, r *http.Request) {
	m.called["UpdateFlow"] = true
}
func (m *mockFlowHandler) DeleteFlow(w http.ResponseWriter, r *http.Request) {
	m.called["DeleteFlow"] = true
}
func (m *mockFlowHandler) CreateFlowStep(w http.ResponseWriter, r *http.Request) {
	m.called["CreateFlowStep"] = true
}
func (m *mockFlowHandler) DeleteFlowStep(w http.ResponseWriter, r *http.Request) {
	m.called["DeleteFlowStep"] = true
}
func (m *mockFlowHandler) UpdateFlowStep(w http.ResponseWriter, r *http.Request) {
	m.called["UpdateFlowStep"] = true
}
func (m *mockFlowHandler) GetFlowSteps(w http.ResponseWriter, r *http.Request) {
	m.called["GetFlowSteps"] = true
}
func (m *mockFlowHandler) GetFlowPath(w http.ResponseWriter, r *http.Request) {
	m.called["GetFlowPath"] = true
}

func TestProductRoutes_SetupRoutes(t *testing.T) {
	mockPlatform := newMockPlatformHandler()
	mockProduct := newMockProductHandler()
	mockFlow := newMockFlowHandler()

	pr := &productRoutes{
		platformHandler: mockPlatform,
		productHandler:  mockProduct,
		flowHandler:     mockFlow,
	}

	r := chi.NewRouter()
	pr.setupRoutes(r)

	tests := []struct {
		method  string
		path    string
		handler string
		mock    map[string]bool
	}{
		{"POST", "/platforms/", "CreatePlatform", mockPlatform.called},
		{"GET", "/platforms/", "GetPlatforms", mockPlatform.called},
		{"GET", "/platforms/1", "GetPlatform", mockPlatform.called},
		{"DELETE", "/platforms/1", "DeletePlatform", mockPlatform.called},
		{"PUT", "/platforms/1", "UpdatePlatform", mockPlatform.called},
		{"GET", "/platforms/1/products", "GetProductsByPlatform", mockProduct.called},

		{"POST", "/products/", "CreateProduct", mockProduct.called},
		{"GET", "/products/1", "GetProductById", mockProduct.called},
		{"DELETE", "/products/1", "DeleteProduct", mockProduct.called},
		{"PUT", "/products/1", "UpdateProduct", mockProduct.called},
		{"POST", "/products/1/flows", "CreateFlow", mockFlow.called},
		{"GET", "/products/1/flows", "GetFlowsByProduct", mockFlow.called},

		{"POST", "/flows/1/steps", "CreateFlowStep", mockFlow.called},
		{"GET", "/flows/1/steps", "GetFlowSteps", mockFlow.called},
		{"GET", "/flows/1/path", "GetFlowPath", mockFlow.called},
		{"GET", "/flows/1", "GetFlowById", mockFlow.called},
		{"PUT", "/flows/1", "UpdateFlow", mockFlow.called},
		{"DELETE", "/flows/1", "DeleteFlow", mockFlow.called},

		{"DELETE", "/flow-steps/1", "DeleteFlowStep", mockFlow.called},
		{"PATCH", "/flow-steps/1", "UpdateFlowStep", mockFlow.called},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if !tt.mock[tt.handler] {
				t.Errorf("expected %s to be called for %s %s", tt.handler, tt.method, tt.path)
			}
		})
	}
}
