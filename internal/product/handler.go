package product

import (
	"net/http"
)

type productHandler struct {
	queries *Queries
}

func NewProductHandler(db DBTX) Handler {
	queries := &Queries{
		db: db,
	}
	return &productHandler{
		queries: queries,
	}
}

type Handler interface {
	CreateProduct(w http.ResponseWriter, r *http.Request)
	GetProductsByPlatform(w http.ResponseWriter, r *http.Request)
	GetProductById(w http.ResponseWriter, r *http.Request)
	UpdateProduct(w http.ResponseWriter, r *http.Request)
	DeleteProduct(w http.ResponseWriter, r *http.Request)
}
