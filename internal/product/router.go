package product

import (
	"context"
	"products/internal/product/db"

	"github.com/go-chi/chi/v5"
)

type productService interface {
	CreateProduct(ctx context.Context, req createProductRequest) (db.Product, error)
	GetProductsByPlatform(ctx context.Context, platformID int) ([]db.Product, error)
	GetProductById(ctx context.Context, id int) (db.Product, error)
	UpdateProduct(ctx context.Context, req updateProductRequest, id int) (int, error)
	DeleteProduct(ctx context.Context, id int) (int, error)
}

func RegisterRoutes(r chi.Router, dbConn db.DBTX) {
	s := newPostgresService(dbConn)
	h := newHandler(s)
	registerRoutesWithHandler(r, h)
}

func registerRoutesWithHandler(r chi.Router, prodHandler *handler) {
	r.Get("/platforms/{id}/products", prodHandler.GetProductsByPlatform)
	r.Route("/products", func(u chi.Router) {
		u.Post("/", prodHandler.CreateProduct)
		u.Route("/{id}", func(u chi.Router) {
			u.Get("/", prodHandler.GetProductById)
			u.Put("/", prodHandler.UpdateProduct)
			u.Delete("/", prodHandler.DeleteProduct)
		})
	})
}
