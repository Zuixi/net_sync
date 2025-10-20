package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/easy-sync/easy-sync/pkg/config"
	"github.com/easy-sync/easy-sync/pkg/security"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type MessageType string

const (
	MessageTypeHello        MessageType = "hello"
	MessageTypeChat         MessageType = "chat"
	MessageTypeFileOffer    MessageType = "file_offer"
	MessageTypeFileOfferAck MessageType = "file_offer_ack"
	MessageTypeDeliveryAck  MessageType = "delivery_ack"
	MessageTypeTyping       MessageType = "typing"
	MessageTypePresence     MessageType = "presence"
)

type Message struct {
	Type      MessageType `json:"type"`
	ID        string      `json:"id,omitempty"`
	Timestamp int64       `json:"timestamp"`
	Device    string      `json:"device,omitempty"`
	Text      string      `json:"text,omitempty"`
	From      string      `json:"from,omitempty"`
	OfferID   string      `json:"offer_id,omitempty"`
	Accepted  bool        `json:"accepted,omitempty"`
}

type FileOfferMessage struct {
	Type    MessageType `json:"type"`
	OfferID string      `json:"offer_id"`
	From    string      `json:"from"`
	Name    string      `json:"name"`
	Size    int64       `json:"size"`
	Mime    string      `json:"mime"`
	SHA256  string      `json:"sha256"`
	URL     string      `json:"url"`
}

type HelloMessage struct {
	Type         string   `json:"type"`
	Device       string   `json:"device"`
	Capabilities []string `json:"capabilities"`
}

type Client struct {
	ID          string
	DeviceName  string
	Connection  *websocket.Conn
	Send        chan Message
	Manager     *Manager
	LastPing    time.Time
	IsConnected bool
	mu          sync.RWMutex
}

type Manager struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan Message
	logger     *logrus.Logger
	config     *config.Config
	auth       *security.AuthService
	upgrader   websocket.Upgrader
	mu         sync.RWMutex
}

func NewManager(cfg *config.Config, logger *logrus.Logger, auth *security.AuthService) *Manager {
	return &Manager{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Message),
		logger:     logger,
		config:     cfg,
		auth:       auth,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  cfg.WebSocket.ReadBufferSize,
			WriteBufferSize: cfg.WebSocket.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for LAN usage
			},
		},
	}
}

func (m *Manager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract token from query parameter or header
	token := r.URL.Query().Get("token")
	if token == "" {
		token = r.Header.Get("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}

	// Validate token
	claims, err := m.auth.ValidateToken(token)
	if err != nil {
		m.logger.WithError(err).Warn("WebSocket connection with invalid token")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		m.logger.WithError(err).Error("Failed to upgrade connection to WebSocket")
		return
	}

	// Create new client
	client := &Client{
		ID:          uuid.New().String(),
		DeviceName:  claims.DeviceName,
		Connection:  conn,
		Send:        make(chan Message, 256),
		Manager:     m,
		LastPing:    time.Now(),
		IsConnected: true,
	}

	// Register client
	m.register <- client

	// Start goroutines
	go client.writePump()
	go client.readPump()
}

func (m *Manager) Start() {
	go m.run()
}

func (m *Manager) run() {
	ticker := time.NewTicker(time.Duration(m.config.WebSocket.PingPeriod) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-m.register:
			m.registerClient(client)

		case client := <-m.unregister:
			m.unregisterClient(client)

		case message := <-m.broadcast:
			m.broadcastMessage(message)

		case <-ticker.C:
			m.checkConnections()
		}
	}
}

func (m *Manager) registerClient(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clients[client.ID] = client

	m.logger.WithFields(logrus.Fields{
		"client_id":   client.ID,
		"device_name": client.DeviceName,
	}).Info("Client connected")

	// Send welcome message
	welcome := Message{
		Type:      MessageTypeHello,
		Device:    m.config.Discovery.DeviceName,
		Timestamp: time.Now().Unix(),
	}

	select {
	case client.Send <- welcome:
	default:
		close(client.Send)
		delete(m.clients, client.ID)
	}
}

func (m *Manager) unregisterClient(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.clients[client.ID]; ok {
		delete(m.clients, client.ID)
		close(client.Send)
		client.Connection.Close()

		m.logger.WithFields(logrus.Fields{
			"client_id":   client.ID,
			"device_name": client.DeviceName,
		}).Info("Client disconnected")
	}
}

func (m *Manager) broadcastMessage(message Message) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, client := range m.clients {
		if client.IsConnected {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(m.clients, client.ID)
			}
		}
	}
}

func (m *Manager) checkConnections() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, client := range m.clients {
		client.mu.RLock()
		isConnected := client.IsConnected
		lastPing := client.LastPing
		client.mu.RUnlock()

		if !isConnected || time.Since(lastPing) > time.Duration(m.config.WebSocket.PingPeriod*2)*time.Second {
			m.unregister <- client
		} else {
			// Send ping
			ping := Message{
				Type:      "ping",
				Timestamp: time.Now().Unix(),
			}

			select {
			case client.Send <- ping:
			default:
				m.unregister <- client
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.Manager.unregister <- c
		c.Connection.Close()
	}()

	c.Connection.SetReadLimit(512)
	c.Connection.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Connection.SetPongHandler(func(string) error {
		c.mu.Lock()
		c.LastPing = time.Now()
		c.mu.Unlock()
		c.Connection.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg json.RawMessage
		err := c.Connection.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Manager.logger.WithError(err).Error("WebSocket error")
			}
			break
		}

		c.handleMessage(msg)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Connection.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Connection.WriteJSON(message); err != nil {
				c.Manager.logger.WithError(err).Error("Failed to write message")
				return
			}

		case <-ticker.C:
			c.Connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(msg json.RawMessage) {
	var baseMsg struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(msg, &baseMsg); err != nil {
		c.Manager.logger.WithError(err).Error("Failed to parse message type")
		return
	}

	switch MessageType(baseMsg.Type) {
	case MessageTypeHello:
		c.handleHelloMessage(msg)

	case MessageTypeChat:
		c.handleChatMessage(msg)

	case MessageTypeFileOfferAck:
		c.handleFileOfferAckMessage(msg)

	default:
		c.Manager.logger.WithField("type", baseMsg.Type).Warn("Unknown message type")
	}
}

func (c *Client) handleHelloMessage(msg json.RawMessage) {
	var helloMsg HelloMessage
	if err := json.Unmarshal(msg, &helloMsg); err != nil {
		c.Manager.logger.WithError(err).Error("Failed to parse hello message")
		return
	}

	c.mu.Lock()
	c.DeviceName = helloMsg.Device
	c.mu.Unlock()

	c.Manager.logger.WithFields(logrus.Fields{
		"client_id":   c.ID,
		"device_name": helloMsg.Device,
	}).Info("Client identified")

	// Broadcast to other clients
	response := Message{
		Type:      MessageTypePresence,
		Device:    helloMsg.Device,
		Timestamp: time.Now().Unix(),
	}

	c.Manager.broadcast <- response
}

func (c *Client) handleChatMessage(msg json.RawMessage) {
	var chatMsg Message
	if err := json.Unmarshal(msg, &chatMsg); err != nil {
		c.Manager.logger.WithError(err).Error("Failed to parse chat message")
		return
	}

	chatMsg.From = c.DeviceName
	chatMsg.Timestamp = time.Now().Unix()

	c.Manager.broadcast <- chatMsg
}

func (c *Client) handleFileOfferAckMessage(msg json.RawMessage) {
	var ackMsg Message
	if err := json.Unmarshal(msg, &ackMsg); err != nil {
		c.Manager.logger.WithError(err).Error("Failed to parse file offer ack message")
		return
	}

	ackMsg.From = c.DeviceName
	ackMsg.Timestamp = time.Now().Unix()

	c.Manager.broadcast <- ackMsg
}

func (m *Manager) SendFileOffer(offer FileOfferMessage) error {
	// Convert FileOfferMessage to JSON and embed in Message
	offerData, err := json.Marshal(offer)
	if err != nil {
		return err
	}

	// Use a custom type to embed the file offer data
	type FileOfferEmbed struct {
		Type      MessageType `json:"type"`
		OfferID   string      `json:"offer_id"`
		From      string      `json:"from"`
		Name      string      `json:"name"`
		Size      int64       `json:"size"`
		Mime      string      `json:"mime"`
		SHA256    string      `json:"sha256"`
		URL       string      `json:"url"`
		Timestamp int64       `json:"timestamp"`
	}

	var fileOffer FileOfferEmbed
	if err := json.Unmarshal(offerData, &fileOffer); err != nil {
		return err
	}

	// Send as raw JSON to preserve all fields
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, client := range m.clients {
		if client.IsConnected {
			if err := client.Connection.WriteJSON(fileOffer); err != nil {
				m.logger.WithError(err).Error("Failed to send file offer")
			}
		}
	}

	return nil
}

func (m *Manager) GetConnectedDevices() []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	devices := make([]map[string]interface{}, 0, len(m.clients))
	for _, client := range m.clients {
		client.mu.RLock()
		if client.IsConnected {
			devices = append(devices, map[string]interface{}{
				"id":          client.ID,
				"device_name": client.DeviceName,
				"last_ping":   client.LastPing,
			})
		}
		client.mu.RUnlock()
	}

	return devices
}
