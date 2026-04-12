package systemHandler

import (
	"io"
	"log"
	"net/http"
	"time"
)

type SystemCallHandler struct {
}

func (s *SystemCallHandler) GetTime(rw http.ResponseWriter, _ *http.Request) {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := io.WriteString(rw, now)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "text/plain")
}
