package capability

import (
	"context"
	"encoding/json"
	"net/http"
	"products/internal"
	"products/internal/capability/db"
	"time"
)

type Handler interface {
	CreateCapability(w http.ResponseWriter, r *http.Request)
}

func NewHandler(dbConn db.DBTX) Handler {
	queries := db.New(dbConn)
	service := &postgresService{
		queries: queries,
	}
	return &capabilityHandler{
		service: service,
	}
}

type capabilityHandler struct {
	service capabilityService
}

func (h *capabilityHandler) CreateCapability(w http.ResponseWriter, r *http.Request) {
	capReq := &createCapabilityRequest{}
	if err := json.NewDecoder(r.Body).Decode(capReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := capReq.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	newCap, err := h.service.CreateCapability(ctxWithTimeout, *capReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	internal.WriteJSONResponse(w, r, http.StatusCreated, newCap)
}
