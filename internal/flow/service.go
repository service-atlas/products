package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"products/internal/flow/db"
	"strings"
	"time"
)

func newPostgresService(dbConn db.DBTX) *postgresService {
	queries := db.New(dbConn)
	client := &http.Client{Timeout: 5 * time.Second}
	return &postgresService{
		queries: queries,
		client:  client,
	}
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
