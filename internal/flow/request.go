package flow

import (
	"products/internal/flow/db"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type createFlowRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r *createFlowRequest) ToParams(productID int) db.CreateFlowParams {
	return db.CreateFlowParams{
		Name:      r.Name,
		ProductID: productID,
		Description: pgtype.Text{
			Valid:  r.Description != "",
			String: r.Description,
		},
		TimeStamp: pgtype.Timestamptz{
			Valid: true,
			Time:  time.Now().UTC(),
		},
	}
}
