package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	internalConfig "products/internal/config" //nolint:depguard
	"products/router"                         //nolint:depguard

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool" //nolint:depguard
	"github.com/service-atlas/secrets-provider"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dbConn, err := getDbConn()
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	r := router.InitializeRouter(dbConn)
	addr := internalConfig.GetConfigValue("ADDRESS")

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

}

func getDbConn() (*pgxpool.Pool, error) {
	sProvider, err := secretsprovider.NewProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get secret provider: %w", err)
	}
	ctx := context.Background()

	conStrCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	connStr, err := getConnStr(conStrCtx, sProvider)
	if err != nil {
		return nil, err
	}

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		slog.Error("Failed to parse database config", "error", err)
		return nil, err
	}

	config.BeforeConnect = func(ctx context.Context, cfg *pgx.ConnConfig) error {
		newConnStr, err := getConnStr(ctx, sProvider)
		if err != nil {
			slog.Error("Failed to get updated connection string", "error", err)
			return err
		}
		updatedCfg, err := pgx.ParseConfig(newConnStr)
		if err != nil {
			slog.Error("Failed to parse updated connection string", "error", err)
			return err
		}
		cfg.User = updatedCfg.User
		cfg.Password = updatedCfg.Password
		return nil
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		return nil, err
	}

	// Verify connection
	pingCtx, pCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pCancel()
	if err := pool.Ping(pingCtx); err != nil {
		slog.Error("Failed to ping database", "error", err)
		return nil, err
	}

	slog.Info("Successfully connected to database")
	return pool, nil
}

func getConnStr(ctx context.Context, sProvider secretsprovider.Provider) (string, error) {
	dbInfo, err := sProvider.GetDatabaseInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get database info from secret provider: %w", err)
	}

	if !strings.Contains(dbInfo.URL, "://") {
		dbInfo.URL = "postgres://" + dbInfo.URL
	}
	u, err := url.Parse(dbInfo.URL)
	if err != nil {
		slog.Error("Failed to parse DB_URL", "error", err)
		return "", err
	}
	u.Scheme = "postgres"
	u.User = url.UserPassword(dbInfo.Username, dbInfo.Password)

	return u.String(), nil
}
