// Package server provides a server (router) for handling receipts.
package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	hpprof "net/http/pprof"

	"github.com/denpeshkov/receipt-processor/internal/receipt"
)

// Storage represents a storage for receipts. It should be save for use by multiple goroutines.
type Storage interface {
	//  Store saves a new receipt in the storage, generating a unique ID for it.
	Store(ctx context.Context, receipt *receipt.Receipt) error
	// Get retrieves a receipt from the storage by its ID.
	Get(ctx context.Context, id string) (*receipt.Receipt, error)
}

// Options defines the configuration options for the [Server].
type Options struct {
	Logger    *slog.Logger // Logger used for server logging.
	DebugMode bool         // Enables http/pprof debugging endpoints if set to true.
}

// Server is a server for processing receipts.
type Server struct {
	storage Storage
	logger  *slog.Logger
	h       http.Handler
}

// New returns a new configured server with provided receipts storage.
func New(storage Storage, opts Options) *Server {
	if opts.Logger == nil {
		opts.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	s := &Server{
		storage: storage,
		logger:  opts.Logger,
	}
	s.h = s.handler(opts.DebugMode)
	return s
}

func (s *Server) handler(debugMode bool) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /receipts/process", s.errorHandler(s.handleReceipt))
	mux.HandleFunc("GET /receipts/{id}/points", s.errorHandler(s.handlePoints))
	if debugMode {
		mux.HandleFunc("GET /debug/pprof/", hpprof.Index)
		mux.HandleFunc("GET /debug/pprof/cmdline", hpprof.Cmdline)
		mux.HandleFunc("GET /debug/pprof/profile", hpprof.Profile)
		mux.HandleFunc("GET /debug/pprof/symbol", hpprof.Symbol)
		mux.HandleFunc("GET /debug/pprof/trace", hpprof.Trace)
	}
	var h http.Handler = mux
	h = s.recoverPanic(h)
	return h
}

// ServeHTTP handles incoming HTTP requests.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.h.ServeHTTP(w, r)
}
