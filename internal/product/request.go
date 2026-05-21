package product

import (
	"products/internal/product/db"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type createProductRequest struct {
	Name        string `json:"name"`
	PlatformID  int    `json:"platform_id"`
	Description string `json:"description"`
}

func (r *createProductRequest) ToParams() db.CreateProductParams {
	return db.CreateProductParams{
		Name:       r.Name,
		PlatformID: r.PlatformID,
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

type updateProductRequest struct {
	PlatformID  int    `json:"platform_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r *updateProductRequest) ToParams(id int) db.UpdateProductParams {
	return db.UpdateProductParams{
		ID:         id,
		PlatformID: r.PlatformID,
		Name:       r.Name,
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
