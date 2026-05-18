package flow

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"products/internal"
	"products/internal/flow/db"
	"strings"
	"time"
)

func NewHandler(dbConn db.DBTX) Handler {
	queries := db.New(dbConn)
	service := &service{
		queries: queries,
	}
	return &flowHandler{
		flowService: service,
	}
}

type flowService interface {
	CreateFlow(ctx context.Context, req createFlowRequest, id int) (db.Flow, error)
	GetFlowById(ctx context.Context, id int) (db.Flow, error)
	GetFlowsByProduct(ctx context.Context, id int) ([]db.Flow, error)
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
