package platformHandler

type PlatformHandler struct {
	queries PlatformQuerier
}

func NewPlatformHandler(q PlatformQuerier) *PlatformHandler {
	return &PlatformHandler{
		queries: q,
	}
}
