package sampleHandler

import (
	service "chi-boilerplate/service/sample"
	"net/http"
)

func NewSampleCallHandler(s *service.SampleService) *SampleCallHandler {
	return &SampleCallHandler{
		service: s,
	}
}

type SampleCallHandler struct {
	service *service.SampleService
}

func (s *SampleCallHandler) GetSample(rw http.ResponseWriter, _ *http.Request) {
	d, err := s.service.GetSample()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(d))
}

func (s *SampleCallHandler) GetError(rw http.ResponseWriter, _ *http.Request) {
	_, err := s.service.GetError()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	//this code will never be hit
	rw.WriteHeader(http.StatusOK)
}
