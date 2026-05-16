package flow

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type createFlowRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r *createFlowRequest) ToParams(productID int) CreateFlowParams {
	return CreateFlowParams{
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
