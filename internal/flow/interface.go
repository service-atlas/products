package flow

import "net/http"

type Handler interface {
	CreateFlow(w http.ResponseWriter, r *http.Request)
	GetFlowById(w http.ResponseWriter, r *http.Request)
	GetFlowsByProduct(w http.ResponseWriter, r *http.Request)
}
