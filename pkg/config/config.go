package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Config struct {
	Server struct {
		Host     string `json:"host" yaml:"host"`
		Port     int    `json:"port" yaml:"port"`
		HTTPS    bool   `json:"https" yaml:"https"`
		CertFile string `json:"cert_file" yaml:"cert_file"`
		KeyFile  string `json:"key_file" yaml:"key_file"`
	} `json:"server" yaml:"server"`

	Storage struct {
		UploadDir string `json:"upload_dir" yaml:"upload_dir"`
		DataDir   string `json:"data_dir" yaml:"data_dir"`
		MaxSize   int64  `json:"max_size" yaml:"max_size"` // in bytes
	} `json:"storage" yaml:"storage"`

	Discovery struct {
		Enabled     bool   `json:"enabled" yaml:"enabled"`
		ServiceName string `json:"service_name" yaml:"service_name"`
		DeviceName  string `json:"device_name" yaml:"device_name"`
	} `json:"discovery" yaml:"discovery"`

	Security struct {
		JWTSecret    string `json:"jwt_secret" yaml:"jwt_secret"`
		PairingToken string `json:"pairing_token" yaml:"pairing_token"`
		TokenExpiry  int    `json:"token_expiry" yaml:"token_expiry"` // in minutes
	} `json:"security" yaml:"security"`

	WebSocket struct {
		ReadBufferSize  int `json:"read_buffer_size" yaml:"read_buffer_size"`
		WriteBufferSize int `json:"write_buffer_size" yaml:"write_buffer_size"`
		PingPeriod      int `json:"ping_period" yaml:"ping_period"` // in seconds
	} `json:"websocket" yaml:"websocket"`

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

	return &Config{
		Server: struct {
			Host     string `json:"host" yaml:"host"`
			Port     int    `json:"port" yaml:"port"`
			HTTPS    bool   `json:"https" yaml:"https"`
			CertFile string `json:"cert_file" yaml:"cert_file"`
			KeyFile  string `json:"key_file" yaml:"key_file"`
		}{
			Host:     "0.0.0.0",
			Port:     3280,
			HTTPS:    false,
			CertFile: "",
			KeyFile:  "",
		},
		Storage: struct {
			UploadDir string `json:"upload_dir" yaml:"upload_dir"`
			DataDir   string `json:"data_dir" yaml:"data_dir"`
			MaxSize   int64  `json:"max_size" yaml:"max_size"`
		}{
			UploadDir: filepath.Join(homeDir, "EasySync", "Uploads"),
			DataDir:   filepath.Join(homeDir, "EasySync", "Data"),
			MaxSize:   10 * 1024 * 1024 * 1024, // 10GB
		},
		Discovery: struct {
			Enabled     bool   `json:"enabled" yaml:"enabled"`
			ServiceName string `json:"service_name" yaml:"service_name"`
			DeviceName  string `json:"device_name" yaml:"device_name"`
		}{
			Enabled:     true,
			ServiceName: "_lanxfer._tcp",
			DeviceName:  hostname,
		},
		Security: struct {
			JWTSecret    string `json:"jwt_secret" yaml:"jwt_secret"`
			PairingToken string `json:"pairing_token" yaml:"pairing_token"`
			TokenExpiry  int    `json:"token_expiry" yaml:"token_expiry"`
		}{
			JWTSecret:    "", // Will be auto-generated
			PairingToken: "", // Will be auto-generated
			TokenExpiry:  1440, // 24 hours
		},
		WebSocket: struct {
			ReadBufferSize  int `json:"read_buffer_size" yaml:"read_buffer_size"`
			WriteBufferSize int `json:"write_buffer_size" yaml:"write_buffer_size"`
			PingPeriod      int `json:"ping_period" yaml:"ping_period"`
		}{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			PingPeriod:      30,
		},
		Logging: struct {
			Level  string `json:"level" yaml:"level"`
			Format string `json:"format" yaml:"format"`
		}{
			Level:  "info",
			Format: "json",
		},
	}
}

func LoadFromEnv(config *Config) {
	if host := os.Getenv("EASY_SYNC_HOST"); host != "" {
		config.Server.Host = host
	}

	if port := os.Getenv("EASY_SYNC_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Server.Port = p
		}
	}

	if https := os.Getenv("EASY_SYNC_HTTPS"); https != "" {
		config.Server.HTTPS = https == "true"
	}

	if certFile := os.Getenv("EASY_SYNC_CERT_FILE"); certFile != "" {
		config.Server.CertFile = certFile
	}

	if keyFile := os.Getenv("EASY_SYNC_KEY_FILE"); keyFile != "" {
		config.Server.KeyFile = keyFile
	}

	if uploadDir := os.Getenv("EASY_SYNC_UPLOAD_DIR"); uploadDir != "" {
		config.Storage.UploadDir = uploadDir
	}

	if dataDir := os.Getenv("EASY_SYNC_DATA_DIR"); dataDir != "" {
		config.Storage.DataDir = dataDir
	}

	if maxSize := os.Getenv("EASY_SYNC_MAX_SIZE"); maxSize != "" {
		if size, err := strconv.ParseInt(maxSize, 10, 64); err == nil {
			config.Storage.MaxSize = size
		}
	}

	if deviceName := os.Getenv("EASY_SYNC_DEVICE_NAME"); deviceName != "" {
		config.Discovery.DeviceName = deviceName
	}

	if jwtSecret := os.Getenv("EASY_SYNC_JWT_SECRET"); jwtSecret != "" {
		config.Security.JWTSecret = jwtSecret
	}

	if pairingToken := os.Getenv("EASY_SYNC_PAIRING_TOKEN"); pairingToken != "" {
		config.Security.PairingToken = pairingToken
	}

	if logLevel := os.Getenv("EASY_SYNC_LOG_LEVEL"); logLevel != "" {
		config.Logging.Level = logLevel
	}
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
	tokenFile := filepath.Join(c.Storage.DataDir, "pairing-token.txt")

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