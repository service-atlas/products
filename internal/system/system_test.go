package system

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

func TestSystemGetVersion(t *testing.T) {
	req, err := http.NewRequest("GET", "/version", nil)
	if err != nil {
		t.Fatal(err)
	}
	rw := httptest.NewRecorder()

	h := SystemCallHandler{}
	h.GetVersion(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("GetVersion errored with %d", rw.Code)
	}

	expected := `{"version":"dev"}`
	if rw.Body.String() != expected {
		t.Errorf("GetVersion returned unexpected body: got %q want %q", rw.Body.String(), expected)
	}
}

func TestNewSystemCallHandler(t *testing.T) {
	h := NewSystemCallHandler()
	if h == nil {
		t.Fatal("NewSystemCallHandler returned nil")
	}
}
