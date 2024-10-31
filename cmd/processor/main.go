package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/denpeshkov/receipt-processor/internal/server"
	"github.com/denpeshkov/receipt-processor/internal/storage"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running service: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	fs := flag.NewFlagSet("processor", flag.ExitOnError)
	addr := fs.String("addr", "localhost:8080", "address to listen on")
	debugMode := fs.Bool("debug", false, "start in debug mode")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	storage := storage.New()
	server := server.New(storage, server.Options{Logger: logger.With("component", "server"), DebugMode: *debugMode})

	httpServer := &http.Server{
		Addr:              *addr,
		Handler:           server,
		ReadHeaderTimeout: 3 * time.Second,
		ErrorLog:          slog.NewLogLogger(logger.With("component", "http_server").Handler(), slog.LevelError),
	}

	logger.Info("Started application", "addr", httpServer.Addr, "debug", *debugMode)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		<-ctx.Done()

		ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		httpServer.SetKeepAlivesEnabled(false)
		if err := httpServer.Shutdown(ctxShutdown); err != nil {
			return fmt.Errorf("http server shutdown: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server listen: %w", err)
		}
		return nil
	})
	return g.Wait()
}
