package service

import (
	"log"
	"net"
	"sync"

	"github.com/grandcat/zeroconf"
)

// BonjourConfig содержит настройки для публикации Bonjour-сервиса
type BonjourConfig struct {
	ServiceName   string
	ServiceType   string
	Domain        string
	Port          int
	TxtRecords    []string
	InterfaceOpts []net.Interface
}

// DefaultBonjourConfig возвращает конфигурацию по умолчанию
func DefaultBonjourConfig() BonjourConfig {
	return BonjourConfig{
		ServiceName:   "MediaFS",
		ServiceType:   "_http._tcp",
		Domain:        "local.",
		Port:          8000,
		TxtRecords:    []string{"mediafs", "1.0"},
		InterfaceOpts: nil,
	}
}

// BonjourService представляет собой сервис Bonjour
type BonjourService struct {
	config   BonjourConfig
	server   *zeroconf.Server
	stopCh   chan struct{}
	mu       sync.Mutex
	isActive bool
}

// NewBonjourService создаёт новый экземпляр сервиса Bonjour
func NewBonjourService(config BonjourConfig) *BonjourService {
	return &BonjourService{
		config: config,
		stopCh: make(chan struct{}),
	}
}

// Start запускает публикацию Bonjour-сервиса
func (bs *BonjourService) Start() error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if bs.isActive {
		return nil
	}

	var err error
	bs.server, err = zeroconf.Register(
		bs.config.ServiceName,
		bs.config.ServiceType,
		bs.config.Domain,
		bs.config.Port,
		bs.config.TxtRecords,
		bs.config.InterfaceOpts,
	)

	if err != nil {
		log.Printf("❌ Failed to publish Bonjour service: %v", err)
		return err
	}

	log.Printf("✅ Bonjour service '%s.%s.%s' published on port %d",
		bs.config.ServiceName, bs.config.ServiceType, bs.config.Domain, bs.config.Port)

	bs.isActive = true
	return nil
}

// Stop останавливает публикацию Bonjour-сервиса
func (bs *BonjourService) Stop() {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if !bs.isActive || bs.server == nil {
		return
	}

	log.Println("📢 Shutting down Bonjour service...")
	bs.server.Shutdown()
	close(bs.stopCh)
	bs.isActive = false
	bs.server = nil
}

// IsActive проверяет, активен ли сервис
func (bs *BonjourService) IsActive() bool {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	return bs.isActive
}

// RunInBackground запускает сервис Bonjour в фоновом режиме
func (bs *BonjourService) RunInBackground() {
	if err := bs.Start(); err != nil {
		log.Printf("❌ Could not start Bonjour service: %v", err)
		return
	}

	<-bs.stopCh
}
