package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"products/internal"
	"products/internal/flow/db"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type flow interface {
	CreateFlow(ctx context.Context, req createFlowRequest, id int) (db.Flow, error)
	GetFlowById(ctx context.Context, id int) (db.Flow, error)
	GetFlowsByProduct(ctx context.Context, id int) ([]db.Flow, error)
	UpdateFlow(ctx context.Context, req updateFlowRequest, id int) (db.Flow, error)
	DeleteFlow(ctx context.Context, id int) error
}

type flowStep interface {
	CreateFlowStep(ctx context.Context, req createFlowStepRequest) (db.FlowStep, error)
	UpdateFlowStep(ctx context.Context, req updateFlowStepRequest, id int) (db.FlowStep, error)
	DeleteFlowStep(ctx context.Context, id int) error
	GetFlowSteps(ctx context.Context, id int) ([]db.FlowStep, error)
	GetFlowPath(ctx context.Context, id int) (FlowPath, error)
}

type flowService interface {
	flow
	flowStep
}

type postgresService struct {
	queries db.Querier
	client  *http.Client
}

func (s *postgresService) validateDependency(ctx context.Context, current, next string) (bool, error) {
	serviceUrl := os.Getenv("SERVICE_URL")
	if serviceUrl == "" {
		return false, fmt.Errorf("SERVICE_URL is not set")
	}
	serviceUrl = strings.TrimSuffix(serviceUrl, "/")

	url := fmt.Sprintf("%s/services/%s/dependencies", serviceUrl, current)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create dependency request: %w", err)
	}
	resp, err := s.client.Do(httpReq)
	if err != nil {
		return false, fmt.Errorf("failed to fetch dependencies: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			slog.Default().Error("failed to close body in dependency validation", "error", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to fetch dependencies: status %d", resp.StatusCode)
	}

	var dependencies []serviceDependency
	err = json.NewDecoder(resp.Body).Decode(&dependencies)
	if err != nil {
		return false, fmt.Errorf("failed to decode dependencies: %w", err)
	}

	for _, dep := range dependencies {
		if strings.EqualFold(dep.Id, next) && dep.InteractionType == "data" {
			return true, nil
		}
	}

	return false, nil
}

func (s *postgresService) CreateFlow(ctx context.Context, req createFlowRequest, id int) (db.Flow, error) {

	flow, err := s.queries.CreateFlow(ctx, req.ToParams(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Flow{}, internal.NewNotFoundError(id, "Product")
		}
		return db.Flow{}, fmt.Errorf("failed to create flow: %w", err)
	}
	return flow, nil
}

func (s *postgresService) GetFlowById(ctx context.Context, id int) (db.Flow, error) {
	flow, err := s.queries.GetFlow(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Flow{}, internal.NewNotFoundError(id, "Flow")
		}
		return db.Flow{}, fmt.Errorf("failed to fetch flow: %w", err)
	}
	return flow, nil
}

func (s *postgresService) GetFlowsByProduct(ctx context.Context, id int) ([]db.Flow, error) {
	_, err := s.queries.GetProductById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, internal.NewNotFoundError(id, "Product")
		}
		return nil, fmt.Errorf("failed to fetch flows: %w", err)
	}
	flows, err := s.queries.GetFlowsByProduct(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []db.Flow{}, nil
		}
		return nil, fmt.Errorf("failed to fetch flows: %w", err)
	}
	if flows == nil {
		flows = []db.Flow{}
	}
	return flows, nil
}

func (s *postgresService) UpdateFlow(ctx context.Context, req updateFlowRequest, id int) (db.Flow, error) {
	existing, err := s.GetFlowById(ctx, id)
	if err != nil {
		return db.Flow{}, err
	}

	rowsAffected, err := s.queries.UpdateFlow(ctx, req.ToParams(id, existing))
	if err != nil {
		return db.Flow{}, fmt.Errorf("failed to update flow: %w", err)
	}
	if rowsAffected == 0 {
		return db.Flow{}, internal.NewNotFoundError(id, "Flow")
	}

	return s.GetFlowById(ctx, id)
}
func (s *postgresService) DeleteFlow(ctx context.Context, id int) error {
	rowsAffected, err := s.queries.DeleteFlow(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete flow: %w", err)
	}
	if rowsAffected == 0 {
		return internal.NewNotFoundError(id, "Flow")
	}
	return nil
}

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

func (s *postgresService) UpdateFlowStep(ctx context.Context, req updateFlowStepRequest, id int) (db.FlowStep, error) {
	existing, err := s.queries.GetFlowStep(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.FlowStep{}, internal.NewNotFoundError(id, "FlowStep")
		}
		return db.FlowStep{}, fmt.Errorf("failed to fetch flow step: %w", err)
	}

	params := req.ToParams(id, existing)
	rowsAffected, err := s.queries.UpdateFlowStep(ctx, params)
	if err != nil {
		return db.FlowStep{}, fmt.Errorf("failed to update flow step: %w", err)
	}
	if rowsAffected == 0 {
		return db.FlowStep{}, internal.NewNotFoundError(id, "FlowStep")
	}

	updated, err := s.queries.GetFlowStep(ctx, id)
	if err != nil {
		return db.FlowStep{}, fmt.Errorf("failed to fetch updated flow step: %w", err)
	}

	return updated, nil
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
