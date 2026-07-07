package capability

import (
	"errors"
	"products/internal/capability/db"
	"strings"
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
	if r.FlowId == 0 {
		return errors.New("flow_id is required")
	}
	return nil
}

type updateCapabilityRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Id          int    `json:"id"`
}

func (r *updateCapabilityRequest) Validate() error {
	if r.Id == 0 {
		return errors.New("id is required")
	}
	if len(strings.TrimSpace(r.Name)) == 0 {
		return errors.New("name is required")
	}
	return nil
}

func (r *updateCapabilityRequest) ToParams() db.UpdateCapabilityParams {
	return db.UpdateCapabilityParams{
		Name: r.Name,
		ID:   r.Id,
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
