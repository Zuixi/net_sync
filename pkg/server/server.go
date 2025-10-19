package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/easy-sync/easy-sync/pkg/config"
	"github.com/easy-sync/easy-sync/pkg/websocket"
	"github.com/easy-sync/easy-sync/pkg/upload"
	"github.com/easy-sync/easy-sync/pkg/download"
	"github.com/easy-sync/easy-sync/pkg/security"
)

type Server struct {
	config       *config.Config
	httpServer   *http.Server
	router       *gin.Engine
	logger       *logrus.Logger
	wsManager    *websocket.Manager
	tusHandler   *upload.TusHandler
	downloadHandler *download.Handler
	auth         *security.AuthService
}

func NewServer(cfg *config.Config, logger *logrus.Logger) (*Server, error) {
	if cfg.Server.HTTPS {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Logging middleware
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		logger.WithFields(logrus.Fields{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"duration":   duration,
			"client_ip":  c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		}).Info("HTTP Request")
	})

	auth := security.NewAuthService(cfg, logger)
	wsManager := websocket.NewManager(cfg, logger, auth)

	tusHandler, err := upload.NewTusHandler(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUS handler: %w", err)
	}

	downloadHandler := download.NewHandler(cfg, logger)

	server := &Server{
		config:          cfg,
		router:          router,
		logger:          logger,
		wsManager:       wsManager,
		tusHandler:      tusHandler,
		downloadHandler: downloadHandler,
		auth:           auth,
	}

	server.setupRoutes()

	return server, nil
}

func (s *Server) setupRoutes() {
	// Health check
	s.router.GET("/health", s.healthCheck)

	// API routes
	api := s.router.Group("/api")
	{
		// Device pairing and discovery
		api.GET("/qr", s.auth.RequireAuth(), s.generateQR)
		api.POST("/pair", s.pairDevice)
		api.GET("/devices", s.auth.RequireAuth(), s.listDevices)
		api.DELETE("/devices/:id", s.auth.RequireAuth(), s.removeDevice)

		// File management
		api.GET("/files", s.auth.RequireAuth(), s.listFiles)
		api.DELETE("/files/:id", s.auth.RequireAuth(), s.deleteFile)

		// Messages
		api.GET("/messages", s.auth.RequireAuth(), s.getMessages)
	}

	// WebSocket endpoint
	s.router.GET("/ws", s.wsManager.HandleWebSocket)

	// TUS file upload endpoints
	s.router.Any("/tus/*filepath", s.tusHandler.HandleRequest)

	// File download endpoint
	s.router.GET("/files/:id", s.downloadHandler.HandleDownload)
	s.router.GET("/files/:id/sha256", s.downloadHandler.HandleSHA256)

	// Static files (web UI)
	s.router.Static("/static", "./web/public")
	s.router.StaticFile("/", "./web/public/index.html")
}

func (s *Server) Start() error {
	addr := s.config.GetAddr()

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	if s.config.Server.HTTPS {
		if s.config.Server.CertFile == "" || s.config.Server.KeyFile == "" {
			return fmt.Errorf("HTTPS enabled but cert file or key file not specified")
		}

		s.httpServer.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}

		s.logger.WithFields(logrus.Fields{
			"address":   addr,
			"cert_file": s.config.Server.CertFile,
			"key_file":  s.config.Server.KeyFile,
		}).Info("Starting HTTPS server")

		return s.httpServer.ListenAndServeTLS(s.config.Server.CertFile, s.config.Server.KeyFile)
	}

	s.logger.WithField("address", addr).Info("Starting HTTP server")
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Shutting down server...")

	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}

	return nil
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	})
}

func (s *Server) generateQR(c *gin.Context) {
	// TODO: Implement QR code generation
	qrData := map[string]interface{}{
		"url":   fmt.Sprintf("http://%s", s.config.GetAddr()),
		"token": s.config.Security.PairingToken,
	}

	c.JSON(200, qrData)
}

func (s *Server) pairDevice(c *gin.Context) {
	var request struct {
		Token    string `json:"token" binding:"required"`
		DeviceID string `json:"device_id" binding:"required"`
		DeviceName string `json:"device_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if request.Token != s.config.Security.PairingToken {
		c.JSON(401, gin.H{"error": "Invalid pairing token"})
		return
	}

	token, err := s.auth.GenerateDeviceToken(request.DeviceID, request.DeviceName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(200, gin.H{
		"token": token,
		"expires_in": s.config.Security.TokenExpiry * 60, // convert to seconds
	})
}

func (s *Server) listDevices(c *gin.Context) {
	// TODO: Implement device listing
	c.JSON(200, gin.H{"devices": []interface{}{}})
}

func (s *Server) removeDevice(c *gin.Context) {
	deviceID := c.Param("id")
	// TODO: Implement device removal
	c.JSON(200, gin.H{"message": fmt.Sprintf("Device %s removed", deviceID)})
}

func (s *Server) listFiles(c *gin.Context) {
	// TODO: Implement file listing
	c.JSON(200, gin.H{"files": []interface{}{}})
}

func (s *Server) deleteFile(c *gin.Context) {
	fileID := c.Param("id")
	// TODO: Implement file deletion
	c.JSON(200, gin.H{"message": fmt.Sprintf("File %s deleted", fileID)})
}

func (s *Server) getMessages(c *gin.Context) {
	// TODO: Implement message retrieval
	c.JSON(200, gin.H{"messages": []interface{}{}})
}