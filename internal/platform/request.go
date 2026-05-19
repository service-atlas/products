package platform

import (
	"products/internal/platform/db"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type createPlatformRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r *createPlatformRequest) ToParams() db.CreatePlatformParams {
	return db.CreatePlatformParams{
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

func (r *updatePlatformRequest) ToParams(id int) db.UpdatePlatformParams {
	return db.UpdatePlatformParams{
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
