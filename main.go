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
	connStr, err := getConnStr()
	if err != nil {
		return nil, err
	}

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		slog.Error("Failed to parse database config", "error", err)
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		return nil, err
	}

	// Verify connection
	if err := pool.Ping(context.Background()); err != nil {
		slog.Error("Failed to ping database", "error", err)
		return nil, err
	}

	slog.Info("Successfully connected to database")
	return pool, nil
}

func getConnStr() (string, error) {

	sProvider, err := secretsprovider.NewProvider()
	if err != nil {
		return "", fmt.Errorf("failed to get secret provider: %w", err)
	}

	ctx := context.Background()
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
