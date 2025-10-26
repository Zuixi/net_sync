package main

import (
	"context"
	"flag"
	"fmt"
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
	// Parse command-line flags
	configPath := flag.String("config", "", "Path to configuration file (YAML)")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Show version if requested
	if *showVersion {
		fmt.Println("EasySync Server v1.0.0")
		os.Exit(0)
	}

	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("Starting Easy Sync Server")

	// Load configuration with priority: config file > env vars > defaults
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	// Log the config source
	if *configPath != "" {
		logger.WithField("config_file", *configPath).Info("Loaded configuration from file")
	} else {
		logger.Info("Using default configuration")
	}

	// Configure logger from config
	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		logger.WithError(err).Warn("Invalid log level, using info")
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	if cfg.Logging.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

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
		"device_name": cfg.MDNS.DeviceName,
		"discovery":   cfg.MDNS.Enabled,
	}).Info("Configuration loaded")

	// Create server first to generate the pairing token
	srv, err := server.NewServer(cfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create server")
	}

	// Print pairing information after server creation
	logger.WithFields(logrus.Fields{
		"pairing_token": cfg.Security.PairingToken,
		"address":       cfg.GetAddr(),
	}).Info("Pairing information - save this token for device pairing")

	// Initialize mDNS discovery
	var mdns *discovery.MDnsDiscovery
	if cfg.MDNS.Enabled {
		mdns = discovery.NewMDnsDiscovery(cfg, logger)
		if err := mdns.Start(); err != nil {
			logger.WithError(err).Error("Failed to start mDNS discovery")
			mdns = nil
		} else {
			logger.Info("mDNS discovery started")
		}
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

	// Get shutdown timeout from config
	shutdownTimeout, err := cfg.GetShutdownTimeout()
	if err != nil {
		logger.WithError(err).Warn("Invalid shutdown timeout, using default 30s")
		shutdownTimeout = 30 * time.Second
	}

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
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
