package capability

import (
	"errors"
	"products/internal/capability/db"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type createCapabilityRequest struct {
	ProductId   int    `json:"product_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r *createCapabilityRequest) ToParams() db.CreateCapabilityParams {
	return db.CreateCapabilityParams{
		ProductID: r.ProductId,
		Name:      r.Name,
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
	if len(strings.TrimSpace(r.Name)) == 0 {
		return errors.New("name is required")
	}
	if r.ProductId == 0 {
		return errors.New("product_id is required")
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

type createCapabilityStepRequest struct {
	FlowStepId   int    `json:"flow_step_id"`
	CapabilityId int    `json:"capability_id"`
	Target       string `json:"target"`
	Protocol     string `json:"protocol"`
}

func (r *createCapabilityStepRequest) Validate() error {
	if r.CapabilityId == 0 {
		return errors.New("capability_id is required")
	}
	if r.FlowStepId == 0 {
		return errors.New("flow_step_id is required")
	}
	if len(strings.TrimSpace(r.Target)) == 0 {
		return errors.New("target is required")
	}
	if len(strings.TrimSpace(r.Protocol)) == 0 {
		return errors.New("protocol is required")
	}
	return nil
}

func (r *createCapabilityStepRequest) ToParams() db.CreateCapabilityStepParams {
	return db.CreateCapabilityStepParams{
		CapabilityID: r.CapabilityId,
		FlowStepID:   r.FlowStepId,
		Target: pgtype.Text{
			Valid:  r.Target != "",
			String: r.Target,
		},
		Protocol: pgtype.Text{
			Valid:  r.Protocol != "",
			String: r.Protocol,
		},
		Timestamp: pgtype.Timestamptz{
			Valid: true,
			Time:  time.Now().UTC(),
		},
	}
}
