package platformHandler

import (
	"encoding/json"
	"net/http"
	"products/internal/db"

	"github.com/jackc/pgx/v5/pgtype"
)

type PlatformHandler struct {
	queries PlatformQuerier
}

func NewPlatformHandler(q PlatformQuerier) *PlatformHandler {
	return &PlatformHandler{
		queries: q,
	}
}

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

	if err := h.queries.CreatePlatform(r.Context(), db.CreatePlatformParams{
		Name: req.Name,
		Description: pgtype.Text{
			Valid:  true,
			String: req.Description,
		},
	}); err != nil {
		http.Error(w, "Failed to create platform", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
