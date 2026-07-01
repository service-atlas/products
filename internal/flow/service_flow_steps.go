package flow

import (
	"context"
	"errors"
	"fmt"
	"products/internal"
	"products/internal/flow/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *postgresService) CreateFlowStep(ctx context.Context, req createFlowStepRequest) (db.FlowStep, error) {
	_, err := s.GetFlowById(ctx, req.FlowId)
	if err != nil {
		return db.FlowStep{}, err
	}
	params, err := req.ToParams()
	if err != nil {
		return db.FlowStep{}, err
	}

	ok, err := s.validateDependency(ctx, req.Current, req.Next)
	if err != nil {
		return db.FlowStep{}, err
	}
	if !ok {
		return db.FlowStep{}, DependencyValidationError{}
	}

	flowStep, err := s.queries.CreateFlowStep(ctx, params)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return db.FlowStep{}, ConflictError{Message: "Flow step already exists"}
		}
		return db.FlowStep{}, fmt.Errorf("failed to create flow step: %w", err)
	}
	return flowStep, nil
}

func (s *postgresService) DeleteFlowStep(ctx context.Context, id int) error {
	rowsAffected, err := s.queries.DeleteFlowStep(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete flow step: %w", err)
	}
	if rowsAffected == 0 {
		return internal.NewNotFoundError(id, "FlowStep")
	}
	return nil
}

func (s *postgresService) GetFlowSteps(ctx context.Context, id int) ([]db.FlowStep, error) {
	_, err := s.GetFlowById(ctx, id)
	if err != nil {
		return nil, err
	}
	flowSteps, err := s.queries.GetFlowSteps(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []db.FlowStep{}, nil
		}
		return nil, err
	}
	if flowSteps == nil {
		flowSteps = []db.FlowStep{}
	}
	return flowSteps, nil
}

func (s *postgresService) GetFlowPath(ctx context.Context, id int) (FlowPath, error) {
	flowSteps, err := s.GetFlowSteps(ctx, id)
	if err != nil {
		return FlowPath{}, err
	}
	pathMap := make(map[string][]string)
	nextSet := make(map[string]bool)
	for _, step := range flowSteps {
		current, next := step.Current.String(), step.Next.String()
		if _, ok := pathMap[current]; !ok {
			pathMap[current] = []string{}
		}
		pathMap[current] = append(pathMap[current], next)
		nextSet[next] = true
	}
	var queue []string
	for k := range pathMap {
		if _, ok := nextSet[k]; !ok {
			queue = append(queue, k)
		}
	}

	if len(queue) == 0 && len(pathMap) > 0 {
		return FlowPath{}, fmt.Errorf("no entry point found in flow")
	}
	if len(queue) > 1 {
		return FlowPath{}, fmt.Errorf("multiple entry points found in flow")
	}
	var path []PathItem
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if nexts, ok := pathMap[current]; ok {
			queue = append(queue, nexts...)
			path = append(path, PathItem{Current: current, Next: nexts})
		}
	}

	return FlowPath{FlowID: id, Path: path}, nil
}
