package flow

func NewHandler(db DBTX) Handler {
	queries := &Queries{
		db: db,
	}
	return &flowHandler{
		queries: queries,
	}
}

type flowHandler struct {
	queries Querier
}
