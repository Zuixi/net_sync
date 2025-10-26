package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Host            string `json:"host" yaml:"host"`
		Port            int    `json:"port" yaml:"port"`
		HTTPS           bool   `json:"https" yaml:"https"`
		CertFile        string `json:"cert_file" yaml:"cert_file"`
		KeyFile         string `json:"key_file" yaml:"key_file"`
		ShutdownTimeout string `json:"shutdown_timeout" yaml:"shutdown_timeout"` // duration string like "30s"
	} `json:"server" yaml:"server"`

	Storage struct {
		UploadDir        string `json:"upload_dir" yaml:"upload_dir"`
		DataDir          string `json:"data_dir" yaml:"data_dir"`
		MaxFileSize      string `json:"max_file_size" yaml:"max_file_size"`       // size string like "10GB"
		PairingTokenFile string `json:"pairing_token_file" yaml:"pairing_token_file"` // filename only
	} `json:"storage" yaml:"storage"`

	WebSocket struct {
		ReadBufferSize       int    `json:"read_buffer_size" yaml:"read_buffer_size"`
		WriteBufferSize      int    `json:"write_buffer_size" yaml:"write_buffer_size"`
		ReadTimeout          string `json:"read_timeout" yaml:"read_timeout"`   // duration string
		WriteTimeout         string `json:"write_timeout" yaml:"write_timeout"` // duration string
		PingPeriod           string `json:"ping_period" yaml:"ping_period"`     // duration string
		PingMessageInterval  string `json:"ping_message_interval" yaml:"ping_message_interval"` // duration string
		ReadLimit            int64  `json:"read_limit" yaml:"read_limit"` // bytes
		SendChannelBuffer    int    `json:"send_channel_buffer" yaml:"send_channel_buffer"`
	} `json:"websocket" yaml:"websocket"`

	Security struct {
		JWTTokenExpiry       string `json:"jwt_token_expiry" yaml:"jwt_token_expiry"` // duration string
		JWTSecret            string `json:"jwt_secret" yaml:"jwt_secret"`
		PairingToken         string `json:"pairing_token" yaml:"pairing_token"`
		JWTIssuer            string `json:"jwt_issuer" yaml:"jwt_issuer"`
		FallbackJWTSecret    string `json:"fallback_jwt_secret" yaml:"fallback_jwt_secret"`
		FallbackPairingToken string `json:"fallback_pairing_token" yaml:"fallback_pairing_token"`
	} `json:"security" yaml:"security"`

	MDNS struct {
		Enabled     bool   `json:"enabled" yaml:"enabled"`
		ServiceName string `json:"service_name" yaml:"service_name"`
		DeviceName  string `json:"device_name" yaml:"device_name"`
	} `json:"mdns" yaml:"mdns"`

	TUS struct {
		BasePath   string `json:"base_path" yaml:"base_path"`
		TempSuffix string `json:"temp_suffix" yaml:"temp_suffix"`
		MetaSuffix string `json:"meta_suffix" yaml:"meta_suffix"`
	} `json:"tus" yaml:"tus"`

	CORS struct {
		AllowedOrigins []string `json:"allowed_origins" yaml:"allowed_origins"`
		AllowedMethods []string `json:"allowed_methods" yaml:"allowed_methods"`
		AllowedHeaders []string `json:"allowed_headers" yaml:"allowed_headers"`
	} `json:"cors" yaml:"cors"`

	Logging struct {
		Level  string `json:"level" yaml:"level"`
		Format string `json:"format" yaml:"format"`
	} `json:"logging" yaml:"logging"`
}

func DefaultConfig() *Config {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "easy-sync"
	}

	homeDir, _ := os.UserHomeDir()

	cfg := &Config{}

	// Server defaults
	cfg.Server.Host = "0.0.0.0"
	cfg.Server.Port = 3280
	cfg.Server.HTTPS = false
	cfg.Server.CertFile = ""
	cfg.Server.KeyFile = ""
	cfg.Server.ShutdownTimeout = "30s"

	// Storage defaults
	cfg.Storage.UploadDir = filepath.Join(homeDir, "EasySync", "Uploads")
	cfg.Storage.DataDir = filepath.Join(homeDir, "EasySync", "Data")
	cfg.Storage.MaxFileSize = "10GB"
	cfg.Storage.PairingTokenFile = "pairing-token.txt"

	// WebSocket defaults
	cfg.WebSocket.ReadBufferSize = 1024
	cfg.WebSocket.WriteBufferSize = 1024
	cfg.WebSocket.ReadTimeout = "60s"
	cfg.WebSocket.WriteTimeout = "10s"
	cfg.WebSocket.PingPeriod = "30s"
	cfg.WebSocket.PingMessageInterval = "54s"
	cfg.WebSocket.ReadLimit = 512
	cfg.WebSocket.SendChannelBuffer = 256

	// Security defaults
	cfg.Security.JWTTokenExpiry = "1440m" // 24 hours
	cfg.Security.JWTSecret = ""           // Will be auto-generated
	cfg.Security.PairingToken = ""        // Will be auto-generated
	cfg.Security.JWTIssuer = "easy-sync"
	cfg.Security.FallbackJWTSecret = "fallback-secret-change-in-production"
	cfg.Security.FallbackPairingToken = "fallback-token"

	// mDNS defaults
	cfg.MDNS.Enabled = true
	cfg.MDNS.ServiceName = "_lanxfer._tcp"
	cfg.MDNS.DeviceName = hostname

	// TUS defaults
	cfg.TUS.BasePath = "/tus/files"
	cfg.TUS.TempSuffix = ".part"
	cfg.TUS.MetaSuffix = ".meta"

	// CORS defaults
	cfg.CORS.AllowedOrigins = []string{"*"}
	cfg.CORS.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	cfg.CORS.AllowedHeaders = []string{
		"Origin",
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-CSRF-Token",
		"Authorization",
	}

	// Logging defaults
	cfg.Logging.Level = "info"
	cfg.Logging.Format = "json"

	return cfg
}

// LoadFromFile loads configuration from a YAML file
func LoadFromFile(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Expand home directory paths
	config.Storage.UploadDir = expandPath(config.Storage.UploadDir)
	config.Storage.DataDir = expandPath(config.Storage.DataDir)

	return config, nil
}

// LoadFromEnv overrides config values with environment variables
func LoadFromEnv(config *Config) {
	// Server
	if v := os.Getenv("EASYSYNC_SERVER_HOST"); v != "" {
		config.Server.Host = v
	}
	if v := os.Getenv("EASYSYNC_SERVER_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			config.Server.Port = p
		}
	}
	if v := os.Getenv("EASYSYNC_SERVER_HTTPS"); v != "" {
		config.Server.HTTPS = v == "true"
	}
	if v := os.Getenv("EASYSYNC_SERVER_CERT_FILE"); v != "" {
		config.Server.CertFile = v
	}
	if v := os.Getenv("EASYSYNC_SERVER_KEY_FILE"); v != "" {
		config.Server.KeyFile = v
	}
	if v := os.Getenv("EASYSYNC_SERVER_SHUTDOWN_TIMEOUT"); v != "" {
		config.Server.ShutdownTimeout = v
	}

	// Storage
	if v := os.Getenv("EASYSYNC_STORAGE_UPLOAD_DIR"); v != "" {
		config.Storage.UploadDir = expandPath(v)
	}
	if v := os.Getenv("EASYSYNC_STORAGE_DATA_DIR"); v != "" {
		config.Storage.DataDir = expandPath(v)
	}
	if v := os.Getenv("EASYSYNC_STORAGE_MAX_FILE_SIZE"); v != "" {
		config.Storage.MaxFileSize = v
	}
	if v := os.Getenv("EASYSYNC_STORAGE_PAIRING_TOKEN_FILE"); v != "" {
		config.Storage.PairingTokenFile = v
	}

	// WebSocket
	if v := os.Getenv("EASYSYNC_WEBSOCKET_READ_BUFFER_SIZE"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			config.WebSocket.ReadBufferSize = i
		}
	}
	if v := os.Getenv("EASYSYNC_WEBSOCKET_WRITE_BUFFER_SIZE"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			config.WebSocket.WriteBufferSize = i
		}
	}
	if v := os.Getenv("EASYSYNC_WEBSOCKET_READ_TIMEOUT"); v != "" {
		config.WebSocket.ReadTimeout = v
	}
	if v := os.Getenv("EASYSYNC_WEBSOCKET_WRITE_TIMEOUT"); v != "" {
		config.WebSocket.WriteTimeout = v
	}
	if v := os.Getenv("EASYSYNC_WEBSOCKET_PING_PERIOD"); v != "" {
		config.WebSocket.PingPeriod = v
	}
	if v := os.Getenv("EASYSYNC_WEBSOCKET_PING_MESSAGE_INTERVAL"); v != "" {
		config.WebSocket.PingMessageInterval = v
	}
	if v := os.Getenv("EASYSYNC_WEBSOCKET_READ_LIMIT"); v != "" {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			config.WebSocket.ReadLimit = i
		}
	}
	if v := os.Getenv("EASYSYNC_WEBSOCKET_SEND_CHANNEL_BUFFER"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			config.WebSocket.SendChannelBuffer = i
		}
	}

	// Security
	if v := os.Getenv("EASYSYNC_SECURITY_JWT_TOKEN_EXPIRY"); v != "" {
		config.Security.JWTTokenExpiry = v
	}
	if v := os.Getenv("EASYSYNC_SECURITY_JWT_SECRET"); v != "" {
		config.Security.JWTSecret = v
	}
	if v := os.Getenv("EASYSYNC_SECURITY_PAIRING_TOKEN"); v != "" {
		config.Security.PairingToken = v
	}
	if v := os.Getenv("EASYSYNC_SECURITY_JWT_ISSUER"); v != "" {
		config.Security.JWTIssuer = v
	}

	// mDNS
	if v := os.Getenv("EASYSYNC_MDNS_ENABLED"); v != "" {
		config.MDNS.Enabled = v == "true"
	}
	if v := os.Getenv("EASYSYNC_MDNS_SERVICE_NAME"); v != "" {
		config.MDNS.ServiceName = v
	}
	if v := os.Getenv("EASYSYNC_MDNS_DEVICE_NAME"); v != "" {
		config.MDNS.DeviceName = v
	}

	// TUS
	if v := os.Getenv("EASYSYNC_TUS_BASE_PATH"); v != "" {
		config.TUS.BasePath = v
	}
	if v := os.Getenv("EASYSYNC_TUS_TEMP_SUFFIX"); v != "" {
		config.TUS.TempSuffix = v
	}
	if v := os.Getenv("EASYSYNC_TUS_META_SUFFIX"); v != "" {
		config.TUS.MetaSuffix = v
	}

	// Logging
	if v := os.Getenv("EASYSYNC_LOGGING_LEVEL"); v != "" {
		config.Logging.Level = v
	}
	if v := os.Getenv("EASYSYNC_LOGGING_FORMAT"); v != "" {
		config.Logging.Format = v
	}
}

// Load loads configuration with priority: config file > env vars > defaults
func Load(configPath string) (*Config, error) {
	var config *Config
	var err error

	// Try to load from file if path is provided
	if configPath != "" {
		config, err = LoadFromFile(configPath)
		if err != nil {
			return nil, err
		}
	} else {
		// Use default config if no file specified
		config = DefaultConfig()
	}

	// Override with environment variables
	LoadFromEnv(config)

	return config, nil
}

// expandPath expands ~ to user home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[1:])
		}
	}
	return path
}

func (c *Config) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

func (c *Config) EnsureDirectories() error {
	if err := os.MkdirAll(c.Storage.UploadDir, 0755); err != nil {
		return fmt.Errorf("failed to create upload directory: %w", err)
	}

	if err := os.MkdirAll(c.Storage.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	return nil
}

// WritePairingInfo writes pairing information to a file for easy user access
func (c *Config) WritePairingInfo(serverAddr string) error {
	tokenFile := filepath.Join(c.Storage.DataDir, c.Storage.PairingTokenFile)

	protocol := "http"
	if c.Server.HTTPS {
		protocol = "https"
	}

	content := fmt.Sprintf(`========================================
Easy Sync - 配对信息
========================================
服务器地址: %s://%s
配对令牌: %s

使用说明:
  - 自动配对: 在同一网络打开网页自动连接
  - 手动配对: 在网页中输入上述令牌

生成时间: %s
Token文件位置: %s
========================================
`,
		protocol,
		serverAddr,
		c.Security.PairingToken,
		time.Now().Format("2006-01-02 15:04:05"),
		tokenFile,
	)

	if err := os.WriteFile(tokenFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write pairing token file: %w", err)
	}

	return nil
}

// ParseSize parses a size string (e.g., "10GB") to bytes
func ParseSize(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return 0, fmt.Errorf("empty size string")
	}

	multipliers := map[string]int64{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
	}

	for suffix, multiplier := range multipliers {
		if strings.HasSuffix(s, suffix) {
			numStr := strings.TrimSuffix(s, suffix)
			numStr = strings.TrimSpace(numStr)
			num, err := strconv.ParseInt(numStr, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid size format: %w", err)
			}
			return num * multiplier, nil
		}
	}

	// If no suffix, assume bytes
	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %w", err)
	}
	return num, nil
}

// ParseDuration is a helper that wraps time.ParseDuration with better error messages
func ParseDuration(s string) (time.Duration, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration format '%s': %w (examples: 30s, 5m, 1h)", s, err)
	}
	return d, nil
}

// GetMaxFileSizeBytes returns the max file size in bytes
func (c *Config) GetMaxFileSizeBytes() (int64, error) {
	return ParseSize(c.Storage.MaxFileSize)
}

// GetShutdownTimeout returns the shutdown timeout as time.Duration
func (c *Config) GetShutdownTimeout() (time.Duration, error) {
	return ParseDuration(c.Server.ShutdownTimeout)
}

// GetJWTTokenExpiry returns the JWT token expiry as time.Duration
func (c *Config) GetJWTTokenExpiry() (time.Duration, error) {
	return ParseDuration(c.Security.JWTTokenExpiry)
}

// GetWebSocketReadTimeout returns the WebSocket read timeout as time.Duration
func (c *Config) GetWebSocketReadTimeout() (time.Duration, error) {
	return ParseDuration(c.WebSocket.ReadTimeout)
}

// GetWebSocketWriteTimeout returns the WebSocket write timeout as time.Duration
func (c *Config) GetWebSocketWriteTimeout() (time.Duration, error) {
	return ParseDuration(c.WebSocket.WriteTimeout)
}

// GetWebSocketPingPeriod returns the WebSocket ping period as time.Duration
func (c *Config) GetWebSocketPingPeriod() (time.Duration, error) {
	return ParseDuration(c.WebSocket.PingPeriod)
}

// GetWebSocketPingMessageInterval returns the WebSocket ping message interval as time.Duration
func (c *Config) GetWebSocketPingMessageInterval() (time.Duration, error) {
	return ParseDuration(c.WebSocket.PingMessageInterval)
}