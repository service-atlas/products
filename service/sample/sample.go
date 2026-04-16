package sampleService

import repo "products/repo/sample"

type SampleService struct {
	repo *repo.SampleRepo
}

func NewSampleService(r *repo.SampleRepo) *SampleService {
	return &SampleService{
		repo: r,
	}
}

func (s *SampleService) GetSample() (string, error) {
	return s.repo.GetSample()
}

func (s *SampleService) GetError() (string, error) {
	return s.repo.GetError()
}
