package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
)

type NotFoundError struct {
	id       int
	itemType string
}

func (e NotFoundError) Error() string {
	return e.itemType + " not found with ID: " + strconv.Itoa(e.id)
}

func (e NotFoundError) Is(target error) bool {
	var notFoundError NotFoundError
	ok := errors.As(target, &notFoundError)
	return ok
}

func NewNotFoundError(id int, itemType string) NotFoundError {
	return NotFoundError{id: id, itemType: itemType}
}

// WriteJSONResponse encodes the data to JSON and writes it to the response writer.
// It uses a buffer to ensure that encoding is successful before writing the header.
func WriteJSONResponse(w http.ResponseWriter, r *http.Request, status int, data any) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		HandleHttpError(w, ErrorEnvelope{Detail: "Failed to encode response"}, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(buf.Bytes()); err != nil {
		slog.Error("Failed to write response", "request", r.URL.Path, "error", err)
	}
}
