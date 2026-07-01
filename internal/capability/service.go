package capability

import (
	"products/internal/capability/db"
)

func newPostgresService(dbConn db.DBTX) capabilityService {
	queries := db.New(dbConn)
	return &postgresService{
		queries: queries,
	}
}

type postgresService struct {
	queries db.Querier
}
