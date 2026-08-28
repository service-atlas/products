package flow

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"products/internal"
	"strings"
	"time"

	"github.com/service-atlas/go-common/errorenvelope"
	"github.com/service-atlas/go-common/httphelpers"
	"github.com/service-atlas/go-common/httplog"
)

func newHandler(service flowService) *handler {
	return &handler{
		flowService: service,
	}
}

type handler struct {
	flowService flowService
}

func (h *handler) CreateFlow(w http.ResponseWriter, r *http.Request) {
	id, ok := httphelpers.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid product ID"}, http.StatusBadRequest)
		return
	}

	req := &createFlowRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid request body"}, http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Flow name cannot be empty"}, http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeoutCause(r.Context(), 5*time.Second, errors.New("flow creation timed out"))
	defer cancel()
	flow, err := h.flowService.CreateFlow(contextWithTimeOut, *req, id)
	if err != nil {
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: err.Error()}, http.StatusNotFound)
			return
		}
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to create flow"}, http.StatusInternalServerError)
		return
	}
	httphelpers.WriteJSONResponse(w, r, http.StatusCreated, flow)
}

func (h *handler) GetFlowById(w http.ResponseWriter, r *http.Request) {
	id, ok := httphelpers.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid flow ID", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeoutCause(r.Context(), 5*time.Second, errors.New("fetching flow timed out"))
	defer cancel()
	flow, err := h.flowService.GetFlowById(contextWithTimeOut, id)
	if err != nil {
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: err.Error(), Instance: r.URL.Path}, http.StatusNotFound)
			return
		}
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to fetch flow", Instance: r.URL.Path}, http.StatusInternalServerError)
		return
	}
	httphelpers.WriteJSONResponse(w, r, http.StatusOK, flow)
}

func (h *handler) GetFlowsByProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := httphelpers.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid product ID", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeoutCause(r.Context(), 5*time.Second, errors.New("fetching flows timed out"))
	defer cancel()
	flows, err := h.flowService.GetFlowsByProduct(contextWithTimeOut, id)
	if err != nil {
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: err.Error(), Instance: r.URL.Path}, http.StatusNotFound)
			return
		}
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to fetch flows"}, http.StatusInternalServerError)
		return
	}
	httphelpers.WriteJSONResponse(w, r, http.StatusOK, flows)
}

func (h *handler) UpdateFlow(w http.ResponseWriter, r *http.Request) {
	id, ok := httphelpers.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid flow ID", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}

	req := &updateFlowRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid request body", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}

	if req.Name == "" && req.Description == "" {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "No updatable fields provided", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}

	contextWithTimeOut, cancel := context.WithTimeoutCause(r.Context(), 5*time.Second, errors.New("updating flow timed out"))
	defer cancel()

	flow, err := h.flowService.UpdateFlow(contextWithTimeOut, *req, id)
	if err != nil {
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: err.Error(), Instance: r.URL.Path}, http.StatusNotFound)
			return
		}
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to update flow", Instance: r.URL.Path}, http.StatusInternalServerError)
		return
	}

	httphelpers.WriteJSONResponse(w, r, http.StatusOK, flow)
}

func (h *handler) DeleteFlow(w http.ResponseWriter, r *http.Request) {
	id, ok := httphelpers.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid flow ID", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}

	contextWithTimeOut, cancel := context.WithTimeoutCause(r.Context(), 5*time.Second, errors.New("deleting flow timed out"))
	defer cancel()

	err := h.flowService.DeleteFlow(contextWithTimeOut, id)
	if err != nil {
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: err.Error(), Instance: r.URL.Path}, http.StatusNotFound)
			return
		}
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to delete flow", Instance: r.URL.Path}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) CreateFlowStep(w http.ResponseWriter, r *http.Request) {
	id, ok := httphelpers.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid flow ID", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}
	req := &createFlowStepRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid request body", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}
	req.FlowId = id

	if req.Current == "" || req.Next == "" {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Current and next UUIDs are required", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}

	if _, err := toPgUUID(req.Current); err != nil {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid current UUID", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}
	if _, err := toPgUUID(req.Next); err != nil {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid next UUID", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}

	contextWithTimeOut, cancel := context.WithTimeoutCause(r.Context(), 5*time.Second, errors.New("creating flow step timed out"))
	defer cancel()
	flowStep, err := h.flowService.CreateFlowStep(contextWithTimeOut, *req)
	if err != nil {
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: err.Error(), Instance: r.URL.Path}, http.StatusNotFound)
			return
		}

		if _, ok := errors.AsType[DependencyValidationError](err); ok {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: err.Error(), Instance: r.URL.Path}, http.StatusUnprocessableEntity)
			return
		}

		if _, ok := errors.AsType[ConflictError](err); ok {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: err.Error(), Instance: r.URL.Path}, http.StatusConflict)
			return
		}

		logger := httplog.LoggerFromContext(r.Context())
		logger.Error("Failed to create flow step",
			slog.String("error", err.Error()),
			slog.Int("flow_id", id))
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to create flow step", Instance: r.URL.Path}, http.StatusInternalServerError)
		return
	}
	httphelpers.WriteJSONResponse(w, r, http.StatusCreated, flowStep)
}

func (h *handler) DeleteFlowStep(w http.ResponseWriter, r *http.Request) {
	id, ok := httphelpers.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid flow step ID", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeoutCause(r.Context(), 5*time.Second, errors.New("deleting flow step timed out"))
	defer cancel()
	err := h.flowService.DeleteFlowStep(contextWithTimeOut, id)
	if err != nil {
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: err.Error(), Instance: r.URL.Path}, http.StatusNotFound)
			return
		}
		logger := httplog.LoggerFromContext(r.Context())
		logger.Error("Failed to delete flow step",
			slog.String("error", err.Error()),
			slog.Int("flow_step_id", id))
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to delete flow step", Instance: r.URL.Path}, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) GetFlowSteps(w http.ResponseWriter, r *http.Request) {
	id, ok := httphelpers.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid flow ID", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeoutCause(r.Context(), 5*time.Second, errors.New("fetching flow steps timed out"))
	defer cancel()
	flowSteps, err := h.flowService.GetFlowSteps(contextWithTimeOut, id)
	if err != nil {
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: err.Error(), Instance: r.URL.Path}, http.StatusNotFound)
			return
		}
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to fetch flow steps", Instance: r.URL.Path}, http.StatusInternalServerError)
		return
	}
	httphelpers.WriteJSONResponse(w, r, http.StatusOK, flowSteps)
}

func (h *handler) GetFlowPath(w http.ResponseWriter, r *http.Request) {
	id, ok := httphelpers.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid flow ID", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeoutCause(r.Context(), 5*time.Second, errors.New("fetching flow path timed out"))
	defer cancel()
	flowPath, err := h.flowService.GetFlowPath(contextWithTimeOut, id)
	if err != nil {
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: err.Error(), Instance: r.URL.Path}, http.StatusNotFound)
			return
		}
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to fetch flow path", Instance: r.URL.Path}, http.StatusInternalServerError)
		return
	}
	httphelpers.WriteJSONResponse(w, r, http.StatusOK, flowPath)
}
