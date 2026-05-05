package productHandler

import (
	"context"
	"encoding/json"
	"net/http"
	"products/internal"
	"products/internal/db/product"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type UpdateProductRequest struct {
	PlatformID  int32  `json:"platform_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var req UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.PlatformID == 0 {
		http.Error(w, "Name and platform ID are required", http.StatusBadRequest)
		return
	}

	params := product.UpdateProductParams{
		ID:         id,
		PlatformID: req.PlatformID,
		Name:       req.Name,
		Description: pgtype.Text{
			Valid:  req.Description != "",
			String: req.Description,
		},
		UpdatedAt: pgtype.Timestamptz{
			Valid: true,
			Time:  time.Now().UTC(),
		},
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.queries.UpdateProduct(ctx, params); err != nil {
		http.Error(w, "Failed to update product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
