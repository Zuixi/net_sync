package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/easy-sync/easy-sync/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

type AuthService struct {
	config  *config.Config
	logger  *logrus.Logger
	devices map[string]*Device
	mutex   sync.RWMutex
}

type Claims struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	jwt.RegisteredClaims
}

type Device struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	PublicKey string    `json:"public_key"`
	Created   time.Time `json:"created"`
	LastSeen  time.Time `json:"last_seen"`
	Trusted   bool      `json:"trusted"`
}

func NewAuthService(cfg *config.Config, logger *logrus.Logger) *AuthService {
	// Generate JWT secret if not provided
	if cfg.Security.JWTSecret == "" {
		secret, err := generateRandomToken(32)
		if err != nil {
			logger.WithError(err).Error("Failed to generate JWT secret")
			cfg.Security.JWTSecret = "fallback-secret-change-in-production"
		} else {
			cfg.Security.JWTSecret = secret
		}
	}

	// Generate pairing token if not provided
	if cfg.Security.PairingToken == "" {
		token, err := generateRandomToken(16)
		if err != nil {
			logger.WithError(err).Error("Failed to generate pairing token")
			cfg.Security.PairingToken = "fallback-token"
		} else {
			cfg.Security.PairingToken = token
		}
	}

	return &AuthService{
		config:  cfg,
		logger:  logger,
		devices: make(map[string]*Device),
	}
}

func (a *AuthService) GenerateDeviceToken(deviceID, deviceName string) (string, error) {
	claims := &Claims{
		DeviceID:   deviceID,
		DeviceName: deviceName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(a.config.Security.TokenExpiry) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "easy-sync",
			Subject:   deviceID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.config.Security.JWTSecret))
}

func (a *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	// Remove "Bearer " prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.config.Security.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (a *AuthService) ValidatePairingToken(token string) bool {
	return token == a.config.Security.PairingToken
}

func (a *AuthService) GetPairingToken() string {
	return a.config.Security.PairingToken
}

func (a *AuthService) GenerateQRData() (map[string]interface{}, error) {
	return map[string]interface{}{
		"url":       fmt.Sprintf("http://%s", a.config.GetAddr()),
		"token":     a.config.Security.PairingToken,
		"device_id": a.config.Discovery.DeviceName,
		"timestamp": time.Now().Unix(),
	}, nil
}

func (a *AuthService) CreateDevice(deviceID, deviceName string) (*Device, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	device := &Device{
		ID:       deviceID,
		Name:     deviceName,
		Created:  time.Now(),
		LastSeen: time.Now(),
		Trusted:  true, // Auto-trust devices that complete pairing
	}

	// Store device in memory map
	a.devices[deviceID] = device

	a.logger.WithFields(logrus.Fields{
		"device_id":   deviceID,
		"device_name": deviceName,
	}).Info("Device paired")

	return device, nil
}

func (a *AuthService) GetDevice(deviceID string) (*Device, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	device, exists := a.devices[deviceID]
	if !exists {
		return nil, fmt.Errorf("device not found")
	}

	// Update last seen time
	device.LastSeen = time.Now()
	return device, nil
}

func (a *AuthService) ListDevices() ([]*Device, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	// Return slice of all devices
	devices := make([]*Device, 0, len(a.devices))
	for _, device := range a.devices {
		devices = append(devices, device)
	}

	return devices, nil
}

func (a *AuthService) RemoveDevice(deviceID string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if _, exists := a.devices[deviceID]; !exists {
		return fmt.Errorf("device not found")
	}

	delete(a.devices, deviceID)
	a.logger.WithField("device_id", deviceID).Info("Device removed")
	return nil
}

func (a *AuthService) IsDeviceTrusted(deviceID string) bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	device, exists := a.devices[deviceID]
	if !exists {
		return false
	}

	return device.Trusted
}

func generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Middleware function for authentication
func (a *AuthService) RequireAuth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Try to get token from query parameter
			token := c.Query("token")
			if token == "" {
				c.JSON(401, gin.H{"error": "Authorization required"})
				c.Abort()
				return
			}
			authHeader = "Bearer " + token
		}

		claims, err := a.ValidateToken(authHeader)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Store claims in context
		c.Set("device_id", claims.DeviceID)
		c.Set("device_name", claims.DeviceName)
		c.Next()
	})
}
