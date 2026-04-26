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

func (h *PlatformHandler) GetPlatforms(w http.ResponseWriter, r *http.Request) {
	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	platforms, err := h.queries.GetPlatforms(contextWithTimeOut)
	if err != nil {
		http.Error(w, "Failed to fetch platforms", http.StatusInternalServerError)
		return
	}
	if platforms == nil {
		platforms = []db.Platform{}
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(platforms)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

}

func (h *PlatformHandler) GetPlatform(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid platform ID", http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	platform, err := h.queries.GetPlatform(contextWithTimeOut, int32(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Platform not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch platform", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(platform)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
