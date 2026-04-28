package platformHandler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"products/internal/db"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

func (h *PlatformHandler) UpdatePlatform(w http.ResponseWriter, r *http.Request) {
	var req db.Platform
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid platform ID", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if req.ID != int32(id) {
		http.Error(w, "Platform ID does not match path", http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	_, err = h.queries.UpdatePlatform(contextWithTimeOut, db.UpdatePlatformParams{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Platform not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update platform", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
