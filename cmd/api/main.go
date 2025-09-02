package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jame-developer/aeontrac/internal/appcore"
	"go.uber.org/zap/zapcore"

	"github.com/jame-developer/aeontrac/internal/api/router"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	loggerCfg := zap.NewProductionConfig()
	loggerCfg.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	logger, err := loggerCfg.Build()
	webFileFolder := ""

	if folder, ok := os.LookupEnv("AEON_ROOT"); ok {
		webFileFolder = folder
	}
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	config, data, dataFolder, err := appcore.LoadApp()
	if err != nil {
		logger.Sugar().Errorf("error loading app: %w", err)
		return
	}

	// Setup router
	r := router.SetupRouter(logger, config, data, dataFolder, webFileFolder)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server on port :8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("listen: %s\n", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutdown signal received, shutting down server...")

	// Create context with a 5-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to gracefully shutdown the server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown:", zap.Error(err))
	}

	logger.Info("Server exiting")
}
