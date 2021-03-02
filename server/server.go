package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"memory-cache/config"
	"memory-cache/logger"

	"github.com/gorilla/mux"
)

const (
	ReadTimeout  = 10 * time.Second
	WriteTimeout = 10 * time.Second
)

type Server struct {
	cfg        *config.ServerCfg
	httpServer *http.Server
	cacher     Cacher
}

func NewServer(cfg *config.ServerCfg, cacher Cacher) *Server {
	return &Server{
		cfg:        cfg,
		cacher:     cacher,
		httpServer: nil,
	}
}

func (s *Server) Start() error {
	router := mux.NewRouter()
	routesHandler := newRoutesHandler(router, s.cacher)
	routesHandler.registerRoutes()

	s.httpServer = &http.Server{
		Addr:         s.cfg.ListenAddress,
		Handler:      router,
		ReadTimeout:  ReadTimeout,
		WriteTimeout: WriteTimeout,
	}

	listener, err := net.Listen("tcp", s.cfg.ListenAddress)
	if err != nil {
		return fmt.Errorf("listen on address: %v error %v", s.cfg.ListenAddress, err)
	}

	go func() {
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			logger.Panicf("Serve on address: %v error %v", s.cfg.ListenAddress, err)
		}
	}()

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
