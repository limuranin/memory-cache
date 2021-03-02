package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"memory-cache/cache"
	"memory-cache/config"
	"memory-cache/logger"
	"memory-cache/server"
)

const ShutdownServerTimeout = 10 * time.Second

func main() {
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "-h" || arg == "--help" {
			if err := config.PrintHelp(); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Call help error: %s\n", err)
				os.Exit(1)
			}

			os.Exit(0)
		}
	}

	cfg, err := config.Init()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Configuration init error: %s\n", err)
		os.Exit(1)
	}

	if err := logger.Init(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Logger init error: %s\n", err)
		os.Exit(1)
	}

	logger.Infof("Start cache with cleaning interval: %v", cfg.Cache.CleaningInterval)
	cacheCtx, cacheCancelFunc := context.WithCancel(context.Background())
	defer cacheCancelFunc()
	cacheStorage := cache.NewCache(cacheCtx, cfg.Cache)
	cacheStorage.Start()

	logger.Infof("Start server listen address: %v", cfg.Server.ListenAddress)
	srv := server.NewServer(cfg.Server, cacheStorage)
	if err := srv.Start(); err != nil {
		logger.Errorf("Start server error: %v", err)
		os.Exit(1)
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)
	<-shutdownChan

	logger.Info("Stopping cache")
	cacheCancelFunc()

	logger.Infof("Shutting down the server, wait gracefully shutdown for %v", ShutdownServerTimeout)
	shutdownServerCtx, shutdownServerCancelFunc := context.WithTimeout(context.Background(), ShutdownServerTimeout)
	defer shutdownServerCancelFunc()

	if err := srv.Shutdown(shutdownServerCtx); err != nil {
		logger.Error("Can't shutdown server gracefully")
		os.Exit(1)
	}
	logger.Info("Server shutdown gracefully")

	os.Exit(0)
}
