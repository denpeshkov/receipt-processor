package server

import (
	"fmt"
	"net/http"
)

type serverError struct {
	status int
	err    error
}

func (s *serverError) Error() string {
	return fmt.Sprintf("%d (%s): %v", s.status, http.StatusText(s.status), s.err)
}

// errorHandler converts a handler that returns an error into an [http.HandlerFunc].
func (s *Server) errorHandler(h func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			s.serveError(w, r, err)
		}
	}
}

func (s *Server) serveError(w http.ResponseWriter, r *http.Request, err error) {
	s.logger.ErrorContext(r.Context(), "Failed to process request", "url", r.URL.String(), "error", err)

	serr, ok := err.(*serverError)
	if !ok {
		serr = &serverError{status: http.StatusInternalServerError, err: err}
	}
	// Do not expose internal errors to users.
	if serr.status == http.StatusInternalServerError {
		http.Error(w, "Service encountered an error while processing your request", serr.status)
		return
	}
	http.Error(w, serr.err.Error(), serr.status)
}

func (s *Server) recoverPanic(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				s.serveError(w, r, fmt.Errorf("%v", err))
			}
		}()
		h.ServeHTTP(w, r)
	})
}
