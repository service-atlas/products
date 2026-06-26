package capability

import (
	"errors"
	"products/internal/capability/db"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type createCapabilityRequest struct {
	FlowId      int    `json:"flow_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r *createCapabilityRequest) ToParams() db.CreateCapabilityParams {
	return db.CreateCapabilityParams{
		FlowID: r.FlowId,
		Name:   r.Name,
		Description: pgtype.Text{
			Valid:  r.Description != "",
			String: r.Description,
		},
		Timestamp: pgtype.Timestamptz{
			Valid: true,
			Time:  time.Now().UTC(),
		},
	}
}

func (r *createCapabilityRequest) Validate() error {
	if len(r.Name) == 0 {
		return errors.New("name is required")
	}
	return nil
}
