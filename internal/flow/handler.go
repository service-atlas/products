package flow

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"products/internal"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func NewHandler(db DBTX) Handler {
	queries := &Queries{
		db: db,
	}
	return &flowHandler{
		queries: queries,
	}
}

type flowHandler struct {
	queries Querier
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
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Name is required"}, http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeoutCause(r.Context(), 5*time.Second, errors.New("flow creation timed out"))
	defer cancel()
	flow, err := h.queries.CreateFlow(contextWithTimeOut, req.ToParams(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Product not found"}, http.StatusNotFound)
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
	flow, err := h.queries.GetFlow(contextWithTimeOut, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Flow not found", Instance: r.URL.Path}, http.StatusNotFound)
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
	flows, err := h.queries.GetFlowsByProduct(contextWithTimeOut, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "No flows found for product"}, http.StatusNotFound)
			return
		}
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Failed to fetch flows"}, http.StatusInternalServerError)
		return
	}
	internal.WriteJSONResponse(w, r, http.StatusOK, flows)
}
