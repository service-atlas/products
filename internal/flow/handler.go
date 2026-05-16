package flow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
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
	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 5*time.Second)
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

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(flow)
	if err != nil {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Failed to encode response"}, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(buf.Bytes())
	if err != nil {
		slog.Error("Failed to write response", "request", r.URL.Path, "error", err)
	}
}
