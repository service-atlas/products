package platformHandler

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

func (h *PlatformHandler) DeletePlatform(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid platform ID", http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	err = h.queries.DeletePlatform(contextWithTimeOut, int32(id))
	if err != nil {
		http.Error(w, "Failed to fetch platform", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
