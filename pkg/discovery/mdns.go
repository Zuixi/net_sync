package discovery

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/sirupsen/logrus"
	"github.com/easy-sync/easy-sync/pkg/config"
)

type MDnsDiscovery struct {
	config    *config.Config
	logger    *logrus.Logger
	server    *zeroconf.Server
	resolver  *zeroconf.Resolver
	ctx       context.Context
	cancel    context.CancelFunc
	isRunning bool
}

type ServiceInfo struct {
	ServiceName string
	DeviceName  string
	Host        string
	Port        int
	Text        map[string]string
	LastSeen    time.Time
}

func NewMDnsDiscovery(cfg *config.Config, logger *logrus.Logger) *MDnsDiscovery {
	ctx, cancel := context.WithCancel(context.Background())

	return &MDnsDiscovery{
		config:   cfg,
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
		resolver: zeroconf.NewResolver(nil),
	}
}

func (m *MDnsDiscovery) Start() error {
	if !m.config.Discovery.Enabled {
		m.logger.Info("mDNS discovery is disabled")
		return nil
	}

	if m.isRunning {
		return fmt.Errorf("mDNS discovery is already running")
	}

	// Create TXT record with service information
	txt := []string{
		fmt.Sprintf("device_name=%s", m.config.Discovery.DeviceName),
		fmt.Sprintf("port=%d", m.config.Server.Port),
		"supports=http",
	}

	if m.config.Server.HTTPS {
		txt = append(txt, "supports=https")
	}

	// Register the service
	server, err := zeroconf.Register(
		m.config.Discovery.DeviceName,
		m.config.Discovery.ServiceName,
		"local.",
		m.config.Server.Port,
		txt,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to register mDNS service: %w", err)
	}

	m.server = server
	m.isRunning = true

	m.logger.WithFields(logrus.Fields{
		"device_name":  m.config.Discovery.DeviceName,
		"service_name": m.config.Discovery.ServiceName,
		"port":         m.config.Server.Port,
	}).Info("mDNS service registered")

	// Start browsing for other services
	go m.browseServices()

	return nil
}

func (m *MDnsDiscovery) Stop() error {
	if !m.isRunning {
		return nil
	}

	m.cancel()

	if m.server != nil {
		m.server.Shutdown()
	}

	m.isRunning = false
	m.logger.Info("mDNS discovery stopped")
	return nil
}

func (m *MDnsDiscovery) browseServices() {
	entries := make(chan *zeroconf.ServiceEntry)
	go func() {
		for {
			select {
			case entry := <-entries:
				m.handleServiceEntry(entry)
			case <-m.ctx.Done():
				return
			}
		}
	}()

	err := m.resolver.Browse(m.ctx, m.config.Discovery.ServiceName, "")
	if err != nil {
		m.logger.WithError(err).Error("Failed to browse mDNS services")
		return
	}

	<-m.ctx.Done()
}

func (m *MDnsDiscovery) handleServiceEntry(entry *zeroconf.ServiceEntry) {
	// Skip our own service
	if entry.Instance == m.config.Discovery.DeviceName {
		return
	}

	serviceInfo := m.parseServiceEntry(entry)

	m.logger.WithFields(logrus.Fields{
		"device_name": serviceInfo.DeviceName,
		"host":        serviceInfo.Host,
		"port":        serviceInfo.Port,
	}).Info("Discovered mDNS service")

	// TODO: Store discovered services in a database or memory cache
	// TODO: Notify connected clients about discovered services
}

func (m *MDnsDiscovery) parseServiceEntry(entry *zeroconf.ServiceEntry) *ServiceInfo {
	serviceInfo := &ServiceInfo{
		ServiceName: entry.Service,
		DeviceName:  entry.Instance,
		Host:        "",
		Port:        entry.Port,
		Text:        make(map[string]string),
		LastSeen:    time.Now(),
	}

	// Get host from IPv4 addresses
	if len(entry.AddrIPv4) > 0 {
		serviceInfo.Host = entry.AddrIPv4[0].String()
	}

	// Parse TXT records
	for _, txt := range entry.Text {
		parts := strings.SplitN(txt, "=", 2)
		if len(parts) == 2 {
			serviceInfo.Text[parts[0]] = parts[1]
		}
	}

	return serviceInfo
}

func (m *MDnsDiscovery) DiscoverServices(timeout time.Duration) ([]*ServiceInfo, error) {
	if !m.config.Discovery.Enabled {
		return nil, fmt.Errorf("mDNS discovery is disabled")
	}

	entries := make(chan *zeroconf.ServiceEntry)
	services := make([]*ServiceInfo, 0)
	done := make(chan bool)

	go func() {
		for entry := range entries {
			// Skip our own service
			if entry.Instance == m.config.Discovery.DeviceName {
				continue
			}

			serviceInfo := m.parseServiceEntry(entry)
			services = append(services, serviceInfo)
		}
		done <- true
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := m.resolver.Browse(ctx, m.config.Discovery.ServiceName, "")
	if err != nil {
		return nil, fmt.Errorf("failed to browse mDNS services: %w", err)
	}

	<-ctx.Done()
	close(entries)
	<-done

	m.logger.WithField("services_found", len(services)).Info("mDNS service discovery completed")
	return services, nil
}

func (m *MDnsDiscovery) GetServiceURL(service *ServiceInfo) string {
	protocol := "http"
	if service.Text["supports"] == "https" {
		protocol = "https"
	}

	if service.Host == "" {
		return ""
	}

	return fmt.Sprintf("%s://%s:%d", protocol, service.Host, service.Port)
}

func (m *MDnsDiscovery) GetLocalServiceURL() string {
	protocol := "http"
	if m.config.Server.HTTPS {
		protocol = "https"
	}

	// Try to get local IP address
	// For now, use localhost
	return fmt.Sprintf("%s://localhost:%d", protocol, m.config.Server.Port)
}

func (m *MDnsDiscovery) IsRunning() bool {
	return m.isRunning
}

func (m *MDnsDiscovery) GetServiceName() string {
	return m.config.Discovery.ServiceName
}

func (m *MDnsDiscovery) GetDeviceName() string {
	return m.config.Discovery.DeviceName
}