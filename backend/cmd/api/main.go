package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/cvsz/cfc-bot-manager/backend/internal/config"
	"github.com/cvsz/cfc-bot-manager/backend/internal/httpapi"
)

var version = "dev"

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(2)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	server := &http.Server{
		Addr: cfg.HTTPAddr,
		Handler: httpapi.NewHandler(version, nil),
		ReadHeaderTimeout: cfg.ReadHeader,
		ReadTimeout: cfg.Read,
		WriteTimeout: cfg.Write,
		IdleTimeout: cfg.Idle,
		MaxHeaderBytes: 1 << 20,
	}

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("HTTP server starting", "address", cfg.HTTPAddr, "environment", cfg.Environment, "version", version)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Shutdown)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
			_ = server.Close()
			os.Exit(1)
		}
		if err := <-serveErr; err != nil {
			logger.Error("HTTP server stopped with error", "error", err)
			os.Exit(1)
		}
		logger.Info("HTTP server stopped")
	case err := <-serveErr:
		if err != nil {
			logger.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
		logger.Error("HTTP server stopped before shutdown was requested")
		os.Exit(1)
	}
}
