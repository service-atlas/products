package flow

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type createFlowRequest struct {
	Name        string `json:"name"`
	ProductID   int    `json:"product_id"`
	Description string `json:"description"`
}

func (r *createFlowRequest) ToParams() CreateFlowParams {
	return CreateFlowParams{
		Name:      r.Name,
		ProductID: r.ProductID,
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
