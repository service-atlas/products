package capability

import (
	"products/internal/capability/db"
)

type Handler interface {
}

func NewHandler(dbConn db.DBTX) Handler {
	queries := db.New(dbConn)
	service := &postgresService{
		queries: queries,
	}
	return &capabilityHandler{
		service: service,
	}
}

type capabilityHandler struct {
	service capabilityService
}
