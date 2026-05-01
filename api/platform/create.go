package platformHandler

import (
	"context"
	"encoding/json"
	"net/http"
	db "products/internal/db/platform"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type CreatePlatformRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *PlatformHandler) CreatePlatform(w http.ResponseWriter, r *http.Request) {
	var req CreatePlatformRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if err := h.queries.CreatePlatform(contextWithTimeOut, db.CreatePlatformParams{
		Name: req.Name,
		Description: pgtype.Text{
			Valid:  req.Description != "",
			String: req.Description,
		},
		Timestamp: pgtype.Timestamptz{
			Valid: true,
			Time:  time.Now().UTC(),
		},
	}); err != nil {
		http.Error(w, "Failed to create platform", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
