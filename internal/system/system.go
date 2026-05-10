package system

import (
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"net/http"
	"time"
)

var Version = "dev"

func NewSystemCallHandler() *SystemCallHandler {
	logger := slog.Default()
	logger.Info("Software Version: ", "version", Version)
	return &SystemCallHandler{}
}

type SystemCallHandler struct {
}

func (s *SystemCallHandler) GetTime(rw http.ResponseWriter, _ *http.Request) {
	now := time.Now().Format("2006-01-02 15:04:05")
	rw.Header().Set("Content-Type", "text/plain")
	_, err := io.WriteString(rw, now)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *SystemCallHandler) GetVersion(rw http.ResponseWriter, _ *http.Request) {
	resp := struct {
		Version string `json:"version"`
	}{
		Version: Version,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	_, _ = rw.Write(b)
}
