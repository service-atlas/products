package product

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterRoutes(t *testing.T) {
	h := newHandler(&mockProductService{})

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
		"GET /platforms/{id}/products",
		"POST /products/",
		"GET /products/{id}/",
		"PUT /products/{id}/",
		"DELETE /products/{id}/",
	}

	for _, route := range want {
		if !got[route] {
			t.Errorf("expected route %q to be registered", route)
		}
	}
}
