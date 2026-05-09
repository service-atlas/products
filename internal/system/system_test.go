package systemHandler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSystemGetTime(t *testing.T) {
	req, err := http.NewRequest("GET", "/time", nil)
	rw := httptest.NewRecorder()

	h := SystemCallHandler{}
	h.GetTime(rw, req)
	if err != nil {
		t.Errorf("GetTime errored with %s", err.Error())
	}
	if rw.Code != http.StatusOK {
		t.Errorf("GetTime errored with %s", string(rune(rw.Code)))
	}
	r_time, t_err := time.Parse("2006-01-02 15:04:05", rw.Body.String())
	if t_err != nil {
		t.Errorf("Time returned by GetTime errored with %s", t_err.Error())
	}
	if !time.Now().After(r_time) {
		t.Errorf("GetTime time is not before current time, return val: %s", rw.Body.String())
	}
}
