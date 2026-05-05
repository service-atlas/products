package productHandler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"products/internal"
	db "products/internal/db/product"
	"time"

	"github.com/jackc/pgx/v5"
)

// GetProductsByPlatform fetches products by platform ID.
func (h *ProductHandler) GetProductsByPlatform(w http.ResponseWriter, r *http.Request) {
	platformID, ok := internal.GetIntFromRequestPath("platform_id", r)
	if !ok {
		http.Error(w, "Invalid platform ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	products, err := h.queries.GetProductsByPlatform(ctx, platformID)
	if err != nil {
		http.Error(w, "Failed to fetch products", http.StatusInternalServerError)
		return
	}

	if products == nil {
		products = []db.Product{}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(products); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetProductById fetches a single product by ID.
func (h *ProductHandler) GetProductById(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	product, err := h.queries.GetProductById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(product); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
