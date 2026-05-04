package productHandler

import (
	"context"
	"errors"
	"net/http"
	"products/internal"
	"time"

	"github.com/jackc/pgx/v5"
)

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	err := h.queries.DeleteProduct(contextWithTimeOut, id)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "Failed to delete product", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
