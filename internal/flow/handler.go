package flow

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"products/internal"
	"products/internal/flow/db"
	"strings"
	"time"
)

type Handler interface {
	CreateFlow(w http.ResponseWriter, r *http.Request)
	GetFlowById(w http.ResponseWriter, r *http.Request)
	GetFlowsByProduct(w http.ResponseWriter, r *http.Request)
	UpdateFlow(w http.ResponseWriter, r *http.Request)
	DeleteFlow(w http.ResponseWriter, r *http.Request)
	CreateFlowStep(w http.ResponseWriter, r *http.Request)
}

func NewHandler(dbConn db.DBTX) Handler {
	queries := db.New(dbConn)
	client := &http.Client{Timeout: 5 * time.Second}
	service := &postgresService{
		queries: queries,
		client:  client,
	}
	return &flowHandler{
		flowService: service,
	}
}

type flowHandler struct {
	flowService flowService
}

func (h *flowHandler) CreateFlow(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Invalid product ID"}, http.StatusBadRequest)
		return
	}

	req := &createFlowRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Invalid request body"}, http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Flow name cannot be empty"}, http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeoutCause(r.Context(), 5*time.Second, errors.New("flow creation timed out"))
	defer cancel()
	flow, err := h.flowService.CreateFlow(contextWithTimeOut, *req, id)
	if err != nil {
		if errors.Is(err, internal.NotFoundError{}) {
			internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: err.Error()}, http.StatusNotFound)
			return
		}
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Failed to create flow"}, http.StatusInternalServerError)
		return
	}
	internal.WriteJSONResponse(w, r, http.StatusCreated, flow)
}

func (h *flowHandler) GetFlowById(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Invalid flow ID", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeoutCause(r.Context(), 5*time.Second, errors.New("fetching flow timed out"))
	defer cancel()
	flow, err := h.flowService.GetFlowById(contextWithTimeOut, id)
	if err != nil {
		if errors.Is(err, internal.NotFoundError{}) {
			internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: err.Error(), Instance: r.URL.Path}, http.StatusNotFound)
			return
		}
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Failed to fetch flow", Instance: r.URL.Path}, http.StatusInternalServerError)
		return
	}
	internal.WriteJSONResponse(w, r, http.StatusOK, flow)
}

func (h *flowHandler) GetFlowsByProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Invalid product ID", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeoutCause(r.Context(), 5*time.Second, errors.New("fetching flows timed out"))
	defer cancel()
	flows, err := h.flowService.GetFlowsByProduct(contextWithTimeOut, id)
	if err != nil {
		if errors.Is(err, internal.NotFoundError{}) {
			internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: err.Error(), Instance: r.URL.Path}, http.StatusNotFound)
			return
		}
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Failed to fetch flows"}, http.StatusInternalServerError)
		return
	}
	internal.WriteJSONResponse(w, r, http.StatusOK, flows)
}

func (h *flowHandler) UpdateFlow(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Invalid flow ID", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}

	req := &updateFlowRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Invalid request body", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}

	if req.Name == "" && req.Description == "" {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "No updatable fields provided", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}

	contextWithTimeOut, cancel := context.WithTimeoutCause(r.Context(), 5*time.Second, errors.New("updating flow timed out"))
	defer cancel()

	flow, err := h.flowService.UpdateFlow(contextWithTimeOut, *req, id)
	if err != nil {
		if errors.Is(err, internal.NotFoundError{}) {
			internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: err.Error(), Instance: r.URL.Path}, http.StatusNotFound)
			return
		}
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Failed to update flow", Instance: r.URL.Path}, http.StatusInternalServerError)
		return
	}

	internal.WriteJSONResponse(w, r, http.StatusOK, flow)
}

func (h *flowHandler) DeleteFlow(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Invalid flow ID", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}

	contextWithTimeOut, cancel := context.WithTimeoutCause(r.Context(), 5*time.Second, errors.New("deleting flow timed out"))
	defer cancel()

	err := h.flowService.DeleteFlow(contextWithTimeOut, id)
	if err != nil {
		if errors.Is(err, internal.NotFoundError{}) {
			internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: err.Error(), Instance: r.URL.Path}, http.StatusNotFound)
			return
		}
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Failed to delete flow", Instance: r.URL.Path}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *flowHandler) CreateFlowStep(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Invalid flow ID", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}
	req := &createFlowStepRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Invalid request body", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}
	req.FlowId = id

	if req.Current == "" || req.Next == "" {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Current and next UUIDs are required", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}

	if _, err := toPgUUID(req.Current); err != nil {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Invalid current UUID", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}
	if _, err := toPgUUID(req.Next); err != nil {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Invalid next UUID", Instance: r.URL.Path}, http.StatusBadRequest)
		return
	}

	contextWithTimeOut, cancel := context.WithTimeoutCause(r.Context(), 5*time.Second, errors.New("creating flow step timed out"))
	defer cancel()
	flowStep, err := h.flowService.CreateFlowStep(contextWithTimeOut, *req)
	if err != nil {
		if errors.Is(err, internal.NotFoundError{}) {
			internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: err.Error(), Instance: r.URL.Path}, http.StatusNotFound)
			return
		}

		if _, ok := errors.AsType[DependencyValidationError](err); ok {
			internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: err.Error(), Instance: r.URL.Path}, http.StatusUnprocessableEntity)
			return
		}

		if _, ok := errors.AsType[ConflictError](err); ok {
			internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: err.Error(), Instance: r.URL.Path}, http.StatusConflict)
			return
		}

		logger := internal.LoggerFromContext(r.Context())
		logger.Error("Failed to create flow step",
			slog.String("error", err.Error()),
			slog.Int("flow_id", id))
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Failed to create flow step", Instance: r.URL.Path}, http.StatusInternalServerError)
		return
	}
	internal.WriteJSONResponse(w, r, http.StatusCreated, flowStep)
}
