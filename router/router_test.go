package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type mockDBTX struct{}

func (m *mockDBTX) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (m *mockDBTX) Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error) {
	return nil, nil
}
func (m *mockDBTX) QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row {
	return nil
}

func TestInitializeRouter(t *testing.T) {
	r := InitializeRouter(&mockDBTX{})

	got := map[string]bool{}

	err := chi.Walk(r.(*chi.Mux), func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		got[method+" "+route] = true
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"GET /api/time",
		"GET /api/version",
		"POST /platforms/",
		"GET /platforms/",
		"GET /platforms/{id}/",
		"DELETE /platforms/{id}/",
		"PUT /platforms/{id}/",
		"POST /products/",
		"GET /platforms/{id}/products",
		"GET /products/{id}/",
		"PUT /products/{id}/",
		"DELETE /products/{id}/",
		"POST /flows/{id}/steps",
		"GET /flows/{id}/steps",
		"GET /flows/{id}/path",
		"GET /flows/{id}/",
		"PUT /flows/{id}/",
		"DELETE /flows/{id}/",
		"DELETE /flow-steps/{id}",
		"POST /products/{id}/flows",
		"GET /products/{id}/flows",
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

func TestInitializeRouter_NestedRoutesReachHandlers(t *testing.T) {
	r := InitializeRouter(&mockDBTX{})

	cases := []struct{ method, path string }{
		{http.MethodGet, "/platforms/1/products"},
		{http.MethodGet, "/products/1/flows"},
		{http.MethodGet, "/flows/1/capabilities"},
		{http.MethodGet, "/products/1/capabilities"},
	}
	for _, c := range cases {
		req := httptest.NewRequest(c.method, c.path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code == http.StatusNotFound {
			t.Errorf("%s %s: expected route to be reachable, got 404", c.method, c.path)
		}
	}
}
