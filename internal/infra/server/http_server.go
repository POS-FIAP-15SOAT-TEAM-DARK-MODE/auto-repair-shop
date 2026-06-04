package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"go.uber.org/zap"
)

type httpServer struct {
	server *http.Server
	name   string
}

// NewHTTPServer creates a new HTTP server implementation
func NewHTTPServer(name string, addr string, handler http.Handler) Server {
	return &httpServer{
		name: name,
		server: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}
}

func (s *httpServer) Start(_ context.Context) error {
	logger.Global().Info("starting server", zap.String("name", s.name), zap.String("addr", s.server.Addr))
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *httpServer) Stop(ctx context.Context) error {
	logger.Global().Info("stopping server", zap.String("name", s.name))
	return s.server.Shutdown(ctx)
}

func (s *httpServer) Name() string {
	return s.name
}
