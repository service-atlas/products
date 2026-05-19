package platform

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"products/internal"
	"time"

	"github.com/jackc/pgx/v5"
)

func NewPlatformHandler(db DBTX) Handler {
	queries := &Queries{
		db: db,
	}
	return &platformHandler{
		queries: queries,
	}
}

type platformHandler struct {
	queries Querier
}

func (h *platformHandler) CreatePlatform(w http.ResponseWriter, r *http.Request) {
	var req createPlatformRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Invalid request body"}, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Name is required"}, http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if err := h.queries.CreatePlatform(contextWithTimeOut, req.ToParams()); err != nil {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Failed to create platform"}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *platformHandler) UpdatePlatform(w http.ResponseWriter, r *http.Request) {
	var req updatePlatformRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Invalid request body"}, http.StatusBadRequest)
		return
	}
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Invalid platform ID"}, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Name is required"}, http.StatusBadRequest)
		return
	}

	if req.ID != id {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Platform ID does not match path"}, http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	_, err := h.queries.UpdatePlatform(contextWithTimeOut, req.ToParams(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Platform not found"}, http.StatusNotFound)
			return
		}
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Failed to update platform"}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *platformHandler) DeletePlatform(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Invalid platform ID"}, http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	_, err := h.queries.DeletePlatform(contextWithTimeOut, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Platform not found"}, http.StatusNotFound)
			return
		}
		logger := internal.LoggerFromContext(r.Context())
		logger.Error("Failed to delete platform", "error", err, "platform_id", id)
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Internal server error"}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *platformHandler) GetPlatforms(w http.ResponseWriter, r *http.Request) {
	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	platforms, err := h.queries.GetPlatforms(contextWithTimeOut)
	if err != nil {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Failed to fetch platforms"}, http.StatusInternalServerError)
		return
	}
	if platforms == nil {
		platforms = []Platform{}
	}
	internal.WriteJSONResponse(w, r, http.StatusOK, platforms)
}

func (h *platformHandler) GetPlatform(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Invalid platform ID"}, http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	platform, err := h.queries.GetPlatform(contextWithTimeOut, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Platform not found"}, http.StatusNotFound)
			return
		}
		internal.HandleHttpError(w, internal.ErrorEnvelope{Detail: "Failed to fetch platform"}, http.StatusInternalServerError)
		return
	}

	internal.WriteJSONResponse(w, r, http.StatusOK, platform)
}
