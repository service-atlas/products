package internal

import (
	"encoding/json"
	"net/http"
)

type ErrorEnvelope struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status,omitzero"`
	Detail   string `json:"detail,omitzero"`
	Instance string `json:"instance,omitzero"`
}

func HandleHttpError(w http.ResponseWriter, err ErrorEnvelope, statusCode int) {
	if err.Type == "" {
		err.Type = "about:blank"
	}
	if err.Title == "" {
		err.Title = http.StatusText(statusCode)
	}
	err.Status = statusCode

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(err)
}
