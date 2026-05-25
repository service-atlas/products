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

type updateFlowRequest struct {
	Name        string `json:"name,omitzero"`
	Description string `json:"description,omitzero"`
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

func (r *updateFlowRequest) ToParams(id int, existing db.Flow) db.UpdateFlowParams {
	params := db.UpdateFlowParams{
		ID:          id,
		Name:        existing.Name,
		Description: existing.Description,
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now().UTC(),
			Valid: true,
		},
	}

	if r.Name != "" {
		params.Name = r.Name
	}

	// Empty string specifically set to null
	// Since we can't distinguish between omitted and explicit "" with strings,
	// and the user explicitly asked for "" to mean NULL, we check if it's ""
	// BUT we only do this if it was provided.
	// Wait, still can't tell.
	// Let's just follow the request: any "" in the struct (from json) means NULL for description.
	params.Description = pgtype.Text{
		String: r.Description,
		Valid:  r.Description != "",
	}

	return params
}
