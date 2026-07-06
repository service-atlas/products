package platform

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterRoutes(t *testing.T) {
	h := newHandler(&mockPlatformService{})

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
		"POST /platforms/",
		"GET /platforms/",
		"GET /platforms/{id}/",
		"PUT /platforms/{id}/",
		"DELETE /platforms/{id}/",
	}

	for _, route := range want {
		if !got[route] {
			t.Errorf("expected route %q to be registered", route)
		}
	}
}
