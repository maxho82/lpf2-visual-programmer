package main

import (
	"fmt"
	"log"
	"time"
)

// SensorMonitor мониторит состояние устройств и батареи
type SensorMonitor struct {
	hubMgr    *HubManager
	deviceMgr *DeviceManager
	gui       *GUI
	stopCh    chan struct{}
}

// NewSensorMonitor создает новый монитор
func NewSensorMonitor(hubMgr *HubManager, deviceMgr *DeviceManager, gui *GUI) *SensorMonitor {
	return &SensorMonitor{
		hubMgr:    hubMgr,
		deviceMgr: deviceMgr,
		gui:       gui,
		stopCh:    make(chan struct{}),
	}
}

// Start запускает мониторинг
func (sm *SensorMonitor) Start() {
	go sm.monitorLoop()
	log.Println("Мониторинг устройств запущен")
}

// Stop останавливает мониторинг
func (sm *SensorMonitor) Stop() {
	if sm.stopCh != nil {
		close(sm.stopCh)
	}
	log.Println("Мониторинг устройств остановлен")
}

// monitorLoop основной цикл мониторинга
func (sm *SensorMonitor) monitorLoop() {
	ticker := time.NewTicker(500 * time.Millisecond) // Опрос каждые 500мс
	defer ticker.Stop()

	for {
		select {
		case <-sm.stopCh:
			return
		case <-ticker.C:
			if sm.hubMgr.IsConnected() {
				sm.updateBatteryLevel()
				sm.updatePortStatus()
				sm.updateSensorValues()
			}
		}
	}
}

// updateBatteryLevel обновляет уровень батареи
func (sm *SensorMonitor) updateBatteryLevel() {
	// Чтение уровня батареи (стандартная BLE характеристика)
	// UUID для Battery Level: 00002a19-0000-1000-8000-00805f9b34fb
	if sm.hubMgr.IsConnected() { // ИСПРАВЛЕНО
		// В реальном приложении здесь будет чтение BLE характеристики
		// Для теста используем фиктивное значение
		batteryLevel := 85 // Тестовое значение 85%

		sm.hubMgr.connectionMutex.Lock()
		if sm.hubMgr.hubInfo != nil {
			sm.hubMgr.hubInfo.Battery = batteryLevel
		}
		sm.hubMgr.connectionMutex.Unlock()
	}
}

// updatePortStatus обновляет статус портов
func (sm *SensorMonitor) updatePortStatus() {
	// Обновляем информацию о подключенных устройствах
	sm.hubMgr.connectionMutex.Lock()
	defer sm.hubMgr.connectionMutex.Unlock()

	if sm.hubMgr.hubInfo == nil {
		return
	}

	// Обновляем информацию о портах
	// Порты 1, 2, 6 (встроенный светодиод)
	ports := []byte{1, 2, 6}

	for _, port := range ports {
		// Проверяем, есть ли информация о порте
		found := false
		for i, portInfo := range sm.hubMgr.hubInfo.Ports {
			if portInfo.PortID == port {
				found = true
				// Обновляем статус (в реальном приложении - чтение из BLE)
				sm.hubMgr.hubInfo.Ports[i].LastUpdate = time.Now()
				break
			}
		}

		if !found {
			// Добавляем новый порт
			portInfo := PortInfo{
				PortID:      port,
				DeviceType:  sm.getDeviceTypeForPort(port),
				DeviceName:  sm.getDeviceNameForPort(port),
				IsConnected: true,
				LastValue:   []byte{0},
				Mode:        0,
				LastUpdate:  time.Now(), // ДОБАВЛЕНО
			}
			sm.hubMgr.hubInfo.Ports = append(sm.hubMgr.hubInfo.Ports, portInfo)
		}
	}
}

// updateSensorValues обновляет значения датчиков
func (sm *SensorMonitor) updateSensorValues() {
	// Симуляция чтения значений датчиков
	sm.hubMgr.connectionMutex.Lock()
	defer sm.hubMgr.connectionMutex.Unlock()

	if sm.hubMgr.hubInfo == nil {
		return
	}

	// Обновляем значения для каждого порта
	for i := range sm.hubMgr.hubInfo.Ports {
		port := &sm.hubMgr.hubInfo.Ports[i]

		// Генерируем тестовые значения в зависимости от типа устройства
		switch port.DeviceType {
		case 0x01: // Мотор
			// Значение мощности (0-100%)
			port.LastValue = []byte{byte(50 + i*10)} // Тестовое значение

		case 0x02: // Датчик наклона
			// Угол наклона (0-90 градусов)
			port.LastValue = []byte{byte(20 + i*5)} // Тестовое значение

		case 0x17: // RGB светодиод
			// Цвет (R,G,B)
			port.LastValue = []byte{255, 100, 50} // Тестовое значение

		case 0x23: // Датчик расстояния
			// Расстояние (0-10 см)
			port.LastValue = []byte{byte(5 + i)} // Тестовое значение
		}

		port.LastUpdate = time.Now()
	}
}

// getDeviceTypeForPort возвращает тип устройства для порта
func (sm *SensorMonitor) getDeviceTypeForPort(port byte) byte {
	switch port {
	case 1, 2:
		return 0x01 // Мотор (предполагаем по умолчанию)
	case 6:
		return 0x17 // RGB светодиод
	default:
		return 0x00 // Неизвестное
	}
}

// getDeviceNameForPort возвращает имя устройства для порта
func (sm *SensorMonitor) getDeviceNameForPort(port byte) string {
	switch port {
	case 1:
		return "Мотор A"
	case 2:
		return "Мотор B"
	case 6:
		return "Светодиод"
	default:
		return fmt.Sprintf("Порт %d", port)
	}
}

// GetSensorValue возвращает значение датчика в удобном формате
func (sm *SensorMonitor) GetSensorValue(portID byte) (string, string) {
	sm.hubMgr.connectionMutex.RLock()
	defer sm.hubMgr.connectionMutex.RUnlock()

	if sm.hubMgr.hubInfo == nil {
		return "Нет данных", ""
	}

	for _, port := range sm.hubMgr.hubInfo.Ports {
		if port.PortID == portID {
			return sm.formatSensorValue(port), port.DeviceName
		}
	}

	return "Не подключен", ""
}

// formatSensorValue форматирует значение датчика
func (sm *SensorMonitor) formatSensorValue(port PortInfo) string {
	if len(port.LastValue) == 0 {
		return "Нет данных"
	}

	switch port.DeviceType {
	case 0x01: // Мотор
		return fmt.Sprintf("Мощность: %d%%", port.LastValue[0])

	case 0x02: // Датчик наклона
		return fmt.Sprintf("Угол: %d°", port.LastValue[0])

	case 0x17: // RGB светодиод
		if len(port.LastValue) >= 3 {
			return fmt.Sprintf("RGB(%d,%d,%d)",
				port.LastValue[0], port.LastValue[1], port.LastValue[2])
		}
		return "Цвет: неизвестен"

	case 0x23: // Датчик расстояния
		return fmt.Sprintf("Расстояние: %d см", port.LastValue[0])

	default:
		return fmt.Sprintf("Значение: %v", port.LastValue)
	}
}
