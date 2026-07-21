package platform

import (
	"context"
	"encoding/json"
	"net/http"
	"products/internal"
	"products/internal/platform/db"
	"time"

	"github.com/service-atlas/go-common/errorenvelope"
	"github.com/service-atlas/go-common/httphelpers"
	"github.com/service-atlas/go-common/httplog"
)

func newHandler(svc platformService) *handler {
	return &handler{
		service: svc,
	}
}

type handler struct {
	service platformService
}

func (h *handler) CreatePlatform(w http.ResponseWriter, r *http.Request) {
	var req createPlatformRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid request body"}, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Name is required"}, http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	platform, err := h.service.CreatePlatform(contextWithTimeOut, req)
	if err != nil {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to create platform"}, http.StatusInternalServerError)
		return
	}

	httphelpers.WriteJSONResponse(w, r, http.StatusCreated, platform)
}

func (h *handler) UpdatePlatform(w http.ResponseWriter, r *http.Request) {
	var req updatePlatformRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid request body"}, http.StatusBadRequest)
		return
	}
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid platform ID"}, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Name is required"}, http.StatusBadRequest)
		return
	}

	if req.ID != id {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Platform ID does not match path"}, http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	_, err := h.service.UpdatePlatform(contextWithTimeOut, req, id)
	if err != nil {
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Platform not found"}, http.StatusNotFound)
			return
		}
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to update platform"}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) DeletePlatform(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid platform ID"}, http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	_, err := h.service.DeletePlatform(contextWithTimeOut, id)
	if err != nil {
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Platform not found"}, http.StatusNotFound)
			return
		}
		httplog.LoggerFromContext(r.Context()).Error("Failed to delete platform", "error", err, "platform_id", id)
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Internal server error"}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) GetPlatforms(w http.ResponseWriter, r *http.Request) {
	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	platforms, err := h.service.GetPlatforms(contextWithTimeOut)
	if err != nil {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to fetch platforms"}, http.StatusInternalServerError)
		return
	}
	if platforms == nil {
		platforms = []db.Platform{}
	}
	httphelpers.WriteJSONResponse(w, r, http.StatusOK, platforms)
}

func (h *handler) GetPlatform(w http.ResponseWriter, r *http.Request) {
	id, ok := internal.GetIntFromRequestPath("id", r)
	if !ok {
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Invalid platform ID"}, http.StatusBadRequest)
		return
	}
	contextWithTimeOut, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	platform, err := h.service.GetPlatform(contextWithTimeOut, id)
	if err != nil {
		if internal.IsNotFoundError(err) {
			errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Platform not found"}, http.StatusNotFound)
			return
		}
		errorenvelope.HandleHttpError(w, errorenvelope.ErrorEnvelope{Detail: "Failed to fetch platform"}, http.StatusInternalServerError)
		return
	}

	httphelpers.WriteJSONResponse(w, r, http.StatusOK, platform)
}
