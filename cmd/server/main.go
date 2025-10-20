package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/easy-sync/easy-sync/pkg/config"
	"github.com/easy-sync/easy-sync/pkg/server"
	"github.com/easy-sync/easy-sync/pkg/discovery"
)

func main() {
	// Initialize logger
	logger := logrus.New()

	// Try to load log level from environment
	logLevel := os.Getenv("EASY_SYNC_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set log format
	logFormat := os.Getenv("EASY_SYNC_LOG_FORMAT")
	if logFormat == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	logger.Info("Starting Easy Sync Server")

	// Load configuration
	cfg := config.DefaultConfig()
	config.LoadFromEnv(cfg)

	// Ensure directories exist
	if err := cfg.EnsureDirectories(); err != nil {
		logger.WithError(err).Fatal("Failed to create required directories")
	}

	// Print configuration
	logger.WithFields(logrus.Fields{
		"host":        cfg.Server.Host,
		"port":        cfg.Server.Port,
		"https":       cfg.Server.HTTPS,
		"upload_dir":  cfg.Storage.UploadDir,
		"data_dir":    cfg.Storage.DataDir,
		"device_name": cfg.Discovery.DeviceName,
		"discovery":   cfg.Discovery.Enabled,
	}).Info("Configuration loaded")

	// Print pairing information
	logger.WithFields(logrus.Fields{
		"pairing_token": cfg.Security.PairingToken,
		"address":       cfg.GetAddr(),
	}).Info("Pairing information - save this token for device pairing")

	// Initialize mDNS discovery
	var mdns *discovery.MDnsDiscovery
	if cfg.Discovery.Enabled {
		mdns = discovery.NewMDnsDiscovery(cfg, logger)
		if err := mdns.Start(); err != nil {
			logger.WithError(err).Error("Failed to start mDNS discovery")
			mdns = nil
		} else {
			logger.Info("mDNS discovery started")
		}
	}

	// Create server
	srv, err := server.NewServer(cfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create server")
	}

	// Start WebSocket manager
	srv.StartWebSocketManager()

	// Channel to capture OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		if err := srv.Start(); err != nil {
			logger.WithError(err).Error("Server failed to start")
			sigChan <- syscall.SIGTERM
		}
	}()

	// Wait for interrupt signal
	sig := <-sigChan
	logger.WithField("signal", sig).Info("Received shutdown signal")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown gracefully
	logger.Info("Shutting down server...")

	// Stop mDNS discovery
	if mdns != nil {
		if err := mdns.Stop(); err != nil {
			logger.WithError(err).Error("Failed to stop mDNS discovery")
		}
	}

	// Stop server
	if err := srv.Stop(ctx); err != nil {
		logger.WithError(err).Error("Failed to stop server gracefully")
	}

	logger.Info("Server stopped")
}
