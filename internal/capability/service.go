package capability

import (
	"products/internal/capability/db"
)

type capabilityService interface {
	capabilityStepService
}

type capabilityStepService interface {
}

type postgresService struct {
	queries db.Querier
}
