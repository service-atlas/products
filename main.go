package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	internalConfig "products/internal/config"
	"products/router"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	dbConn, err := getDbConn()
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	r := router.SetupRouter(dbConn)
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
	user := internalConfig.GetConfigValue("DB_USERNAME")
	pass := internalConfig.GetConfigValue("DB_PASSWORD")
	dbHostPort := internalConfig.GetConfigValue("DB_URL")

	if user == "" || pass == "" || dbHostPort == "" {
		slog.Error("Database environment variables DB_USERNAME, DB_PASSWORD, or DB_URL are not set")
		return "", errors.New("database environment variables not set")
	}

	if !strings.Contains(dbHostPort, "://") {
		dbHostPort = "postgres://" + dbHostPort
	}
	u, err := url.Parse(dbHostPort)
	if err != nil {
		slog.Error("Failed to parse DB_URL", "error", err)
		return "", err
	}
	u.Scheme = "postgres"
	u.User = url.UserPassword(user, pass)

	return u.String(), nil
}
