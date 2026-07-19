package product

import (
	"context"
	"encoding/json"
	"net/http"
	"products/internal"
	"products/internal/product/db"
	"strings"
	"time"

	"github.com/service-atlas/go-common/errorenvelope"
	"github.com/service-atlas/go-common/httphelpers"
)

func newHandler(svc productService) *handler {

	return &handler{
		service: svc,
	}
}

type handler struct {
	service productService
}

func (h *handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req createProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid request body"}, http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || req.PlatformID == 0 {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Name and platform ID are required"}, http.StatusBadRequest)
		return
	}

	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	newProduct, err := h.service.CreateProduct(contextWithTimeOut, req)
	if err != nil {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to create product"}, http.StatusInternalServerError)
		return
	}
	httphelpers.WriteJSONResponse(w, r, http.StatusCreated, newProduct)
}

func (h *handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid product ID"}, http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if _, err := h.service.DeleteProduct(contextWithTimeOut, id); err != nil {
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Product not found"}, http.StatusNotFound)
			return
		}
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to delete product"}, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid product ID"}, http.StatusBadRequest)
		return
	}

	var req updateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid request body"}, http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || req.PlatformID == 0 {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Name and platform ID are required"}, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if _, err := h.service.UpdateProduct(ctx, req, id); err != nil {
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Product not found"}, http.StatusNotFound)
			return
		}
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to update product"}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetProductsByPlatform fetches products by platform ID.
func (h *handler) GetProductsByPlatform(w http.ResponseWriter, r *http.Request) {
	platformID, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid platform ID"}, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	products, err := h.service.GetProductsByPlatform(ctx, platformID)
	if err != nil {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to fetch products"}, http.StatusInternalServerError)
		return
	}

	if products == nil {
		products = []db.Product{}
	}

	httphelpers.WriteJSONResponse(w, r, http.StatusOK, products)
}
func (h *handler) GetProductById(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid product ID"}, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	product, err := h.service.GetProductById(ctx, id)
	if err != nil {
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Product not found"}, http.StatusNotFound)
			return
		}
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to fetch product"}, http.StatusInternalServerError)
		return
	}

	httphelpers.WriteJSONResponse(w, r, http.StatusOK, product)
}
