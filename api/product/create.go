package productHandler

import (
	"context"
	"encoding/json"
	"net/http"
	"products/internal/db/product"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req product.CreateProductParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.Timestamp = pgtype.Timestamptz{
		Valid: true,
		Time:  time.Now().UTC(),
	}

	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if err := h.queries.CreateProduct(contextWithTimeOut, req); err != nil {
		http.Error(w, "Failed to create product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}
