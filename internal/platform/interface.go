package platform

import "net/http"

type Handler interface {
	CreatePlatform(w http.ResponseWriter, r *http.Request)
	GetPlatforms(w http.ResponseWriter, r *http.Request)
	GetPlatform(w http.ResponseWriter, r *http.Request)
	UpdatePlatform(w http.ResponseWriter, r *http.Request)
	DeletePlatform(w http.ResponseWriter, r *http.Request)
}
