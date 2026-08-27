package internal

import (
	"net/http"
	"strconv"
	"time"
	"uuid"

	"github.com/go-chi/chi/v5"
)

type PathValidator func(string, *http.Request) (string, bool)

func getPathValue(req *http.Request, varName string) string {
	val := req.PathValue(varName)
	if val == "" {
		val = chi.URLParam(req, varName)
	}
	return val
}

func GetGuidFromRequestPath(varName string, req *http.Request) (string, bool) {
	guidVal := getPathValue(req, varName)
	return IsValidGuid(guidVal)
}

func IsValidGuid(guidVal string) (string, bool) {
	_, err := uuid.Parse(guidVal)
	return guidVal, err == nil
}

func GetDateFromRequestPath(varName string, req *http.Request) (time.Time, bool) {
	dateVal := getPathValue(req, varName)
	date, err := time.Parse("2006-01-02", dateVal)
	return date, err == nil
}

func GetIntFromRequestPath(varName string, req *http.Request) (int, bool) {
	val := getPathValue(req, varName)
	if val == "" {
		return 0, false
	}
	id, err := strconv.Atoi(val)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}
