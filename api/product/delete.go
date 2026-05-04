package productHandler

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"products/internal"
	"time"
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
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Failed to delete product", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
