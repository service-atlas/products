package platformHandler

import "products/internal/db/platform"

type PlatformHandler struct {
	queries platform.Querier
}

func NewPlatformHandler(q platform.Querier) *PlatformHandler {
	return &PlatformHandler{
		queries: q,
	}
}
