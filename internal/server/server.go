package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	hpprof "net/http/pprof"

	"github.com/denpeshkov/receipt-processor/internal/receipt"
)

type Storage interface {
	Store(ctx context.Context, receipt *receipt.Receipt) error
	Get(ctx context.Context, id string) (*receipt.Receipt, error)
}

type Options struct {
	Logger    *slog.Logger
	DebugMode bool
}

type Server struct {
	storage Storage
	logger  *slog.Logger
	h       http.Handler
}

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

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.h.ServeHTTP(w, r)
}
