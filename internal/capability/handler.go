package capability

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"products/internal"
	"time"
)

func newHandler(svc capabilityService) *handler {
	return &handler{
		service: svc,
	}
}

type handler struct {
	service capabilityService
}

func (h *handler) CreateCapability(w http.ResponseWriter, r *http.Request) {
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

func (h *handler) GetCapability(w http.ResponseWriter, r *http.Request) {
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

func (h *handler) GetCapabilitiesByFlow(w http.ResponseWriter, r *http.Request) {
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

func (h *handler) GetCapabilitiesByProduct(w http.ResponseWriter, r *http.Request) {
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

func (h *handler) UpdateCapability(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Invalid capability ID"}, http.StatusBadRequest)
		return
	}
	req := &updateCapabilityRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.Id = id
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	capability, err := h.service.UpdateCapability(ctxWithTimeout, *req)
	if err != nil {
		if errors.Is(err, NotFoundError{}) {
			internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Capability not found"}, http.StatusNotFound)
			return
		}
		slog.Error("Failed to update capability", slog.Int("capability_id", id), slog.String("error", err.Error()))
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Failed to update capability"}, http.StatusInternalServerError)
		return
	}
	internal.WriteJSONResponse(w, r, http.StatusOK, capability)
}

func (h *handler) DeleteCapability(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		http.Error(w, "Invalid capability ID", http.StatusBadRequest)
		return
	}
	ctxWithTimout, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	err := h.service.DeleteCapability(ctxWithTimout, id)
	if err != nil {
		if errors.Is(err, NotFoundError{}) {
			internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Capability not found"}, http.StatusNotFound)
			return
		}
		slog.Error("Failed to delete capability", slog.Int("capability_id", id), slog.String("error", err.Error()))
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Failed to delete capability"}, http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}
