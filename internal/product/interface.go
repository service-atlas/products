package product

import "net/http"

type Handler interface {
	CreateProduct(w http.ResponseWriter, r *http.Request)
	GetProductsByPlatform(w http.ResponseWriter, r *http.Request)
	GetProductById(w http.ResponseWriter, r *http.Request)
	UpdateProduct(w http.ResponseWriter, r *http.Request)
	DeleteProduct(w http.ResponseWriter, r *http.Request)
}
