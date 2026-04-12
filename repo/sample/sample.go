package sampleRepo

import "errors"

type SampleRepo struct {
	db string
}

func NewSampleRepo(s string) *SampleRepo {
	return &SampleRepo{
		db: s,
	}
}

func (s *SampleRepo) GetSample() (string, error) {
	return s.db, nil
}

func (s *SampleRepo) GetError() (string, error) {
	return "", errors.New(s.db + " error")
}
