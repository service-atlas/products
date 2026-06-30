package capability

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"products/internal"
	"products/internal/capability/db"
	"time"
)

type Handler interface {
	CreateCapability(w http.ResponseWriter, r *http.Request)
	GetCapability(w http.ResponseWriter, r *http.Request)
	GetCapabilitiesByFlow(w http.ResponseWriter, r *http.Request)
	GetCapabilitiesByProduct(w http.ResponseWriter, r *http.Request)
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

func (h *capabilityHandler) GetCapability(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Invalid capability ID"}, http.StatusBadRequest)
		return
	}
	ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	capability, err := h.service.GetCapability(ctxWithTimeout, id)
	if err != nil {
		if errors.Is(err, NotFoundError{}) {
			internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Capability not found"}, http.StatusNotFound)
			return
		}
		slog.Error("Failed to fetch capability", slog.Int("capability_id", id), slog.String("error", err.Error()))
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Failed to fetch capability"}, http.StatusInternalServerError)
		return
	}
	internal.WriteJSONResponse(w, r, http.StatusOK, capability)
}

func (h *capabilityHandler) GetCapabilitiesByFlow(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Invalid capability ID"}, http.StatusBadRequest)
		return
	}
	ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	capabilities, err := h.service.GetCapabilitiesByFlow(ctxWithTimeout, id)
	if err != nil {
		if errors.Is(err, NotFoundError{}) {
			internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Flow not found"}, http.StatusNotFound)
			return
		}
		slog.Error("Failed to fetch capabilities", slog.Int("flow_id", id), slog.String("error", err.Error()))
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Failed to fetch capabilities"}, http.StatusInternalServerError)
		return
	}
	internal.WriteJSONResponse(w, r, http.StatusOK, capabilities)
}

func (h *capabilityHandler) GetCapabilitiesByProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Invalid capability ID"}, http.StatusBadRequest)
		return
	}
	ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	capabilities, err := h.service.GetCapabilitiesByProduct(ctxWithTimeout, id)
	if err != nil {
		if errors.Is(err, NotFoundError{}) {
			internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Product not found"}, http.StatusNotFound)
			return
		}
		slog.Error("Failed to fetch capabilities", slog.Int("product_id", id), slog.String("error", err.Error()))
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Failed to fetch capabilities"}, http.StatusInternalServerError)
		return
	}
	internal.WriteJSONResponse(w, r, http.StatusOK, capabilities)
}
