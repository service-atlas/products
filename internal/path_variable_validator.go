package internal

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type PathValidator func(string, *http.Request) (string, bool)

func GetGuidFromRequestPath(varName string, req *http.Request) (string, bool) {
	guidVal := req.PathValue(varName)
	return IsValidGuid(guidVal)
}

func IsValidGuid(guidVal string) (string, bool) {
	err := uuid.Validate(guidVal)
	return guidVal, err == nil
}

func GetDateFromRequestPath(varName string, req *http.Request) (time.Time, bool) {
	dateVal := req.PathValue(varName)
	date, err := time.Parse("2006-01-02", dateVal)
	return date, err == nil
}

func GetIntFromRequestPath(varName string, req *http.Request) (int32, bool) {
	val := req.PathValue(varName)
	if val == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(val, 10, 32)
	if err != nil || id <= 0 {
		return 0, false
	}
	return int32(id), true
}
