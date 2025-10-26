package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/easy-sync/easy-sync/pkg/config"
	"github.com/easy-sync/easy-sync/pkg/download"
	"github.com/easy-sync/easy-sync/pkg/security"
	"github.com/easy-sync/easy-sync/pkg/upload"
	"github.com/easy-sync/easy-sync/pkg/websocket"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Server struct {
	config          *config.Config
	httpServer      *http.Server
	router          *gin.Engine
	logger          *logrus.Logger
	wsManager       *websocket.Manager
	tusHandler      *upload.TusHandler
	downloadHandler *download.Handler
	auth            *security.AuthService
	finalAddr       string // Store the final bound address
}

func NewServer(cfg *config.Config, logger *logrus.Logger) (*Server, error) {
	if cfg.Server.HTTPS {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	// CORS middleware
	router.Use(func(c *gin.Context) {
		// Allow origins from config
		origin := "*"
		if len(cfg.CORS.AllowedOrigins) > 0 {
			origin = cfg.CORS.AllowedOrigins[0]
			// If multiple origins, check if request origin is in the list
			if len(cfg.CORS.AllowedOrigins) > 1 {
				requestOrigin := c.Request.Header.Get("Origin")
				for _, allowed := range cfg.CORS.AllowedOrigins {
					if allowed == requestOrigin || allowed == "*" {
						origin = allowed
						break
					}
				}
			}
		}
		c.Header("Access-Control-Allow-Origin", origin)

		// Set allowed methods from config
		if len(cfg.CORS.AllowedMethods) > 0 {
			methods := ""
			for i, method := range cfg.CORS.AllowedMethods {
				if i > 0 {
					methods += ", "
				}
				methods += method
			}
			c.Header("Access-Control-Allow-Methods", methods)
		}

		// Set allowed headers from config
		if len(cfg.CORS.AllowedHeaders) > 0 {
			headers := ""
			for i, header := range cfg.CORS.AllowedHeaders {
				if i > 0 {
					headers += ", "
				}
				headers += header
			}
			c.Header("Access-Control-Allow-Headers", headers)
		}

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
		auth:            auth,
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
		api.GET("/pairing-token", s.getPairingToken) // No auth required for auto-pairing
		api.POST("/pair", s.pairDevice)
		api.GET("/devices", s.auth.RequireAuth(), s.listDevices)
		api.DELETE("/devices/:id", s.auth.RequireAuth(), s.removeDevice)

		// File management
		api.GET("/files", s.auth.RequireAuth(), s.listFiles)
		api.DELETE("/files/:id", s.auth.RequireAuth(), s.deleteFile)

		// Messages
		api.GET("/messages", s.auth.RequireAuth(), s.getMessages)

		// Configuration endpoint
		api.GET("/config", s.getConfig)
	}

	// WebSocket endpoint
	s.router.GET("/ws", func(c *gin.Context) { s.wsManager.HandleWebSocket(c.Writer, c.Request) })

	// TUS file upload endpoints
	s.router.POST("/tus/*filepath", func(c *gin.Context) { s.tusHandler.HandleRequest(c.Writer, c.Request) })
	s.router.PATCH("/tus/*filepath", func(c *gin.Context) { s.tusHandler.HandleRequest(c.Writer, c.Request) })
	s.router.HEAD("/tus/*filepath", func(c *gin.Context) { s.tusHandler.HandleRequest(c.Writer, c.Request) })
	s.router.GET("/tus/*filepath", func(c *gin.Context) { s.tusHandler.HandleRequest(c.Writer, c.Request) })
	s.router.OPTIONS("/tus/*filepath", func(c *gin.Context) { s.tusHandler.HandleRequest(c.Writer, c.Request) })

	// File download endpoint
	s.router.GET("/files/:id", func(c *gin.Context) { s.downloadHandler.HandleDownload(c.Writer, c.Request) })
	s.router.GET("/files/:id/sha256", func(c *gin.Context) { s.downloadHandler.HandleSHA256(c.Writer, c.Request) })

	// Static files (web UI)
	s.router.Static("/static", "./web/public")
	s.router.StaticFile("/", "./web/public/index.html")
}

func (s *Server) Start() error {
	// Prepare HTTP server
	addr := s.config.GetAddr()
	_ = addr // not used directly; we use host/port from config

	host := s.config.Server.Host
	port := s.config.Server.Port

	// Candidate hosts: try configured host, then 127.0.0.1 for Windows policies
	hosts := []string{host}
	if host == "0.0.0.0" {
		hosts = []string{host, "127.0.0.1"}
	}

	// Candidate ports: configured port then a small range fallback
	candidatePorts := []int{port, port + 1, port + 2}

	var listener net.Listener
	var listenErr error

outer:
	for _, h := range hosts {
		for _, p := range candidatePorts {
			address := fmt.Sprintf("%s:%d", h, p)
			ln, err := net.Listen("tcp", address)
			if err != nil {
				s.logger.WithFields(logrus.Fields{"address": address}).WithError(err).Warn("Address unavailable, trying next")
				listenErr = err
				continue
			}
			listener = ln
			break outer
		}
	}

	if listener == nil {
		// Ephemeral port fallback on last host
		lastHost := hosts[len(hosts)-1]
		ephemeralAddr := fmt.Sprintf("%s:%d", lastHost, 0)
		ln, err := net.Listen("tcp", ephemeralAddr)
		if err != nil {
			s.logger.WithError(err).Error("Failed to bind any port")
			if listenErr != nil {
				return listenErr
			}
			return err
		}
		listener = ln
	}

	finalAddr := listener.Addr().String()
	s.finalAddr = finalAddr // Store for API access

	s.httpServer = &http.Server{
		Addr:    finalAddr,
		Handler: s.router,
	}

	// Write pairing information to file with actual bound address
	if err := s.config.WritePairingInfo(finalAddr); err != nil {
		s.logger.WithError(err).Warn("Failed to write pairing token file")
	} else {
		s.logger.Info("Pairing token written to file")
	}

	// Print final address to console
	fmt.Printf("\n🚀 Easy-Sync Server is running at: %s\n", finalAddr)
	if s.config.Server.HTTPS {
		fmt.Printf("   HTTPS URL: https://%s\n", finalAddr)
	} else {
		fmt.Printf("   HTTP URL: http://%s\n", finalAddr)
	}
	fmt.Printf("   Health Check: http://%s/health\n", finalAddr)
	fmt.Printf("   API Config: http://%s/api/config\n\n", finalAddr)

	if s.config.Server.HTTPS {
		if s.config.Server.CertFile == "" || s.config.Server.KeyFile == "" {
			return fmt.Errorf("HTTPS enabled but cert file or key file not specified")
		}

		// Minimal TLS config; http.Server.ServeTLS will wrap listener
		s.httpServer.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}

		s.logger.WithFields(logrus.Fields{
			"address":   finalAddr,
			"cert_file": s.config.Server.CertFile,
			"key_file":  s.config.Server.KeyFile,
		}).Info("Starting HTTPS server")

		return s.httpServer.ServeTLS(listener, s.config.Server.CertFile, s.config.Server.KeyFile)
	}

	s.logger.WithField("address", finalAddr).Info("Starting HTTP server")
	return s.httpServer.Serve(listener)
}

func (s *Server) StartWebSocketManager() {
	if s.wsManager != nil {
		s.wsManager.Start()
	}
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

func (s *Server) getPairingToken(c *gin.Context) {
	protocol := "http"
	if s.config.Server.HTTPS {
		protocol = "https"
	}

	c.JSON(200, gin.H{
		"token":      s.config.Security.PairingToken,
		"server_url": fmt.Sprintf("%s://%s", protocol, s.finalAddr),
		"timestamp":  time.Now().Unix(),
	})
}

func (s *Server) pairDevice(c *gin.Context) {
	var request struct {
		Token      string `json:"token" binding:"required"`
		DeviceID   string `json:"device_id" binding:"required"`
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

	// Create device record
	_, err := s.auth.CreateDevice(request.DeviceID, request.DeviceName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create device"})
		return
	}

	token, err := s.auth.GenerateDeviceToken(request.DeviceID, request.DeviceName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	tokenExpiry, err := s.config.GetJWTTokenExpiry()
	if err != nil {
		s.logger.WithError(err).Warn("Invalid JWT token expiry, using default 24h")
		tokenExpiry = 24 * time.Hour
	}

	c.JSON(200, gin.H{
		"token":      token,
		"expires_in": int(tokenExpiry.Seconds()), // convert to seconds
	})
}

func (s *Server) listDevices(c *gin.Context) {
	// Return currently connected devices from WebSocket manager
	devices := s.wsManager.GetConnectedDevices()
	c.JSON(200, gin.H{"devices": devices})
}

func (s *Server) removeDevice(c *gin.Context) {
	deviceID := c.Param("id")

	err := s.auth.RemoveDevice(deviceID)
	if err != nil {
		if err.Error() == "device not found" {
			c.JSON(404, gin.H{"error": "Device not found"})
		} else {
			c.JSON(500, gin.H{"error": "Failed to remove device"})
		}
		return
	}

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

func (s *Server) getConfig(c *gin.Context) {
	protocol := "http"
	if s.config.Server.HTTPS {
		protocol = "https"
	}

	configData := gin.H{
		"server": gin.H{
			"address":         s.finalAddr,
			"protocol":        protocol,
			"url":             fmt.Sprintf("%s://%s", protocol, s.finalAddr),
			"configured_host": s.config.Server.Host,
			"configured_port": s.config.Server.Port,
			"https_enabled":   s.config.Server.HTTPS,
		},
		"endpoints": gin.H{
			"health":    fmt.Sprintf("%s://%s/health", protocol, s.finalAddr),
			"websocket": fmt.Sprintf("ws://%s/ws", s.finalAddr),
			"upload":    fmt.Sprintf("%s://%s/tus/", protocol, s.finalAddr),
			"api_base":  fmt.Sprintf("%s://%s/api", protocol, s.finalAddr),
		},
		"version":   "1.0.0",
		"timestamp": time.Now().Unix(),
	}

	c.JSON(200, configData)
}
