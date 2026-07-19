package capability

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"products/internal"
	"strconv"
	"time"

	"github.com/service-atlas/go-common/errorenvelope"
	"github.com/service-atlas/go-common/httphelpers"
	"github.com/service-atlas/go-common/httplog"
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
	httphelpers.WriteJSONResponse(w, r, http.StatusCreated, newCap)
}

func (h *handler) GetCapability(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid capability ID"}, http.StatusBadRequest)
		return
	}
	ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	capability, err := h.service.GetCapability(ctxWithTimeout, id)
	if err != nil {
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Capability not found"}, http.StatusNotFound)
			return
		}
		slog.Error("Failed to fetch capability", slog.Int("capability_id", id), slog.String("error", err.Error()))
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to fetch capability"}, http.StatusInternalServerError)
		return
	}
	httphelpers.WriteJSONResponse(w, r, http.StatusOK, capability)
}

func (h *handler) GetCapabilitiesByFlow(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid capability ID"}, http.StatusBadRequest)
		return
	}
	ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	capabilities, err := h.service.GetCapabilitiesByFlow(ctxWithTimeout, id)
	if err != nil {
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Flow not found"}, http.StatusNotFound)
			return
		}
		slog.Error("Failed to fetch capabilities", slog.Int("flow_id", id), slog.String("error", err.Error()))
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to fetch capabilities"}, http.StatusInternalServerError)
		return
	}
	httphelpers.WriteJSONResponse(w, r, http.StatusOK, capabilities)
}

func (h *handler) GetCapabilitiesByProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid capability ID"}, http.StatusBadRequest)
		return
	}
	ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	capabilities, err := h.service.GetCapabilitiesByProduct(ctxWithTimeout, id)
	if err != nil {
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Product not found"}, http.StatusNotFound)
			return
		}
		slog.Error("Failed to fetch capabilities", slog.Int("product_id", id), slog.String("error", err.Error()))
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to fetch capabilities"}, http.StatusInternalServerError)
		return
	}
	httphelpers.WriteJSONResponse(w, r, http.StatusOK, capabilities)
}

func (h *handler) UpdateCapability(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid capability ID"}, http.StatusBadRequest)
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
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Capability not found"}, http.StatusNotFound)
			return
		}
		slog.Error("Failed to update capability", slog.Int("capability_id", id), slog.String("error", err.Error()))
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to update capability"}, http.StatusInternalServerError)
		return
	}
	httphelpers.WriteJSONResponse(w, r, http.StatusOK, capability)
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
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Capability not found"}, http.StatusNotFound)
			return
		}
		slog.Error("Failed to delete capability", slog.Int("capability_id", id), slog.String("error", err.Error()))
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to delete capability"}, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) CreateCapabilityStep(w http.ResponseWriter, r *http.Request) {
	req := &createCapabilityStepRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid Body"}, http.StatusBadRequest)
		return
	}
	err := req.Validate()
	if err != nil {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid Body"}, http.StatusBadRequest)
		return
	}
	ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	capStep, err := h.service.CreateCapabilityStep(ctxWithTimeout, *req)
	if err != nil {
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: err.Error()}, http.StatusNotFound)
			return
		} else if internal.IsValidationError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: err.Error()}, http.StatusBadRequest)
			return
		}
		httplog.LoggerFromContext(r.Context()).Error("Failed to create capability step", slog.String("error", err.Error()))
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to create capability step"}, http.StatusInternalServerError)
		return
	}
	httphelpers.WriteJSONResponse(w, r, http.StatusCreated, capStep)
}

func (h *handler) DeleteCapabilityStep(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "id is invalid"}, http.StatusBadRequest)
		return
	}
	ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	err := h.service.DeleteCapabilityStep(ctxWithTimeout, id)
	if err != nil {
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Capability step not found", Instance: strconv.Itoa(id)}, http.StatusNotFound)
			return
		}
		httplog.LoggerFromContext(r.Context()).Error("Error deleting capability step", slog.String("error", err.Error()))
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to delete capability step", Instance: strconv.Itoa(id)}, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) GetCapabilitySteps(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid capability ID"}, http.StatusBadRequest)
		return
	}
	ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	steps, err := h.service.GetCapabilitySteps(ctxWithTimeout, id)
	if err != nil {
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Capability not found"}, http.StatusNotFound)
			return
		}
		slog.Error("Failed to fetch capability steps", slog.Int("capability_id", id), slog.String("error", err.Error()))
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to fetch capability steps"}, http.StatusInternalServerError)
		return
	}
	httphelpers.WriteJSONResponse(w, r, http.StatusOK, steps)
}
