package platform

import (
	"net/http"
)

type platformHandler struct {
	queries *Queries
}

func NewPlatformHandler(db DBTX) Handler {
	queries := &Queries{
		db: db,
	}
	return &platformHandler{
		queries: queries,
	}
}

type Handler interface {
	CreatePlatform(w http.ResponseWriter, r *http.Request)
	GetPlatforms(w http.ResponseWriter, r *http.Request)
	GetPlatform(w http.ResponseWriter, r *http.Request)
	UpdatePlatform(w http.ResponseWriter, r *http.Request)
	DeletePlatform(w http.ResponseWriter, r *http.Request)
}
