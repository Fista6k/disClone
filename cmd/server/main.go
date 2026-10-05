package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fista6k/disClone/internal"
	"github.com/Fista6k/disClone/internal/app"
	"github.com/Fista6k/disClone/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg := config.Load()

	storage, err := internal.ConnToStorage(ctx)
	if err != nil {
		return err
	}

	application := app.New(storage, cfg.JWTSecret)
	defer func() {
		if err := application.Close(); err != nil {
			slog.Error("close storage", "err", err)
		}
	}()

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           application.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errChan := make(chan error, 1)

	go func() {
		slog.Info("server started", "addr", cfg.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		slog.Info("shutting down")
	}

	shutdownCtx, cancelShutDown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutDown()

	return server.Shutdown(shutdownCtx)
}
