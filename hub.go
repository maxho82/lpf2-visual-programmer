package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	tinybluetooth "tinygo.org/x/bluetooth"
)

// HubManager управляет подключением к LPF2-хабу
type HubManager struct {
	adapter         *tinybluetooth.Adapter
	device          tinybluetooth.Device
	deviceAddress   string
	isConnected     bool
	connectionMutex sync.RWMutex
	hubInfo         *HubInfo
	stopScan        context.CancelFunc
}

// HubInfo содержит информацию о подключенном хабе
type HubInfo struct {
	Name        string
	Address     string
	Battery     int
	Firmware    string
	Ports       []PortInfo
	LastUpdated time.Time
}

// PortInfo информация о порте хаба
type PortInfo struct {
	PortID      byte
	DeviceType  byte
	DeviceName  string
	IsConnected bool
	LastValue   []byte
	Mode        byte
}

// NewHubManager создает новый менеджер хаба
func NewHubManager() (*HubManager, error) {
	adapter := tinybluetooth.DefaultAdapter
	if adapter == nil {
		return nil, fmt.Errorf("BLE адаптер не найден")
	}

	// Включение адаптера
	if err := adapter.Enable(); err != nil {
		return nil, fmt.Errorf("ошибка включения BLE адаптера: %v", err)
	}

	return &HubManager{
		adapter: adapter,
		hubInfo: &HubInfo{
			Ports: make([]PortInfo, 6),
		},
	}, nil
}

// ScanForHubs сканирует LPF2-хабы
func (hm *HubManager) ScanForHubs(timeout time.Duration) ([]HubInfo, error) {
	var foundHubs []HubInfo
	var scanMutex sync.Mutex

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	hm.stopScan = cancel

	log.Println("=== Начало сканирования LPF2 хабов ===")

	err := hm.adapter.Scan(func(adapter *tinybluetooth.Adapter, result tinybluetooth.ScanResult) {
		select {
		case <-ctx.Done():
			return
		default:
		}

		name := result.LocalName()
		address := result.Address.String()

		// Логируем все устройства для отладки
		if name != "" {
			log.Printf("Устройство: %s [%s]", name, address)
		}

		// Ищем LPF2 хаб по имени
		if name == "LPF2 Smart Hub 2 I/O" {
			log.Printf("!!! НАЙДЕН LPF2 ХАБ: %s [%s]", name, address)

			scanMutex.Lock()
			foundHubs = append(foundHubs, HubInfo{
				Name:    name,
				Address: address,
			})
			scanMutex.Unlock()

			// Останавливаем сканирование при нахождении
			adapter.StopScan()
			cancel()
			return
		}

		// Альтернативные имена хаба
		if name == "LEGO Hub" || name == "Wedo" || name == "LEGO HUB" ||
			name == "LEGO Boost" || name == "Move Hub" {
			log.Printf("Найден LEGO хаб: %s [%s]", name, address)

			scanMutex.Lock()
			foundHubs = append(foundHubs, HubInfo{
				Name:    name,
				Address: address,
			})
			scanMutex.Unlock()
		}
	})

	if err != nil {
		return nil, fmt.Errorf("ошибка сканирования: %v", err)
	}

	// Ждем завершения таймаута или отмены
	<-ctx.Done()
	hm.adapter.StopScan()

	log.Printf("Сканирование завершено. Найдено хабов: %d", len(foundHubs))
	return foundHubs, nil
}

// Connect подключается к выбранному хабу
func (hm *HubManager) Connect(address string) error {
	hm.connectionMutex.Lock()
	defer hm.connectionMutex.Unlock()

	// Если уже подключены, отключаемся
	if hm.isConnected {
		hm.Disconnect()
	}

	log.Printf("Попытка подключения к %s", address)

	// Находим устройство через сканирование
	var targetDevice tinybluetooth.ScanResult
	found := false

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Сканируем для получения устройства...")

	err := hm.adapter.Scan(func(adapter *tinybluetooth.Adapter, result tinybluetooth.ScanResult) {
		if result.Address.String() == address {
			log.Printf("Найдено устройство для подключения: %s", result.LocalName())
			adapter.StopScan()
			targetDevice = result
			found = true
			cancel()
		}
	})

	if err != nil {
		return fmt.Errorf("ошибка сканирования: %v", err)
	}

	<-ctx.Done()
	hm.adapter.StopScan()

	if !found {
		return fmt.Errorf("устройство с адресом %s не найдено", address)
	}

	// Подключаемся (как в тестовом коде)
	log.Printf("Устанавливаем соединение с %s...", address)
	device, err := hm.adapter.Connect(targetDevice.Address, tinybluetooth.ConnectionParams{})
	if err != nil {
		return fmt.Errorf("ошибка подключения: %v", err)
	}

	hm.device = device
	hm.deviceAddress = address
	hm.isConnected = true

	// Обновляем информацию о хабе
	hm.hubInfo.Name = targetDevice.LocalName()
	hm.hubInfo.Address = address
	hm.hubInfo.LastUpdated = time.Now()

	log.Printf("Успешно подключено к %s (%s)", address, hm.hubInfo.Name)
	return nil
}

// Disconnect отключается от хаба
func (hm *HubManager) Disconnect() {
	hm.connectionMutex.Lock()
	defer hm.connectionMutex.Unlock()

	if hm.isConnected {
		log.Println("Отключение от хаба...")
		hm.device.Disconnect()
		hm.isConnected = false
		hm.hubInfo = &HubInfo{
			Ports: make([]PortInfo, 6),
		}
		log.Println("Отключено")
	}
}

// IsConnected возвращает статус подключения
func (hm *HubManager) IsConnected() bool {
	hm.connectionMutex.RLock()
	defer hm.connectionMutex.RUnlock()
	return hm.isConnected
}

// GetHubInfo возвращает информацию о хабе
func (hm *HubManager) GetHubInfo() HubInfo {
	hm.connectionMutex.RLock()
	defer hm.connectionMutex.RUnlock()

	if hm.hubInfo == nil {
		return HubInfo{}
	}
	return *hm.hubInfo
}

// WriteCharacteristic записывает данные в характеристику
func (hm *HubManager) WriteCharacteristic(uuid string, data []byte) error {
	hm.connectionMutex.RLock()
	defer hm.connectionMutex.RUnlock()

	if !hm.IsConnected() {
		return fmt.Errorf("не подключено к хабу")
	}

	log.Printf("Запись в характеристику %s: %v", uuid, data)

	// В упрощенной версии просто логируем
	// В реальной реализации нужно найти характеристику и отправить данные
	return nil
}
