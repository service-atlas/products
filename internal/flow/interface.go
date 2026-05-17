package flow

import "net/http"

type Handler interface {
	CreateFlow(w http.ResponseWriter, r *http.Request)
	GetFlowById(w http.ResponseWriter, r *http.Request)
}
