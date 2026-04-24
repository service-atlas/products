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
	"products/internal/db"
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

	if dbPool, ok := dbConn.(*pgxpool.Pool); ok {
		defer dbPool.Close()
	}

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

func getDbConn() (db.DBTX, error) {
	user := internalConfig.GetConfigValue("DB_USERNAME")
	pass := internalConfig.GetConfigValue("DB_PASSWORD")
	dbHostPort := internalConfig.GetConfigValue("DB_URL")

	if user == "" || pass == "" || dbHostPort == "" {
		slog.Error("Database environment variables DB_USERNAME, DB_PASSWORD, or DB_URL are not set")
		return nil, errors.New("database environment variables not set")
	}

	if !strings.Contains(dbHostPort, "://") {
		dbHostPort = "postgres://" + dbHostPort
	}
	u, err := url.Parse(dbHostPort)
	if err != nil {
		slog.Error("Failed to parse DB_URL", "error", err)
		return nil, err
	}
	u.Scheme = "postgres"
	u.User = url.UserPassword(user, pass)

	connStr := u.String()
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
