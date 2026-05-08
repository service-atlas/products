package platform

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type createPlatformRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r *createPlatformRequest) ToParams() CreatePlatformParams {
	return CreatePlatformParams{
		Name: r.Name,
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

type updatePlatformRequest struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r *updatePlatformRequest) ToParams(id int) UpdatePlatformParams {
	return UpdatePlatformParams{
		ID:   id,
		Name: r.Name,
		Description: pgtype.Text{
			Valid:  r.Description != "",
			String: r.Description,
		},
		UpdatedAt: pgtype.Timestamptz{
			Valid: true,
			Time:  time.Now().UTC(),
		},
	}
}
